#include <signal.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TBufferTransports.h>
#include <thrift/transport/TServerSocket.h>

#include "../utils.h"
#include "../utils_mongodb.h"
#include "../utils_redis.h"
#include "../utils_thrift.h"
#include "SocialGraphHandler.h"
#include "../InfluxClient.h"

using json = nlohmann::json;
using apache::thrift::protocol::TBinaryProtocolFactory;
using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::transport::TServerSocket;
using namespace social_network;

NEW_INFLUX_CLIENT
void sigintHandler(int sig)
{
  CLOSE_INFLUX_CLIENT
  exit(EXIT_SUCCESS);
}

int main(int argc, char *argv[])
{
  signal(SIGINT, sigintHandler);
  init_logger();

  SetUpTracer("config/jaeger-config.yml", "social-graph-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0)
  {
    exit(EXIT_FAILURE);
  }

  int port = config_json["social-graph-service"]["port"];

  int mongodb_conns = config_json["social-graph-mongodb"]["connections"];
  int mongodb_timeout = config_json["social-graph-mongodb"]["timeout_ms"];

  std::string user_addr = config_json["user-service"]["addr"];
  int user_port = config_json["user-service"]["port"];
  int user_conns = config_json["user-service"]["connections"];
  int user_timeout = config_json["user-service"]["timeout_ms"];
  int user_keepalive = config_json["user-service"]["keepalive_ms"];

  mongoc_client_pool_t *mongodb_client_pool =
      init_mongodb_client_pool(config_json, "social-graph", mongodb_conns);

  if (mongodb_client_pool == nullptr)
  {
    return EXIT_FAILURE;
  }

  ClientPool<ThriftClient<UserServiceClient>> user_client_pool(
      "social-graph", user_addr, user_port, 0, user_conns, user_timeout,
      user_keepalive, config_json);

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(mongodb_client_pool);
  if (!mongodb_client)
  {
    LOG(fatal) << "Failed to pop mongoc client";
    return EXIT_FAILURE;
  }
  bool r = false;
  while (!r)
  {
    r = CreateIndex(mongodb_client, "social-graph", "user_id", true);
    if (!r)
    {
      LOG(error) << "Failed to create mongodb index, try again";
      sleep(1);
    }
  }
  mongoc_client_pool_push(mongodb_client_pool, mongodb_client);

  Redis redis_client_pool = init_redis_client_pool(config_json, "social-graph");
  std::shared_ptr<TServerSocket> server_socket = get_server_socket(config_json, "0.0.0.0", port);

  TThreadedServer server(
      std::make_shared<SocialGraphServiceProcessor>(
          std::make_shared<SocialGraphHandler>(
              mongodb_client_pool, &redis_client_pool, &user_client_pool, INFLUX_CLIENT_VAR)),
      server_socket,
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>());

  LOG(info) << "Starting the social-graph-service server ...";
  server.serve();
}
