#!/bin/bash

CURR_DIR=$(dirname $0)
if [[ $OSTYPE == 'darwin'* ]]; then
  cat << EOF > ${CURR_DIR}/generated/kind-cluster.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.22.2
  extraPortMappings:
    - containerPort: 32000
      hostPort: 32000
EOF
else
  cat << EOF > ${CURR_DIR}/generated/kind-cluster.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: kindest/node:v1.22.2
EOF
fi

kind create cluster --name=loginapp --config=${CURR_DIR}/generated/kind-cluster.yaml
