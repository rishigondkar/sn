#!/usr/bin/env bash
# Generate Go code from enrichment_threat_service.proto.
# Usage: ./generate_proto.sh [proto_name] [go|descriptor]
# Example: ./generate_proto.sh enrichment_threat_service go

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

PROTO_NAME="${1:-enrichment_threat_service}"
MODE="${2:-go}"

if ! command -v protoc &>/dev/null; then
  echo "protoc not found; install Protocol Buffers compiler" >&2
  exit 1
fi

OUT_DIR="./${PROTO_NAME}"
mkdir -p "$OUT_DIR"

if [ "$MODE" = "go" ]; then
  protoc --go_out="$OUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
    -I. \
    "${PROTO_NAME}.proto"
  echo "Generated Go code in $OUT_DIR"
else
  echo "Unknown mode: $MODE (use 'go' or 'descriptor')" >&2
  exit 1
fi
