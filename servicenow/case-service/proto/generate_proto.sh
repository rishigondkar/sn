#!/usr/bin/env bash
# Generate Go code from case_service.proto.
# Usage: ./generate_proto.sh case_service go

set -e
PROTO_NAME="${1:-case_service}"
MODE="${2:-go}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUT_DIR="${SCRIPT_DIR}/${PROTO_NAME}"
mkdir -p "$OUT_DIR"

if ! command -v protoc &>/dev/null; then
  echo "protoc not found; install Protocol Buffers compiler"
  exit 1
fi

# Prefer Go plugin from PATH; typical install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
export PATH="${PATH}:$(go env GOPATH)/bin"

case "$MODE" in
  go)
    protoc -I "$SCRIPT_DIR" \
      --go_out="$OUT_DIR" --go_opt=paths=source_relative \
      --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
      "${SCRIPT_DIR}/${PROTO_NAME}.proto"
    echo "Generated Go code in ${OUT_DIR}/"
    ;;
  *)
    echo "Usage: $0 <proto_file_without_extension> go"
    exit 1
    ;;
esac
