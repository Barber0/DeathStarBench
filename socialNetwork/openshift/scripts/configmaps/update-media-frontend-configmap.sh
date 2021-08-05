#!/bin/bash

NS=social-network

cd $(dirname $0)/../..

$CTL create cm media-frontend-nginx --from-file=media-frontend-config/nginx.conf  --dry-run --save-config -o yaml -n ${NS} | $CTL apply -f - -n ${NS}
$CTL create cm media-frontend-lua   --from-file=media-frontend-config/lua-scripts --dry-run --save-config -o yaml -n ${NS} | $CTL apply -f - -n ${NS}
