#!/usr/bin/env bash

cd $(dirname $0)/..

docker build -t simonalphafang/social_network:0.0.3 -f Dockerfile .