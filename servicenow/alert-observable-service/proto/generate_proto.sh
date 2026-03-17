#!/usr/bin/env bash
# Generate Go stubs from alert_observable_service.proto.
# Usage: ./generate_proto.sh alert_observable_service go
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
NAME="${1:-alert_observable_service}"
OUT_DIR="${SCRIPT_DIR}/${NAME}"
mkdir -p "$OUT_DIR"
protoc --go_out="$OUT_DIR" --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
  -I . \
  "${NAME}.proto"
