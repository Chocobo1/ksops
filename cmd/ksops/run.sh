#!/bin/sh

ksopsEndpoint="unix:/tmp/ksops.sock"  # "tcp:127.0.0.1:5123"
sopsEndpoint="unix:/tmp/sops.sock"  # "127.0.0.1:6123"

go run *.go \
  --pgp.key="0xF5380EA79A2C21C3687500B6BB350F9E0D2B1137" \
  "$ksopsEndpoint" \
  "$sopsEndpoint" \
  "$@"
