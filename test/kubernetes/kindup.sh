#!/bin/bash

CURR_DIR=$(dirname $0)

kind create cluster --name=loginapp --config=${CURR_DIR}/kind-cluster.yaml
