#!/usr/bin/env bash
# Generate Go code from audit_service.proto.
# Usage: ./generate_proto.sh audit_service go

set -e
PROTO_NAME="${1:-audit_service}"
MODE="${2:-go}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="${SCRIPT_DIR}/${PROTO_NAME}"

if ! command -v protoc &>/dev/null; then
  echo "protoc not found. Install Protocol Buffers compiler." >&2
  exit 1
fi

mkdir -p "$OUT_DIR"

# Include path for google/protobuf/*.proto (protoc default or Homebrew)
PROTOC_INCLUDE=""
for d in /usr/local/include /opt/homebrew/include; do
  if [ -f "${d}/google/protobuf/timestamp.proto" ]; then
    PROTOC_INCLUDE="$d"
    break
  fi
done
if [ -z "$PROTOC_INCLUDE" ]; then
  echo "Could not find google/protobuf/timestamp.proto. Install protoc with standard includes." >&2
  exit 1
fi

if [ "$MODE" = "go" ]; then
  protoc \
    -I "$SCRIPT_DIR" \
    -I "$PROTOC_INCLUDE" \
    --go_out="$OUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
    "${SCRIPT_DIR}/${PROTO_NAME}.proto"
  echo "Generated Go code in ${OUT_DIR}"
fi
