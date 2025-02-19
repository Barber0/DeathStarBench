#!/bin/bash

cd $(dirname $0)/../..

# This script creates an OpenShift ConfigMap for all the services
# built upon the C++ jaeger client, which uses the jaeger-config.yml
# to find the jaeger end-point URL.
$CTL create cm jaeger-config-yaml  --from-file=config/jaeger-config.yml               -n social-network

# Since the nginx-thrift service is not built upon the C++ jaeger client,
# this service requires the jaeger-config.json in a different format than
# the one in the ConfigMap jaeger-config. Then, we create a new ConfigMap.
$CTL create cm nginx-thrift-jaeger --from-file=nginx-thrift-config/jaeger-config.json -n social-network

