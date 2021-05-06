#!/bin/sh

k8sEndpoint="unix:/tmp/k8s.sock" # "127.0.0.1:5123"

go run *.go \
  "$k8sEndpoint" \
  "$@"
