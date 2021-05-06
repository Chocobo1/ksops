#!/bin/sh

sops \
  keyservice \
  --network unix \
  --address "/tmp/sops.sock" \
  >/dev/null \
  2>&1 \
  &

(
  cd "../cmd/ksops"
  go build
  ./ksops \
    --pgp.key="FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4" \
    "unix:/tmp/ksops.sock" \
    "unix:/tmp/sops.sock" \
    >/dev/null \
    2>&1 \
    &
)

go run \
  "client/client.go" \
  "unix:/tmp/ksops.sock"
err="$?"

killall ksops sops
rm "../cmd/ksops/ksops"

exit "$err"
