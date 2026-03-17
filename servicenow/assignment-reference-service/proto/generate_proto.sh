#!/usr/bin/env bash
# Generate Go code from assignment_reference_service.proto
# Usage: ./generate_proto.sh assignment_reference_service go

set -e
PROTO_NAME="${1:-assignment_reference_service}"
MODE="${2:-go}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="${SCRIPT_DIR}/${PROTO_NAME}"
mkdir -p "$OUT_DIR"

if ! command -v protoc &>/dev/null; then
  echo "protoc not found. Install Protocol Buffers compiler."
  exit 1
fi

if [ "$MODE" = "go" ]; then
  protoc --go_out="$OUT_DIR" --go_opt=paths=source_relative \
    --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
    -I "$SCRIPT_DIR" \
    "${SCRIPT_DIR}/${PROTO_NAME}.proto"
  echo "Generated Go code in ${OUT_DIR}"
else
  echo "Unknown mode: $MODE (use 'go')"
  exit 1
fi
