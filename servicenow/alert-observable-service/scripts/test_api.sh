#!/usr/bin/env bash
# Run alert-observable-service API integration tests.
# Prereqs: Postgres running with DB from migrations; start the server in another terminal:
#   go run ./cmd/ - then run this script.
# Or: make sure TEST_GRPC_ADDR is set if server runs on a different host:port.

set -e
cd "$(dirname "$0")/.."
export TEST_GRPC_ADDR="${TEST_GRPC_ADDR:-localhost:50052}"
echo "Testing gRPC at $TEST_GRPC_ADDR ..."
go test -v -count=1 ./integration/ -timeout 60s
