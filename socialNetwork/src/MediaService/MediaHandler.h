#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_

#include <chrono>
#include <iostream>
#include <string>

#include "../../gen-cpp/MediaService.h"
#include "../logger.h"
#include "../tracing.h"
#include "../InfluxClient.h"

// 2018-01-01 00:00:00 UTC
#define CUSTOM_EPOCH 1514764800000

namespace social_network
{

  class MediaHandler : public MediaServiceIf
  {
  public:
    MediaHandler(INFLUX_CLIENT_PLACEHOLDER) : INJECT_INFLUX_CLIENT_DEFAULT {}
    ~MediaHandler() override = default;

    void ComposeMedia(std::vector<Media> &_return, int64_t,
                      const std::vector<std::string> &,
                      const std::vector<int64_t> &,
                      const std::map<std::string, std::string> &) override;

  private:
    ANNOUNCE_INFLUX_CLIENT
  };

  void MediaHandler::ComposeMedia(
      std::vector<Media> &_return, int64_t req_id,
      const std::vector<std::string> &media_types,
      const std::vector<int64_t> &media_ids,
      const std::map<std::string, std::string> &carrier)
  {
    START_SPAN(compose_media_server)

    if (media_types.size() != media_ids.size())
    {
      ServiceException se;
      se.errorCode = ErrorCode::SE_THRIFT_HANDLER_ERROR;
      se.message =
          "The lengths of media_id list and media_type list are not equal";
      throw se;
    }

    for (int i = 0; i < media_ids.size(); ++i)
    {
      Media new_media;
      new_media.media_id = media_ids[i];
      new_media.media_type = media_types[i];
      _return.emplace_back(new_media);
    }

    span_compose_media_server->Finish();
  }

} // namespace social_network

#endif // SOCIAL_NETWORK_MICROSERVICES_SRC_MEDIASERVICE_MEDIAHANDLER_H_
