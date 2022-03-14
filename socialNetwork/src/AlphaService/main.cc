#ifndef ALPHA
#define ALPHA

#include <iostream>
#include <thrift/transport/TServerSocket.h>
#include <memory>

using ::apache::thrift::transport::TServerSocket;

int main(int argc, char const *argv[])
{
    std::cout << "alpha" << std::endl;
    auto serverSocket = std::make_shared<TServerSocket>(
        "0.0.0.0",
        8080);
    return 0;
}

#endif