# Audit Service PRD
Version: 1.0  
Service name: audit-service  
Language: Golang  
Primary responsibility: consume pub/sub events and provide immutable audit history query APIs

## Mission
Provide a centralized, append-only audit trail for the platform. Consume audit events published by other services, persist them reliably, support filtering and timeline retrieval, and remain decoupled from business write paths.


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
- subscribe to event bus
- validate and persist audit event envelopes
- idempotent event consumption
- query audit history by case, observable, entity, actor, and time range
- expose query APIs
- support timeline reconstruction

# Out of Scope
- participating synchronously in domain write flows
- acting as source of truth for business entities
- modifying business state
- creating audit events on behalf of other services except internal system events related to its own operation if needed

# Owned Data Model

## `audit_events`
- `id` UUID PK
- `event_id` VARCHAR(100) unique not null
- `event_type` VARCHAR(100) not null
- `source_service` VARCHAR(100) not null
- `entity_type` VARCHAR(100) not null
- `entity_id` UUID not null
- `parent_entity_type` VARCHAR(100) nullable
- `parent_entity_id` UUID nullable
- `case_id` UUID nullable
- `observable_id` UUID nullable
- `action` VARCHAR(50) not null
- `actor_user_id` UUID nullable
- `actor_name` VARCHAR(255) nullable
- `request_id` VARCHAR(100) nullable
- `correlation_id` VARCHAR(100) nullable
- `change_summary` VARCHAR(1000) nullable
- `before_data` JSONB nullable
- `after_data` JSONB nullable
- `metadata` JSONB nullable
- `occurred_at` TIMESTAMPTZ not null
- `ingested_at` TIMESTAMPTZ not null

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Event Consumption Rules
- consume from pub/sub only
- each message body MUST be the audit event envelope JSON defined in the platform contract (no wrapper object)
- validate required envelope fields
- dedupe using `event_id`
- preserve payload faithfully
- failed events should go to dead-letter handling after configured retries
- no synchronous dependency on producing services during consumption

# Query APIs
Required queries:
- list audit events by case id
- list audit events by observable id
- list audit events by entity type + entity id
- list audit events by actor user id
- list audit events by correlation id
- list audit events in time range

Sort newest first by default, with optional oldest-first for timeline reconstruction.

# gRPC API Contracts
- `ListAuditEventsByCase`
- `ListAuditEventsByObservable`
- `ListAuditEventsByEntity`
- `ListAuditEventsByActor`
- `ListAuditEventsByCorrelationId`

# REST API Surface
Served through gateway:
- `GET /api/v1/cases/{caseId}/audit-events`
- `GET /api/v1/observables/{observableId}/audit-events`
- `GET /api/v1/audit-events?entity_type=...&entity_id=...`

# Non-Functional Requirements
- append-only persistence
- resilient consumer behavior
- event ingestion idempotency
- efficient timeline queries
- support large event volumes over time

# Indexing Recommendations
- unique index on `event_id`
- index on `case_id`
- index on `observable_id`
- index on (`entity_type`, `entity_id`)
- index on `actor_user_id`
- index on `correlation_id`
- index on `occurred_at`

# Implementation Guidance for Agents
1. Implement event envelope validation first.
2. Implement idempotent consumer second.
3. Add dead-letter and retry behavior.
4. Implement query repository with strong indexes.
5. Add tests for duplicate events, malformed events, and query filters.
6. Keep this service append-only.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). This service **deviates**: gRPC **query** server only (no domain command RPCs); plus an **event consumer** that subscribes to pub/sub and persists audit events idempotently.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   └── audit_query.go
├── service/
│   ├── service.go
│   └── audit_query.go
├── repository/
│   └── repository.go
├── consumer/
│   └── consumer.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── audit_service.proto
│   ├── generate_proto.sh
│   └── audit_service/
│       ├── audit_service.pb.go
│       └── audit_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `cmd/main.go`: Start gRPC server and **event consumer** (e.g. goroutine or worker); on shutdown, drain both (consumer stop, then gRPC GracefulStop, HTTP Shutdown).
- `handler/`: gRPC **query** handlers only (ListAuditEventsByCase, ListAuditEventsByObservable, ListAuditEventsByEntity, ListAuditEventsByActor, ListAuditEventsByCorrelationId); delegate to Service.
- `service/` and `repository/`: Query logic and audit_events table; indexes per Indexing Recommendations.
- `consumer/`: Subscribe to event bus; validate envelope; dedupe by `event_id`; persist; dead-letter and retry on failure; no synchronous callbacks to producing services.
- **No** `api/` unless REST is added later (gateway typically calls this service via gRPC only).
- `proto/audit_service.proto`: `go_package = "<module>/proto/audit_service"`; **query RPCs only** per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/audit-service` |
| Proto | `audit_service.proto`; go_package `<module>/proto/audit_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health only) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint; graceful shutdown (drain consumer then gRPC/HTTP); idempotent event consumption by event_id; dead-letter and retry for failed messages; no sync callbacks to producers; gRPC error mapping; timeouts; no secrets in logs.

## Agent Checklist

- [ ] Set module name and proto go_package; run proto generation (query RPCs only).
- [ ] Implement repository (audit_events table, indexes); migrations.
- [ ] Implement consumer: subscribe, validate envelope, idempotent insert, retry/dead-letter.
- [ ] Implement service and handler for query RPCs.
- [ ] main.go: start gRPC server + consumer; graceful shutdown for both.
- [ ] Tests: duplicate event idempotency, malformed event handling, query filters.
- [ ] Update README.

# Do Not Do
- Do not call back synchronously into source services for ingestion.
- Do not mutate or enrich business facts at ingestion time beyond safe indexing/normalization.
- Do not become a reporting engine in v1.
- Do not participate in business transactions.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE audit_events (
  id UUID PRIMARY KEY,
  event_id VARCHAR(100) NOT NULL UNIQUE,
  event_type VARCHAR(100) NOT NULL,
  source_service VARCHAR(100) NOT NULL,
  entity_type VARCHAR(100) NOT NULL,
  entity_id UUID NOT NULL,
  parent_entity_type VARCHAR(100),
  parent_entity_id UUID,
  case_id UUID,
  observable_id UUID,
  action VARCHAR(50) NOT NULL,
  actor_user_id UUID,
  actor_name VARCHAR(255),
  request_id VARCHAR(100),
  correlation_id VARCHAR(100),
  change_summary VARCHAR(1000),
  before_data JSONB,
  after_data JSONB,
  metadata JSONB,
  occurred_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_events_case_id ON audit_events(case_id);
CREATE INDEX idx_audit_events_observable_id ON audit_events(observable_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_actor_user_id ON audit_events(actor_user_id);
CREATE INDEX idx_audit_events_correlation_id ON audit_events(correlation_id);
CREATE INDEX idx_audit_events_occurred_at ON audit_events(occurred_at);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package audit_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/audit_service";

service AuditService {
  rpc ListAuditEventsByCase(ListAuditEventsByCaseRequest) returns (ListAuditEventsResponse);
  rpc ListAuditEventsByObservable(ListAuditEventsByObservableRequest) returns (ListAuditEventsResponse);
  rpc ListAuditEventsByEntity(ListAuditEventsByEntityRequest) returns (ListAuditEventsResponse);
  rpc ListAuditEventsByActor(ListAuditEventsByActorRequest) returns (ListAuditEventsResponse);
  rpc ListAuditEventsByCorrelationId(ListAuditEventsByCorrelationIdRequest) returns (ListAuditEventsResponse);
}

message ListAuditEventsByCaseRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional google.protobuf.Timestamp occurred_after = 4;
  optional google.protobuf.Timestamp occurred_before = 5;
  bool oldest_first = 6;
}

message ListAuditEventsByObservableRequest {
  string observable_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional google.protobuf.Timestamp occurred_after = 4;
  optional google.protobuf.Timestamp occurred_before = 5;
  bool oldest_first = 6;
}

message ListAuditEventsByEntityRequest {
  string entity_type = 1;
  string entity_id = 2;
  int32 page_size = 3;
  string page_token = 4;
  optional google.protobuf.Timestamp occurred_after = 5;
  optional google.protobuf.Timestamp occurred_before = 6;
  bool oldest_first = 7;
}

message ListAuditEventsByActorRequest {
  string actor_user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional google.protobuf.Timestamp occurred_after = 4;
  optional google.protobuf.Timestamp occurred_before = 5;
  bool oldest_first = 6;
}

message ListAuditEventsByCorrelationIdRequest {
  string correlation_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional google.protobuf.Timestamp occurred_after = 4;
  optional google.protobuf.Timestamp occurred_before = 5;
  bool oldest_first = 6;
}

message ListAuditEventsResponse {
  repeated AuditEvent items = 1;
  string next_page_token = 2;
}

message AuditEvent {
  string id = 1;
  string event_id = 2;
  string event_type = 3;
  string source_service = 4;
  string entity_type = 5;
  string entity_id = 6;
  string parent_entity_type = 7;
  string parent_entity_id = 8;
  string case_id = 9;
  string observable_id = 10;
  string action = 11;
  string actor_user_id = 12;
  string actor_name = 13;
  string request_id = 14;
  string correlation_id = 15;
  string change_summary = 16;
  string before_data_json = 17;
  string after_data_json = 18;
  string metadata_json = 19;
  google.protobuf.Timestamp occurred_at = 20;
  google.protobuf.Timestamp ingested_at = 21;
}
```
