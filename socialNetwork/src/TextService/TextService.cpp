#include <signal.h>
#include <memory>
#include <string>
#include <grpcpp/ext/proto_server_reflection_plugin.h>
#include <grpcpp/grpcpp.h>
#include <grpcpp/health_check_service_interface.h>

#include "../utils.h"
#include "../utils_thrift.h"
#include "TextHandler.h"

using grpc::Server;
using grpc::ServerBuilder;
using namespace social_network;

void sigintHandler(int sig) { exit(EXIT_SUCCESS); }

int main(int argc, char *argv[])
{
    signal(SIGINT, sigintHandler);
    init_logger();
    SetUpTracer("config/jaeger-config.yml", "text-service");

    json config_json;
    if (load_config_file("config/service-config.json", &config_json) == 0)
    {
        int port = config_json["text-service"]["port"];

        std::string url_addr = config_json["url-shorten-service"]["addr"];
        int url_port = config_json["url-shorten-service"]["port"];
        int url_conns = config_json["url-shorten-service"]["connections"];
        int url_timeout = config_json["url-shorten-service"]["timeout_ms"];
        int url_keepalive = config_json["url-shorten-service"]["keepalive_ms"];

        std::string user_mention_addr = config_json["user-mention-service"]["addr"];
        int user_mention_port = config_json["user-mention-service"]["port"];
        int user_mention_conns = config_json["user-mention-service"]["connections"];
        int user_mention_timeout =
            config_json["user-mention-service"]["timeout_ms"];
        int user_mention_keepalive =
            config_json["user-mention-service"]["keepalive_ms"];

        ClientPool<ThriftClient<UrlShortenServiceClient>> url_client_pool(
            "url-shorten-service", url_addr, url_port, 0, url_conns, url_timeout,
            url_keepalive, config_json);

        ClientPool<ThriftClient<UserMentionServiceClient>> user_mention_pool(
            "user-mention-service", user_mention_addr, user_mention_port, 0,
            user_mention_conns, user_mention_timeout, user_mention_keepalive, config_json);

        // grpc hack
        std::string server_address = "0.0.0.0:" + intToString(port);
        grpc::EnableDefaultHealthCheckService(true);
        grpc::reflection::InitProtoReflectionServerBuilderPlugin();

        ServerBuilder builder;
        builder.AddListeningPort(server_address, grpc::InsecureServerCredentials());

        TextHandler service(&url_client_pool, &user_mention_pool);
        builder.RegisterService(&service);

        std::unique_ptr<Server> server(builder.BuildAndStart());
        std::cout << "Server listening on " << server_address << std::endl;

        server->Wait();
    }
    else
        exit(EXIT_FAILURE);
}
