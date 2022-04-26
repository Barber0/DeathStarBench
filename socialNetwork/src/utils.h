#ifndef SOCIAL_NETWORK_MICROSERVICES_UTILS_H
#define SOCIAL_NETWORK_MICROSERVICES_UTILS_H

#include <string>
#include <fstream>
#include <iostream>
#include <nlohmann/json.hpp>
#include <cstdio>

#include "logger.h"

namespace social_network
{
  using json = nlohmann::json;

  int load_config_file(const std::string &file_name, json *config_json)
  {
    std::ifstream json_file;
    json_file.open(file_name);
    if (json_file.is_open())
    {
      json_file >> *config_json;
      json_file.close();
      return 0;
    }
    else
    {
      LOG(error) << "Cannot open service-config.json";
      return -1;
    }
  };

  std::string intToString(int v)
  {
    char buf[32] = {0};
    snprintf(buf, sizeof(buf), "%u", v);

    std::string str = buf;
    return str;
  };

} // namespace social_network

#endif // SOCIAL_NETWORK_MICROSERVICES_UTILS_H
