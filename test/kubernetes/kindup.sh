#!/bin/bash

CURR_DIR=$(dirname $0)

kind create cluster --name=loginapp --config=${CURR_DIR}/kind-cluster.yaml

NODE_IP=$(docker inspect loginapp-control-plane -f '{{ .NetworkSettings.Networks.bridge.IPAddress }}')

cat <<EOF
Now you can run:

${CURR_DIR}/genconf.sh $NODE_IP
EOF
