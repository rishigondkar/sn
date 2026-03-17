# Attachment Service – Implementation Plan

**Source:** `06_attachment_service_prd.md` + `00_platform_integration_contract.md`  
**Policy:** Plan mode first; no ad-hoc changes. Check off as work completes.

---

## 1. Project layout (per PRD Go Project Structure)

- [ ] Create directory tree: `cmd/`, `handler/`, `api/`, `service/`, `repository/`, `storage/`, `config/`, `logging/`, `proto/`, `migrations/`.
- [ ] `go.mod`: module `github.com/soc-platform/attachment-service`, Go 1.23; grpc, protobuf, testify, DB driver (e.g. pgx), optional router for REST.
- [ ] `Dockerfile`, `.gitlab-ci.yml`, `README.md` stubs.

---

## 2. Proto (Appendix B)

- [ ] Add `proto/attachment_service.proto` with exact Appendix B content; `go_package = "github.com/soc-platform/attachment-service/proto/attachment_service"`.
- [ ] Add `proto/generate_proto.sh`; run to produce `proto/attachment_service/*.pb.go` and `*_grpc.pb.go`.

---

## 3. Migrations (Appendix A)

- [ ] Add `migrations/` with initial migration: `attachments` table DDL from Appendix A (id, case_id, file_name, file_size_bytes, content_type, storage_provider, storage_key, storage_bucket, uploaded_by_user_id, uploaded_at, is_deleted, deleted_at, metadata, created_at, updated_at).
- [ ] Document migration run (e.g. in README); no cross-service FKs.

---

## 4. Storage abstraction

- [ ] `storage/storage.go`: interface `Store` with `Put(ctx, key, bucket, content []byte) error`, `Get(ctx, key, bucket) ([]byte, error)`, `Delete(ctx, key, bucket) error`; server-generated keys only; no path traversal.
- [ ] Implementation (e.g. `storage/s3.go` or in-memory for tests): timeouts and retries per config; document key generation (e.g. `caseID/attachmentID/filename`).
- [ ] Config: storage provider, bucket, timeouts, retries.

---

## 5. Repository

- [ ] `repository/repository.go`: struct with DB pool; `currentLayer` for logs.
- [ ] Methods: `CreateAttachment`, `GetByID`, `ListByCaseID` (exclude `is_deleted=true` unless `include_deleted`), `SoftDelete`, optional `UpdateMetadata`; domain struct for attachment (no proto in repo).
- [ ] All outbound DB calls with context and timeouts.

---

## 6. Audit event publisher

- [ ] Abstract interface to publish audit event envelope (JSON) per platform contract: `event_id`, `event_type`, `source_service`, `entity_type`, `entity_id`, `action`, `actor_*`, `request_id`, `correlation_id`, `occurred_at`, etc.
- [ ] Implementations: no-op for tests; real implementation (e.g. SNS/SQS or stub) that publishes to event bus; document in README.
- [ ] Service layer calls publisher after successful commit for: `attachment.uploaded`, `attachment.deleted`, and `attachment.metadata.updated` if metadata update is implemented.

---

## 7. Service layer

- [ ] `service/service.go`: Repository, Storage, AuditPublisher; optional Case Service client if strict case validation is added later.
- [ ] `service/attachment.go`:
  - **CreateAttachment:** validate (case_id, uploaded_by_user_id, file size, content type); generate UUID and storage key; **upload to storage first**, then insert metadata row; on DB failure document compensation (e.g. delete object or mark for cleanup); publish `attachment.uploaded`.
  - **ListAttachmentsByCase:** delegate to repository with pagination (page_size, page_token).
  - **DeleteAttachment:** soft-delete row, then delete object (or mark for deletion); publish `attachment.deleted`.
  - Optional: **UpdateAttachmentMetadata** → publish `attachment.metadata.updated`.
- [ ] Config: max file size, allowed content types (or deny list), filename sanitization rules; all validated and tested.

---

## 8. Handlers and API

- [ ] **gRPC (`handler/`):** `handler.go` embeds `UnimplementedAttachmentServiceServer`, holds `*service.Service`; `attachment.go` implements `CreateAttachment`, `ListAttachmentsByCase`, `DeleteAttachment`; extract actor/request-id/correlation-id from metadata; map errors to gRPC codes (InvalidArgument, NotFound, Internal).
- [ ] **REST (`api/`):** `api.go` router with `/api/v1` prefix; `attachment_handlers.go`: `POST /api/v1/attachments`, `GET /api/v1/cases/{caseId}/attachments`, `DELETE /api/v1/attachments/{attachmentId}`; same error shape and status codes per platform contract; parse `X-User-Id`, `X-Request-Id`, `X-Correlation-Id` where applicable.

---

## 9. main.go – wiring and P0

- [ ] Load config (env); init logging (structured, no secrets); create DB pool and repository; create storage implementation; create audit publisher; create service; create gRPC handler and register `AttachmentService`; create HTTP server with REST routes + `GET /health` (200 ok or 503 if deps down); gRPC port 50051, HTTP 8080; server read/write timeouts; graceful shutdown on SIGTERM/SIGINT (drain gRPC + HTTP, then exit with timeout).

---

## 10. Tests

- [ ] Unit: service layer with mocked repository, storage, and audit publisher; upload success, upload DB failure (compensation path), delete success, validation failures (size, content type).
- [ ] Repository tests: create, get, list by case, soft delete (use test DB or in-memory if applicable).
- [ ] Handler tests: gRPC (and optionally REST) with in-process server or test client; happy path and failure paths per RPC/endpoint.
- [ ] Storage: mock for service tests; optional integration test with local emulator if available.

---

## 11. PRD-specific and checklist

- [ ] Do not store binary in DB; do not allow client-chosen storage keys; do not bypass audit; do not expose bucket details.
- [ ] Filename sanitization and path-traversal prevention in storage key generation.
- [ ] Idempotency: consider `Idempotency-Key` for POST if required by platform (document if deferred).
- [ ] README: project overview, layout, setup (proto gen, migrations, env), run, test, REST vs gRPC, event bus config.

---

## Review / results (fill before concluding)

- **Build:** `go build -o /tmp/attachment-service ./cmd` succeeds.
- **Tests:** `go test ./...` passes (service, handler, storage).
- **Agent Checklist (PRD):** Module and proto go_package set; proto generated. Storage interface + memory impl. Repository + migrations. Service (upload order, compensation, delete, audit). Handler + API; main with gRPC, HTTP, health, graceful shutdown. Tests for upload/delete, compensation, handler. README updated.
- **Remaining:** Optional: repository integration test against real Postgres; real S3 storage impl; real audit publisher (SNS/SQS); Idempotency-Key for POST. Migration run is manual (psql or migrate tool).
