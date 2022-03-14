#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_

#include <bson/bson.h>
#include <mongoc.h>
#include <sw/redis++/redis++.h>

#include <future>
#include <iostream>
#include <string>

#include "../../gen-cpp/PostStorageService.h"
#include "../../gen-cpp/UserTimelineService.h"
#include "../ClientPool.h"
#include "../ThriftClient.h"
#include "../logger.h"
#include "../tracing.h"
#include "../InfluxClient.h"

using namespace sw::redis;

namespace social_network
{

  class UserTimelineHandler : public UserTimelineServiceIf
  {
  public:
    UserTimelineHandler(Redis *, mongoc_client_pool_t *,
                        ClientPool<ThriftClient<PostStorageServiceClient>> *,
                        INFLUX_CLIENT_PTR);
    ~UserTimelineHandler() override = default;

    void WriteUserTimeline(
        int64_t req_id, int64_t post_id, int64_t user_id, int64_t timestamp,
        const std::map<std::string, std::string> &carrier) override;

    void ReadUserTimeline(std::vector<Post> &, int64_t, int64_t, int, int,
                          const std::map<std::string, std::string> &) override;

  private:
    Redis *_redis_client_pool;
    mongoc_client_pool_t *_mongodb_client_pool;
    ClientPool<ThriftClient<PostStorageServiceClient>> *_post_client_pool;
    ANNOUNCE_INFLUX_CLIENT
  };

  UserTimelineHandler::UserTimelineHandler(
      Redis *redis_pool, mongoc_client_pool_t *mongodb_pool,
      ClientPool<ThriftClient<PostStorageServiceClient>> *post_client_pool,
      INFLUX_CLIENT_PLACEHOLDER) : INJECT_INFLUX_CLIENT_DEFAULT
  {
    _redis_client_pool = redis_pool;
    _mongodb_client_pool = mongodb_pool;
    _post_client_pool = post_client_pool;
  }

  void UserTimelineHandler::WriteUserTimeline(
      int64_t req_id, int64_t post_id, int64_t user_id, int64_t timestamp,
      const std::map<std::string, std::string> &carrier)
  {
    START_SPAN(write_user_timeline_server)

    mongoc_client_t *mongodb_client =
        mongoc_client_pool_pop(_mongodb_client_pool);
    if (!mongodb_client)
    {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to pop a client from MongoDB pool";
      throw se;
    }
    auto collection = mongoc_client_get_collection(
        mongodb_client, "user-timeline", "user-timeline");
    if (!collection)
    {
      ServiceException se;
      se.errorCode = ErrorCode::SE_MONGODB_ERROR;
      se.message = "Failed to create collection user-timeline from MongoDB";
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
      throw se;
    }
    bson_t *query = bson_new();

    BSON_APPEND_INT64(query, "user_id", user_id);
    bson_t *update =
        BCON_NEW("$push", "{", "posts", "{", "$each", "[", "{", "post_id",
                 BCON_INT64(post_id), "timestamp", BCON_INT64(timestamp), "}",
                 "]", "$position", BCON_INT32(0), "}", "}");
    bson_error_t error;
    bson_t reply;
    START_SPAN_WITH_CARRIER(write_user_timeline_mongo_insert_client, next_carrier_write_user_timeline_server)
    bool updated = mongoc_collection_find_and_modify(collection, query, nullptr,
                                                     update, nullptr, false, true,
                                                     true, &reply, &error);
    span_write_user_timeline_mongo_insert_client->Finish();

    if (!updated)
    {
      // update the newly inserted document (upsert: false)
      updated = mongoc_collection_find_and_modify(collection, query, nullptr,
                                                  update, nullptr, false, false,
                                                  true, &reply, &error);
      if (!updated)
      {
        LOG(error) << "Failed to update user-timeline for user " << user_id
                   << " to MongoDB: " << error.message;
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = error.message;
        bson_destroy(update);
        bson_destroy(query);
        bson_destroy(&reply);
        mongoc_collection_destroy(collection);
        mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
        throw se;
      }
    }

    bson_destroy(update);
    bson_destroy(&reply);
    bson_destroy(query);
    mongoc_collection_destroy(collection);
    mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);

    // Update user's timeline in redis
    START_SPAN_WITH_CARRIER(write_user_timeline_redis_update_client, next_carrier_write_user_timeline_server)
    try
    {
      _redis_client_pool->zadd(std::to_string(user_id), std::to_string(post_id),
                               timestamp, UpdateType::NOT_EXIST);
    }
    catch (const Error &err)
    {
      LOG(error) << err.what();
      throw err;
    }
    span_write_user_timeline_redis_update_client->Finish();
    span_write_user_timeline_server->Finish();
  }

  void UserTimelineHandler::ReadUserTimeline(
      std::vector<Post> &_return, int64_t req_id, int64_t user_id, int start,
      int stop, const std::map<std::string, std::string> &carrier)
  {
    START_SPAN(read_user_timeline_server)

    if (stop <= start || start < 0)
    {
      return;
    }

    START_SPAN_WITH_CARRIER(read_user_timeline_redis_find_client, next_carrier_read_user_timeline_server)

    std::vector<std::string> post_ids_str;
    try
    {
      _redis_client_pool->zrevrange(std::to_string(user_id), start, stop - 1,
                                    std::back_inserter(post_ids_str));
    }
    catch (const Error &err)
    {
      LOG(error) << err.what();
      throw err;
    }
    span_read_user_timeline_redis_find_client->Finish();

    std::vector<int64_t> post_ids;
    for (auto &post_id_str : post_ids_str)
    {
      post_ids.emplace_back(std::stoul(post_id_str));
    }

    // find in mongodb
    int mongo_start = start + post_ids.size();
    std::unordered_map<std::string, double> redis_update_map;
    if (mongo_start < stop)
    {
      // Instead find post_ids from mongodb
      mongoc_client_t *mongodb_client =
          mongoc_client_pool_pop(_mongodb_client_pool);
      if (!mongodb_client)
      {
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = "Failed to pop a client from MongoDB pool";
        throw se;
      }
      auto collection = mongoc_client_get_collection(
          mongodb_client, "user-timeline", "user-timeline");
      if (!collection)
      {
        ServiceException se;
        se.errorCode = ErrorCode::SE_MONGODB_ERROR;
        se.message = "Failed to create collection user-timeline from MongoDB";
        throw se;
      }

      bson_t *query = BCON_NEW("user_id", BCON_INT64(user_id));
      bson_t *opts = BCON_NEW("projection", "{", "posts", "{", "$slice", "[",
                              BCON_INT32(0), BCON_INT32(stop), "]", "}", "}");

      START_SPAN_WITH_CARRIER(user_timeline_mongo_find_client, next_carrier_read_user_timeline_server)
      mongoc_cursor_t *cursor =
          mongoc_collection_find_with_opts(collection, query, opts, nullptr);
      span_user_timeline_mongo_find_client->Finish();
      const bson_t *doc;
      bool found = mongoc_cursor_next(cursor, &doc);
      if (found)
      {
        bson_iter_t iter_0;
        bson_iter_t iter_1;
        bson_iter_t post_id_child;
        bson_iter_t timestamp_child;
        int idx = 0;
        bson_iter_init(&iter_0, doc);
        bson_iter_init(&iter_1, doc);
        while (bson_iter_find_descendant(
                   &iter_0, ("posts." + std::to_string(idx) + ".post_id").c_str(),
                   &post_id_child) &&
               BSON_ITER_HOLDS_INT64(&post_id_child) &&
               bson_iter_find_descendant(
                   &iter_1,
                   ("posts." + std::to_string(idx) + ".timestamp").c_str(),
                   &timestamp_child) &&
               BSON_ITER_HOLDS_INT64(&timestamp_child))
        {
          auto curr_post_id = bson_iter_int64(&post_id_child);
          auto curr_timestamp = bson_iter_int64(&timestamp_child);
          if (idx >= mongo_start)
          {
            post_ids.emplace_back(curr_post_id);
          }
          redis_update_map.insert(std::make_pair(std::to_string(curr_post_id),
                                                 (double)curr_timestamp));
          bson_iter_init(&iter_0, doc);
          bson_iter_init(&iter_1, doc);
          idx++;
        }
      }
      bson_destroy(opts);
      bson_destroy(query);
      mongoc_cursor_destroy(cursor);
      mongoc_collection_destroy(collection);
      mongoc_client_pool_push(_mongodb_client_pool, mongodb_client);
    }

    std::future<std::vector<Post>> post_future =
        std::async(std::launch::async, [&]()
                   {
        auto post_client_wrapper = _post_client_pool->Pop();
        if (!post_client_wrapper) {
          ServiceException se;
          se.errorCode = ErrorCode::SE_THRIFT_CONN_ERROR;
          se.message = "Failed to connect to post-storage-service";
          throw se;
        }
        std::vector<Post> _return_posts;
        auto post_client = post_client_wrapper->GetClient();
        try {
          START_SPAN_WITH_CARRIER(user_timeline_read_timeline_read_posts, next_carrier_read_user_timeline_server)
          post_client->ReadPosts(_return_posts, req_id, post_ids,
                                 next_carrier_read_user_timeline_server);
          span_user_timeline_read_timeline_read_posts->Finish();
        } catch (...) {
          _post_client_pool->Remove(post_client_wrapper);
          LOG(error) << "Failed to read posts from post-storage-service";
          throw;
        }
        _post_client_pool->Keepalive(post_client_wrapper);
        return _return_posts; });

    if (redis_update_map.size() > 0)
    {
      START_SPAN_WITH_CARRIER(user_timeline_redis_update_client, next_carrier_read_user_timeline_server)
      try
      {
        _redis_client_pool->zadd(std::to_string(user_id),
                                 redis_update_map.begin(),
                                 redis_update_map.end());
      }
      catch (const Error &err)
      {
        LOG(error) << err.what();
        throw err;
      }
      span_user_timeline_redis_update_client->Finish();
    }

    try
    {
      _return = post_future.get();
    }
    catch (...)
    {
      LOG(error) << "Failed to get post from post-storage-service";
      throw;
    }
    span_read_user_timeline_redis_find_client->Finish();
  }

} // namespace social_network

#endif // SOCIAL_NETWORK_MICROSERVICES_SRC_USERTIMELINESERVICE_USERTIMELINEHANDLER_H_
