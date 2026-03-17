# Case Service PRD
Version: 1.0  
Service name: case-service  
Language: Golang  
Primary responsibility: source of truth for case lifecycle and worknotes

## Mission
Own the lifecycle of security incidents/cases. Manage creation, updates, assignment metadata references, state transitions, worknotes, closure state, active duration, and case-level analyst determinations.


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
- create/read/update cases
- manage case state transitions
- assignment updates
- case closure
- worknotes
- active duration tracking
- case-level accuracy, determination, and impact
- publish audit events for all writes

# Out of Scope
- observables and child observables
- alerts
- threat lookups and enrichment payload storage
- attachment binary storage
- audit event persistence
- user/group ownership beyond storing reference IDs

# Owned Data Model

## `cases`
Fields:
- `id` UUID PK
- `case_number` VARCHAR unique not null
- `short_description` VARCHAR(500) not null
- `description` TEXT nullable
- `state` VARCHAR(50) not null
- `priority` VARCHAR(20) not null
- `severity` VARCHAR(20) not null
- `opened_by_user_id` UUID not null
- `opened_time` TIMESTAMPTZ not null
- `event_occurred_time` TIMESTAMPTZ nullable
- `event_received_time` TIMESTAMPTZ nullable
- `affected_user_id` UUID nullable
- `assigned_user_id` UUID nullable
- `assignment_group_id` UUID nullable
- `alert_rule_id` UUID nullable
- `active_duration_seconds` BIGINT not null default 0
- `accuracy` VARCHAR(50) nullable
- `determination` VARCHAR(100) nullable
- `impact` VARCHAR(50) nullable
- `closure_code` VARCHAR(50) nullable
- `closure_reason` TEXT nullable
- `closed_by_user_id` UUID nullable
- `closed_time` TIMESTAMPTZ nullable
- `is_active` BOOLEAN not null default true
- `version_no` INTEGER not null default 1
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

## `case_worknotes`
Fields:
- `id` UUID PK
- `case_id` UUID not null
- `note_text` TEXT not null
- `note_type` VARCHAR(30) not null default `worknote`
- `created_by_user_id` UUID not null
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ nullable
- `is_deleted` BOOLEAN not null default false

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Business Rules

## Case Creation
Required:
- short_description
- state
- priority
- severity
- opened_by_user_id
- opened_time

Optional:
- description
- affected_user_id
- assigned_user_id
- assignment_group_id
- alert_rule_id
- event_occurred_time
- event_received_time
- accuracy
- determination
- impact

Behavior:
- generate `id`
- generate unique `case_number`
- initialize `version_no=1`
- set `is_active=true` unless state is immediately terminal
- publish `case.created`

## Case Number Format
Use a stable business key format, for example `INC-000001`.
Implementation may use a sequence table or dedicated key generator.
Do not derive the business key from UUIDs.

## State Model
Minimum states:
- `new`
- `triage`
- `in_progress`
- `pending`
- `resolved`
- `closed`

Terminal states:
- `resolved`
- `closed`

Rules:
- cannot move from `closed` back to non-terminal without explicit reopen feature; reopen is out of scope for v1
- closing a case requires closure fields
- `is_active=false` when terminal
- `closed_time` and `closed_by_user_id` required when state becomes `closed`

## Assignment Rules
- `assigned_user_id` may be null
- `assignment_group_id` may be null
- assignment updates publish `case.assigned`
- service may store IDs without synchronously validating existence on every write if platform chooses eventual validation, but preferred behavior is gRPC validation through Assignment & Reference Service on command paths

## Active Duration
- maintain `active_duration_seconds`
- for v1, count elapsed time in active states (`new`, `triage`, `in_progress`)
- do not count `pending`, `resolved`, `closed`
- update on reads or writes using deterministic logic, or via periodic updater if chosen
- document exact algorithm in code comments and tests

## Worknotes
- append-only creation model
- note edits optional; if supported, publish update events
- soft delete allowed only if explicitly approved; default behavior is immutable notes after creation
- creating a worknote publishes `worknote.created`

## Closure Rules
To close:
- set `state=closed`
- require `closure_code`
- require `closure_reason`
- require actor as `closed_by_user_id`
- set `closed_time`
- set `is_active=false`
- publish `case.closed`

# REST API Surface
Public UI access should come through API Gateway, but this service may expose internal REST or admin routes if needed.

## Required Internal Capabilities
- create case
- get case by id
- search/list cases
- patch case
- assign case
- close case
- create worknote
- list worknotes

# gRPC API Contracts

## Commands
- `CreateCase`
- `UpdateCase`
- `AssignCase`
- `CloseCase`
- `AddWorknote`

## Queries
- `GetCase`
- `ListCases`
- `ListWorknotes`

Every command must accept actor and request metadata.

# Integration Dependencies
- Assignment & Reference Service for optional validation of `opened_by_user_id`, `affected_user_id`, `assigned_user_id`, `assignment_group_id`
- Alert & Observable Service only for aggregated reads or future denormalized counts, not for core write path
- Audit publication via event bus

# Audit Events To Publish
- `case.created`
- `case.updated`
- `case.state.changed`
- `case.assigned`
- `case.closed`
- `worknote.created`

Include before/after snapshots for mutable fields where feasible.

# Validation Rules
- `short_description` length 1..500
- enums for state/priority/severity/accuracy/impact must be validated. Allowed values: **state** per State Model above; **priority** and **severity** per Shared Enum Conventions (POC) in platform contract (e.g. priority: `P1`, `P2`, `P3`, `P4`; severity: `critical`, `high`, `medium`, `low`). Document accuracy and impact allowed values in this service's API contract if used.
- `event_received_time` should not be before `event_occurred_time` unless explicitly allowed by source corrections
- closure fields required only on close
- reject unknown fields if strict decoding is used

# Concurrency Rules
- use optimistic concurrency through `version_no`
- reject conflicting updates with clear error
- test simultaneous patch behavior

# Search/List Expectations
Minimum filters:
- state
- priority
- severity
- assignment_group_id
- assigned_user_id
- affected_user_id
- opened_time range

Sorting:
- opened_time desc default

# Non-Functional Requirements
- p95 read under 150 ms for single-case fetch
- p95 write under 200 ms excluding downstream validation
- audit event publication with transactional outbox preferred
- high test coverage for state transitions

# Implementation Guidance for Agents
1. Define domain model and allowed enums first.
2. Implement migration files.
3. Implement repository layer with optimistic concurrency.
4. Implement service layer with explicit transition validators.
5. Implement gRPC transport.
6. Implement event outbox and publisher.
7. Add tests for all state transitions and closure edge cases.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). This service uses the full layout with optional REST (internal/admin).

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   ├── case.go
│   └── worknote.go
├── api/
│   ├── api.go
│   ├── case_handlers.go
│   └── worknote_handlers.go
├── service/
│   ├── service.go
│   ├── case.go
│   └── worknote.go
├── repository/
│   └── repository.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── case_service.proto
│   ├── generate_proto.sh
│   └── case_service/
│       ├── case_service.pb.go
│       └── case_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `cmd/main.go`: Wire gRPC server (e.g. 50051) + HTTP server (8080 for health and optional REST), repository, service, handler, api (if REST), event publisher, graceful shutdown.
- `handler/`: gRPC only; embed `Unimplemented*Server`; case.go, worknote.go by resource; delegate to Service; map errors to gRPC status codes.
- `api/`: Optional; same Service layer; case_handlers.go, worknote_handlers.go for internal REST.
- `service/`: Business logic; Repository + optional gRPC clients (e.g. Reference for validation); case.go, worknote.go.
- `repository/`: DB access only; optimistic concurrency via `version_no`.
- `proto/case_service.proto`: `go_package = "<module>/proto/case_service"`; define Commands and Queries per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/case-service` |
| Proto | `case_service.proto` (or case.proto); go_package `<module>/proto/case_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health + optional REST) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint `GET /health` (200 when up; optional 503 if DB unavailable); graceful shutdown (drain gRPC and HTTP); gRPC and REST error mapping per platform contract; timeouts on inbound and outbound calls; audit event publication (transactional outbox preferred); optimistic concurrency in repository; no secrets in logs.

## Agent Checklist

- [ ] Set module name in go.mod and all imports; rename proto and set go_package.
- [ ] Run `proto/generate_proto.sh case_service go`.
- [ ] Implement repository with version_no; migrations for cases and case_worknotes.
- [ ] Implement service layer (state/closure rules, transition validators).
- [ ] Implement handler (gRPC) and optional api (REST); map errors to status codes.
- [ ] Wire event publisher; publish after successful commit (outbox preferred).
- [ ] main.go: gRPC + HTTP, server timeouts, graceful shutdown.
- [ ] Add tests (state transitions, closure, optimistic concurrency, handler tests).
- [ ] Update README (structure, gRPC/REST, setup, tests).

# Do Not Do
- Do not store observables in this service.
- Do not perform threat lookup logic here.
- Do not write to audit tables directly.
- Do not let API clients assign arbitrary case numbers.
- Do not skip optimistic concurrency on patch operations.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE cases (
  id UUID PRIMARY KEY,
  case_number VARCHAR(50) NOT NULL UNIQUE,
  short_description VARCHAR(500) NOT NULL,
  description TEXT,
  state VARCHAR(50) NOT NULL,
  priority VARCHAR(20) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  opened_by_user_id UUID NOT NULL,
  opened_time TIMESTAMPTZ NOT NULL,
  event_occurred_time TIMESTAMPTZ,
  event_received_time TIMESTAMPTZ,
  affected_user_id UUID,
  assigned_user_id UUID,
  assignment_group_id UUID,
  alert_rule_id UUID,
  active_duration_seconds BIGINT NOT NULL DEFAULT 0,
  accuracy VARCHAR(50),
  determination VARCHAR(100),
  impact VARCHAR(50),
  closure_code VARCHAR(50),
  closure_reason TEXT,
  closed_by_user_id UUID,
  closed_time TIMESTAMPTZ,
  is_active BOOLEAN NOT NULL DEFAULT true,
  version_no INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE case_worknotes (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL REFERENCES cases(id),
  note_text TEXT NOT NULL,
  note_type VARCHAR(30) NOT NULL DEFAULT 'worknote',
  created_by_user_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT false
);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package case_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/case_service";

service CaseService {
  rpc CreateCase(CreateCaseRequest) returns (CreateCaseResponse);
  rpc UpdateCase(UpdateCaseRequest) returns (UpdateCaseResponse);
  rpc AssignCase(AssignCaseRequest) returns (AssignCaseResponse);
  rpc CloseCase(CloseCaseRequest) returns (CloseCaseResponse);
  rpc AddWorknote(AddWorknoteRequest) returns (AddWorknoteResponse);
  rpc GetCase(GetCaseRequest) returns (GetCaseResponse);
  rpc ListCases(ListCasesRequest) returns (ListCasesResponse);
  rpc ListWorknotes(ListWorknotesRequest) returns (ListWorknotesResponse);
}

message CreateCaseRequest {
  string short_description = 1;
  string description = 2;
  string state = 3;
  string priority = 4;
  string severity = 5;
  string opened_by_user_id = 6;
  google.protobuf.Timestamp opened_time = 7;
  google.protobuf.Timestamp event_occurred_time = 8;
  google.protobuf.Timestamp event_received_time = 9;
  string affected_user_id = 10;
  string assigned_user_id = 11;
  string assignment_group_id = 12;
  string alert_rule_id = 13;
  string accuracy = 14;
  string determination = 15;
  string impact = 16;
}

message CreateCaseResponse {
  string id = 1;
  string case_number = 2;
}

message UpdateCaseRequest {
  string id = 1;
  optional string short_description = 2;
  optional string description = 3;
  optional string state = 4;
  optional string priority = 5;
  optional string severity = 6;
  optional string affected_user_id = 7;
  optional string assigned_user_id = 8;
  optional string assignment_group_id = 9;
  optional string accuracy = 10;
  optional string determination = 11;
  optional string impact = 12;
  int32 version_no = 13;
}

message UpdateCaseResponse {
  Case case = 1;
}

message AssignCaseRequest {
  string case_id = 1;
  string assigned_user_id = 2;
  string assignment_group_id = 3;
  int32 version_no = 4;
}

message AssignCaseResponse {
  Case case = 1;
}

message CloseCaseRequest {
  string case_id = 1;
  string closure_code = 2;
  string closure_reason = 3;
  int32 version_no = 4;
}

message CloseCaseResponse {
  Case case = 1;
}

message AddWorknoteRequest {
  string case_id = 1;
  string note_text = 2;
  string note_type = 3;
  string created_by_user_id = 4;
}

message AddWorknoteResponse {
  Worknote worknote = 1;
}

message GetCaseRequest {
  string case_id = 1;
}

message GetCaseResponse {
  Case case = 1;
}

message ListCasesRequest {
  int32 page_size = 1;
  string page_token = 2;
  optional string state = 3;
  optional string priority = 4;
  optional string severity = 5;
  optional string assignment_group_id = 6;
  optional string assigned_user_id = 7;
  optional string affected_user_id = 8;
  optional google.protobuf.Timestamp opened_time_after = 9;
  optional google.protobuf.Timestamp opened_time_before = 10;
}

message ListCasesResponse {
  repeated Case items = 1;
  string next_page_token = 2;
}

message ListWorknotesRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListWorknotesResponse {
  repeated Worknote items = 1;
  string next_page_token = 2;
}

message Case {
  string id = 1;
  string case_number = 2;
  string short_description = 3;
  string description = 4;
  string state = 5;
  string priority = 6;
  string severity = 7;
  string opened_by_user_id = 8;
  google.protobuf.Timestamp opened_time = 9;
  google.protobuf.Timestamp event_occurred_time = 10;
  google.protobuf.Timestamp event_received_time = 11;
  string affected_user_id = 12;
  string assigned_user_id = 13;
  string assignment_group_id = 14;
  string alert_rule_id = 15;
  int64 active_duration_seconds = 16;
  string accuracy = 17;
  string determination = 18;
  string impact = 19;
  string closure_code = 20;
  string closure_reason = 21;
  string closed_by_user_id = 22;
  google.protobuf.Timestamp closed_time = 23;
  bool is_active = 24;
  int32 version_no = 25;
  google.protobuf.Timestamp created_at = 26;
  google.protobuf.Timestamp updated_at = 27;
}

message Worknote {
  string id = 1;
  string case_id = 2;
  string note_text = 3;
  string note_type = 4;
  string created_by_user_id = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
  bool is_deleted = 8;
}
```
