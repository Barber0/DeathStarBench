
add_executable(
    TextService
    TextService.cpp
    ${GRPC_GEN_CPP_DIR}/social_network.grpc.pb.cc
    ${GRPC_GEN_CPP_DIR}/social_network.pb.cc
    ${THRIFT_GEN_CPP_DIR}/TextService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserMentionService.cpp
    ${THRIFT_GEN_CPP_DIR}/UrlShortenService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    TextService PRIVATE
    /usr/local/include/jaegertracing
)

target_link_libraries(
    TextService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
    gRPC::grpc++
    ${_REFLECTION}
)

install(TARGETS TextService DESTINATION ./)