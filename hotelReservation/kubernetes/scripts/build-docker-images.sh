#!/bin/bash

cd $(dirname $0)/../..

EXEC=docker
USER="simonalphafang"
TAG="0.0.15-simmemc"
ROOT_FOLDER=$(pwd)

PROTO_NAME=hotel_reserve_proto
FULL_PROTO_NAME=$USER/$PROTO_NAME:$TAG

go mod vendor
$EXEC build -t $FULL_PROTO_NAME .
$EXEC push $FULL_PROTO_NAME

rm -rf vendor

#for mod in frontend geo profile rate recommendation reserve search user; do
#  MOD_NAME=hotel_reserve_${mod}
#  FULL_MOD_NAME=$USER/$MOD_NAME:$TAG
#  $EXEC tag $FULL_PROTO_NAME $FULL_MOD_NAME
#  $EXEC push $FULL_MOD_NAME
#done
