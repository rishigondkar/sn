# Attachment Service PRD
Version: 1.0  
Service name: attachment-service  
Language: Golang  
Primary responsibility: attachment metadata and object storage integration

## Mission
Manage file attachments linked to cases. Store attachment metadata in the service database, file content in object storage, expose upload/list/delete capabilities, and publish audit events for all attachment changes.


# Agent Execution Policy

This PRD is written for LLMs and coding agents. Treat this section as binding.

## Workflow Orchestration
### 1. Plan Mode by Default
- Enter plan mode for any non-trivial task, especially anything with 3 or more steps, architectural choices, schema changes, interface changes, or migrations.
- If something goes sideways, stop and re-plan immediately. Do not keep pushing a broken approach.
- Use plan mode for verification steps too, not just implementation.
- Write detailed specs up front to reduce ambiguity before changing code.

### 2. Subagent Strategy
- Use subagents liberally to keep the main context window clean.
- Offload research, code reading, dependency analysis, and parallel investigation to subagents.
- For complex problems, use multiple focused subagents in parallel.
- One task per subagent. Keep each subagent narrowly scoped.

### 3. Self-Improvement Loop
- After any user correction, update `tasks/lessons.md`.
- Write a rule that would have prevented the mistake.
- Review `tasks/lessons.md` at the start of any relevant future session.
- Keep reducing repeated mistakes; do not repeat the same error pattern.

### 4. Verification Before Done
- Never mark a task complete without proving it works.
- Compare before and after behavior when relevant.
- Ask: "Would a staff engineer approve this?"
- Run tests, inspect logs, validate edge cases, and demonstrate correctness.

### 5. Demand Elegance, But Stay Balanced
- For non-trivial changes, pause and ask whether there is a more elegant solution.
- If the current fix feels hacky, step back and implement the cleaner design.
- Do not over-engineer simple fixes.
- Challenge your own work before presenting it.

### 6. Autonomous Bug Fixing
- When given a bug report, fix it. Do not ask the user to do exploratory debugging for you.
- Start from logs, errors, stack traces, and failing tests.
- Minimize user context switching.
- If CI is failing, investigate and repair the failures without being told exactly how.

## Task Management Rules
1. Create `tasks/todo.md` before implementation with checkable items.
2. Check in against the plan before beginning implementation.
3. Mark items complete as work progresses.
4. Add high-level summaries of changes as work proceeds.
5. Add a review/results section to `tasks/todo.md` before concluding.
6. Update `tasks/lessons.md` after corrections or notable mistakes.

## Core Engineering Principles
- **Simplicity First:** implement the simplest design that correctly satisfies the requirements.
- **No Laziness:** fix root causes; do not ship temporary hacks as final solutions.
- **Minimal Impact:** change only what is necessary and preserve existing behavior unless the PRD explicitly requires otherwise.

## Global Do-Not-Do List
- Do not invent APIs, events, or schema fields that are not specified in this PRD or the shared platform contract.
- Do not silently change shared contracts.
- Do not couple directly to another service's database.
- Do not bypass the event bus for audit publication.
- Do not mark a task done if tests are missing or failing.
- Do not skip idempotency, validation, error handling, or observability.
- Do not create hidden behavior such as implicit external lookups unless explicitly required.
- Do not use ad hoc JSON blobs where strongly typed fields are required by this PRD.
- Do not expose internal gRPC APIs directly to the public internet.
- Do not break backward compatibility in REST or gRPC without updating the shared contract and migration guidance.

## Golang Standards
- Language: Go 1.22+.
- Prefer standard library first; add third-party dependencies only with clear justification.
- Use context propagation on all inbound and outbound calls.
- Use structured logging.
- Use table-driven tests.
- Use dependency injection via interfaces where it improves testability.
- Separate transport, service, repository, and domain logic layers.
- Validate all request payloads before entering business logic.
- Apply timeouts to all outbound network calls.
- Propagate correlation IDs, request IDs, and actor metadata.


# In Scope
- upload attachment metadata and content
- list attachments by case
- delete attachment
- optional metadata update
- object storage integration
- publish audit events

# Out of Scope
- malware scanning unless explicitly added later
- arbitrary document management unrelated to cases
- owning case lifecycle
- audit event persistence

# Owned Data Model

## `attachments`
- `id` UUID PK
- `case_id` UUID not null
- `file_name` VARCHAR(500) not null
- `file_size_bytes` BIGINT not null
- `content_type` VARCHAR(255) nullable
- `storage_provider` VARCHAR(50) not null default `s3`
- `storage_key` VARCHAR(1000) not null
- `storage_bucket` VARCHAR(255) nullable
- `uploaded_by_user_id` UUID not null
- `uploaded_at` TIMESTAMPTZ not null
- `is_deleted` BOOLEAN not null default false
- `deleted_at` TIMESTAMPTZ nullable
- `metadata` JSONB nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Business Rules

## Upload
- each attachment belongs to a case
- store binary in object storage
- store metadata row after or with successful upload transaction design
- define safe rollback/compensation behavior if DB write or storage write fails
- publish `attachment.uploaded`

## Delete
- default behavior: soft delete metadata and delete object or mark for object deletion according to retention policy
- publish `attachment.deleted`
- deleted attachments should not appear in default list queries

## Allowed File Types and Limits
Configure:
- max file size
- allowed content types or deny list
- filename sanitation
These rules must be explicit in config and tested.

## Security
- never trust client-provided content type alone
- sanitize filenames
- do not allow path traversal semantics in storage keys
- generate storage keys server-side
- restrict direct bucket exposure

# REST API Surface
- `POST /api/v1/attachments`
- `GET /api/v1/cases/{caseId}/attachments`
- `DELETE /api/v1/attachments/{attachmentId}`

## Upload Flow (POC)
- V1 uses **gateway-mediated upload**:
  - Client sends a multipart `POST /api/v1/attachments` request to the API Gateway.
  - Gateway validates metadata and streams file data (or an equivalent payload) to Attachment Service via gRPC/REST.
  - Attachment Service writes the binary to object storage and persists metadata, then publishes audit events.
- A signed-URL workflow may be added in a later phase; if adopted, both Gateway and Attachment Service PRDs must be updated to describe the new flow.

# gRPC API Contracts
- `CreateAttachment`
- `ListAttachmentsByCase`
- `DeleteAttachment`

# Integration Dependencies
- Case Service for validating case existence if strict validation is enabled
- object storage provider
- Audit via pub/sub

# Audit Events To Publish
- `attachment.uploaded`
- `attachment.deleted`
- `attachment.metadata.updated` if metadata edits are supported

# Validation Rules
- case_id required
- uploaded_by_user_id required
- file size must match actual uploaded content size if available
- reject prohibited content types
- reject oversized uploads

# Non-Functional Requirements
- reliable storage writes
- clear retry strategy for transient storage failures
- strong logging of upload failures without leaking secrets
- object storage operations must use timeouts

# Implementation Guidance for Agents
1. Decide and document upload transaction/compensation order.
2. Implement storage client abstraction for testability.
3. Add repository after storage contract is stable.
4. Add integration tests against local object storage emulator if possible.
5. Verify audit event publication on upload/delete.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). Full layout with a **storage abstraction** for object storage (e.g. S3); no binary data in DB.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   └── attachment.go
├── api/
│   ├── api.go
│   └── attachment_handlers.go
├── service/
│   ├── service.go
│   └── attachment.go
├── repository/
│   └── repository.go
├── storage/
│   └── storage.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── attachment_service.proto
│   ├── generate_proto.sh
│   └── attachment_service/
│       ├── attachment_service.pb.go
│       └── attachment_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `handler/` and `api/`: gRPC and optional REST for CreateAttachment, ListAttachmentsByCase, DeleteAttachment; delegate to Service.
- `service/`: Upload flow (storage then DB or documented compensation); delete (soft-delete metadata + storage delete/mark); audit events after success.
- `repository/`: attachments table only; no binary content.
- `storage/`: Interface and implementation for object storage (Put, Get, Delete); timeouts and retries; server-generated keys; no path traversal.
- `proto/attachment_service.proto`: `go_package = "<module>/proto/attachment_service"`; CreateAttachment, ListAttachmentsByCase, DeleteAttachment per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/attachment-service` |
| Proto | `attachment_service.proto`; go_package `<module>/proto/attachment_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health + optional REST for upload/list/delete) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint; graceful shutdown; gRPC/REST error mapping; timeouts on all outbound calls (DB, storage); upload transaction/compensation documented and tested; audit events on upload/delete; no binaries in DB; no secrets in logs.

## Agent Checklist

- [ ] Set module name and proto go_package; run proto generation.
- [ ] Implement storage interface and implementation (timeouts, server-side keys).
- [ ] Implement repository (attachments metadata); migrations.
- [ ] Implement service (upload order, compensation, delete, audit publisher).
- [ ] Implement handler and optional api; main.go: gRPC + HTTP, graceful shutdown.
- [ ] Tests: upload/delete, compensation path, storage abstraction mock, handler tests.
- [ ] Update README.

# Do Not Do
- Do not store binaries in the relational database.
- Do not let clients choose raw storage keys.
- Do not bypass audit event publication.
- Do not expose internal bucket details unnecessarily.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE attachments (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  file_name VARCHAR(500) NOT NULL,
  file_size_bytes BIGINT NOT NULL,
  content_type VARCHAR(255),
  storage_provider VARCHAR(50) NOT NULL DEFAULT 's3',
  storage_key VARCHAR(1000) NOT NULL,
  storage_bucket VARCHAR(255),
  uploaded_by_user_id UUID NOT NULL,
  uploaded_at TIMESTAMPTZ NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT false,
  deleted_at TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package attachment_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/attachment_service";

service AttachmentService {
  rpc CreateAttachment(CreateAttachmentRequest) returns (CreateAttachmentResponse);
  rpc ListAttachmentsByCase(ListAttachmentsByCaseRequest) returns (ListAttachmentsByCaseResponse);
  rpc DeleteAttachment(DeleteAttachmentRequest) returns (DeleteAttachmentResponse);
}

message CreateAttachmentRequest {
  string case_id = 1;
  string file_name = 2;
  string content_type = 3;
  string uploaded_by_user_id = 4;
  bytes content = 5;
}

message CreateAttachmentResponse {
  string id = 1;
  Attachment attachment = 2;
}

message ListAttachmentsByCaseRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  bool include_deleted = 4;
}

message ListAttachmentsByCaseResponse {
  repeated Attachment items = 1;
  string next_page_token = 2;
}

message DeleteAttachmentRequest {
  string attachment_id = 1;
}

message DeleteAttachmentResponse {}

message Attachment {
  string id = 1;
  string case_id = 2;
  string file_name = 3;
  int64 file_size_bytes = 4;
  string content_type = 5;
  string storage_provider = 6;
  string uploaded_by_user_id = 7;
  google.protobuf.Timestamp uploaded_at = 8;
  bool is_deleted = 9;
  google.protobuf.Timestamp deleted_at = 10;
  string metadata_json = 11;
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}
```
