add_executable(
    HomeTimelineService
    ${THRIFT_GEN_CPP_DIR}/PostStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/SocialGraphService.cpp
    HomeTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/HomeTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    HomeTimelineService PRIVATE
    /usr/local/include/jaegertracing
    /usr/local/include/hiredis
    /usr/local/include/sw
)

target_link_libraries(
    HomeTimelineService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
    OpenSSL::SSL
    /usr/local/lib/libhiredis.a
    /usr/local/lib/libhiredis_ssl.a
    /usr/local/lib/libredis++.a
    InfluxDB
)

install(TARGETS HomeTimelineService DESTINATION ./)