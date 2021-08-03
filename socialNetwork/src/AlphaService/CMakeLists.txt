add_executable(
    AlphaService
    main.cc
)

target_link_libraries(
    AlphaService
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    nlohmann_json::nlohmann_json
    Boost::log
    Boost::log_setup
    OpenSSL::SSL
)

install(TARGETS AlphaService DESTINATION ./)