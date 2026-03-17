# Attachment Service

Part of the multi-service security operations platform. Manages file attachments linked to cases: stores metadata in the service database, file content in object storage, and publishes audit events for all attachment changes.

## Features

- **gRPC (service-to-service):** CreateAttachment, ListAttachmentsByCase, DeleteAttachment
- **REST (UI):** POST /api/v1/attachments, GET /api/v1/cases/{caseId}/attachments, DELETE /api/v1/attachments/{attachmentId}
- **Health:** GET /health returns 200 when the process is up
- **Audit events:** attachment.uploaded, attachment.deleted (to event bus per platform contract)

## Layout

```
├── cmd/main.go
├── handler/          # gRPC handlers
├── api/              # REST handlers
├── service/          # Business logic
├── repository/       # Attachments table (metadata only)
├── storage/          # Object storage abstraction (Put, Get, Delete)
├── audit/            # Audit event publisher
├── config/           # Env-based config
├── logging/
├── proto/            # attachment_service.proto + generated Go
├── migrations/       # PostgreSQL DDL (attachments table)
├── domain/           # Domain types
└── go.mod
```

## Prerequisites

- Go 1.23+
- PostgreSQL (for attachments metadata)
- `protoc` with Go and gRPC plugins (for proto generation)

## Setup

1. Clone and enter the repo:
   ```bash
   cd attachment-service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Generate gRPC stubs:
   ```bash
   ./proto/generate_proto.sh attachment_service go
   ```

4. Run migrations (Appendix A):
   ```bash
   psql $DATABASE_URL -f migrations/000001_create_attachments.up.sql
   ```
   Or use your migration runner (e.g. golang-migrate) pointing at `migrations/`.

## Configuration (environment)

| Variable | Description | Default |
|----------|-------------|---------|
| DATABASE_URL | PostgreSQL connection string | (required) |
| GRPC_PORT | gRPC server port | 50051 |
| HTTP_PORT | HTTP server (REST + health) port | 8080 |
| STORAGE_PROVIDER | Object storage provider | s3 |
| STORAGE_BUCKET | Bucket name | attachments |
| MAX_FILE_SIZE_BYTES | Max upload size | 52428800 (50 MiB) |
| ALLOWED_CONTENT_TYPES | Comma-separated allowed types (empty = all) | (empty) |
| DENIED_CONTENT_TYPES | Comma-separated denied types | (empty) |
| SHUTDOWN_TIMEOUT | Graceful shutdown deadline | 30s |

## Run

```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
go run ./cmd
```

- gRPC: `localhost:50051`
- HTTP: `localhost:8080` (GET /health, POST /api/v1/attachments, etc.)

## Tests

```bash
go test ./...
```

- **Service:** Unit tests with mocked repository, storage, and audit (upload success, validation, compensation on DB failure, delete).
- **Handler:** In-process gRPC server tests (Create, List, Delete happy path and NotFound).
- **Storage:** Memory store Put/Get/Delete and key generator.

## Upload flow (POC)

1. Client sends multipart POST to API Gateway (or direct to this service).
2. Attachment Service receives content (gRPC bytes or REST body).
3. **Storage first:** Write binary to object storage (server-generated key).
4. **Then DB:** Insert metadata row. On DB failure, compensation: delete object from storage.
5. Publish `attachment.uploaded` audit event.

## Audit event bus

Events follow the platform audit envelope (JSON). For POC, the audit publisher is a no-op; plug in SNS/SQS or Kafka and document topic/queue in the deployment README.

## Do not

- Store binary content in the relational database.
- Allow clients to choose raw storage keys.
- Bypass audit event publication for upload/delete.
- Expose internal bucket details to clients.
