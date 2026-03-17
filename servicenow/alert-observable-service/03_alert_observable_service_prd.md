# Alert & Observable Service PRD
Version: 1.0  
Service name: alert-observable-service  
Language: Golang  
Primary responsibility: source of truth for alerts, observables, child observables, case-observable links, and similar incident relationships

## Mission
Manage the security detection artifacts associated with incidents. This includes source alerts, canonical observables, case-observable relationships, child observable relationships, observable tracking metadata, and automatic linkage of similar incidents that share one or more observables.


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
- alert rule storage
- alert storage
- observable storage
- linking observables to cases
- tracking observables within a case
- child observable relationships
- similar incident computation and storage
- publish audit events for all writes

# Out of Scope
- owning case lifecycle
- worknotes
- threat lookup persistence
- enrichment payload persistence
- attachment persistence
- audit event storage

# Owned Data Model

## `alert_rules`
- `id` UUID PK
- `rule_name` VARCHAR(255) not null
- `rule_type` VARCHAR(100) nullable
- `source_system` VARCHAR(100) nullable
- `external_rule_key` VARCHAR(255) nullable
- `description` TEXT nullable
- `is_active` BOOLEAN not null default true
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

## `alerts`
- `id` UUID PK
- `case_id` UUID not null
- `alert_rule_id` UUID nullable
- `source_system` VARCHAR(100) not null
- `source_alert_id` VARCHAR(255) nullable
- `title` VARCHAR(500) nullable
- `description` TEXT nullable
- `event_occurred_time` TIMESTAMPTZ nullable
- `event_received_time` TIMESTAMPTZ nullable
- `severity` VARCHAR(20) nullable
- `raw_payload` JSONB nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

## `observables`
- `id` UUID PK
- `observable_type` VARCHAR(50) not null
- `observable_value` VARCHAR(1000) not null
- `normalized_value` VARCHAR(1000) nullable
- `first_seen_time` TIMESTAMPTZ nullable
- `last_seen_time` TIMESTAMPTZ nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Recommended uniqueness:
- unique on (`observable_type`, `normalized_value`) when normalization exists

## `case_observables`
- `id` UUID PK
- `case_id` UUID not null
- `observable_id` UUID not null
- `role_in_case` VARCHAR(50) nullable
- `tracking_status` VARCHAR(50) nullable
- `is_primary` BOOLEAN not null default false
- `accuracy` VARCHAR(50) nullable
- `determination` VARCHAR(100) nullable
- `impact` VARCHAR(50) nullable
- `added_by_user_id` UUID nullable
- `added_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Recommended uniqueness:
- unique on (`case_id`, `observable_id`)

## `child_observables`
- `id` UUID PK
- `parent_observable_id` UUID not null
- `child_observable_id` UUID not null
- `relationship_type` VARCHAR(100) not null
- `relationship_direction` VARCHAR(20) nullable
- `confidence` NUMERIC(5,2) nullable
- `source_name` VARCHAR(100) nullable
- `source_record_id` VARCHAR(255) nullable
- `metadata` JSONB nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Recommended uniqueness:
- unique on (`parent_observable_id`, `child_observable_id`, `relationship_type`)

## `similar_security_incidents`
- `id` UUID PK
- `case_id` UUID not null
- `similar_case_id` UUID not null
- `match_reason` VARCHAR(100) not null default `shared_observable`
- `shared_observable_count` INTEGER not null default 1
- `shared_observable_ids` JSONB not null
- `shared_observable_values` JSONB nullable
- `similarity_score` NUMERIC(10,2) nullable
- `last_computed_at` TIMESTAMPTZ not null
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Recommended uniqueness:
- unique on (`case_id`, `similar_case_id`)

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Business Rules

## Alert Creation
- alerts belong to a case
- `source_system` required
- raw source payload allowed in `raw_payload`
- creating an alert publishes `alert.created`

## Observable Canonicalization
Normalize observable values by type where possible, examples:
- IP addresses normalized to canonical textual form
- domains lowercased and trimmed
- URLs normalized consistently without losing query/path semantics unless defined
- hashes lowercased
- emails lowercased for address portion

Store:
- original `observable_value`
- canonical `normalized_value`

Do not overwrite the original value with the normalized one.

## Linking Observable to Case
Command behavior:
1. validate case reference shape
2. create or resolve canonical observable
3. create `case_observables` row if not present
4. compute/update similar incidents
5. publish audit events

This flow must be idempotent when the same observable is linked repeatedly with the same natural key.

## Observable Tracking
Allowed example values for `tracking_status`:
- `new`
- `under_review`
- `enriched`
- `dismissed`
- `confirmed`

## Child Observable Relationships
Use this table for observable-to-observable relationships such as:
- domain resolves_to IP
- URL contains domain
- file drops hash
- process spawns process
- email contains attachment hash

Child relationships are independent of case membership.

## Similar Incident Computation
Requirement:
- every ticket that has at least one same observable must appear as a similar incident

Algorithm baseline:
1. when a `case_observable` is created or removed, find all other cases linked to that observable
2. aggregate shared observables by other case
3. upsert `similar_security_incidents` rows
4. store count and IDs of shared observables
5. optionally store display-friendly values
6. update both directions (`A->B` and `B->A`) for query simplicity

Important:
- this is not fuzzy matching in v1
- similarity is based on exact shared observable identity
- similarity score may equal shared observable count or a normalized function, but must be deterministic and documented

# gRPC API Contracts

## Commands
- `CreateAlertRule`
- `CreateAlert`
- `CreateOrGetObservable`
- `LinkObservableToCase`
- `UpdateCaseObservable`
- `CreateChildObservableRelation`
- `RecomputeSimilarIncidentsForCase`

## Queries
- `GetAlert`
- `ListCaseAlerts`
- `GetObservable`
- `ListCaseObservables`
- `ListChildObservables`
- `ListSimilarIncidents`
- `FindCasesByObservable`

# Integration Dependencies
- Case Service for validating case existence if strong validation is enabled
- Enrichment & Threat Lookup Service for read aggregation only, not required for core writes
- Audit via pub/sub

# Audit Events To Publish
- `alert.created`
- `observable.created`
- `observable.linked.to_case`
- `observable.updated`
- `child_observable.created`
- `similar_incident.linked`

# Validation Rules
- `observable_type` is required and for v1 MUST be one of: `ip`, `domain`, `hash`, `url`, `email`, `file` (see Shared Enum Conventions in the platform contract for updates).
- prevent parent and child from being the same observable unless explicitly supported; default reject self-links
- prevent duplicate case-observable links
- prevent duplicate child links
- `case_id` and `similar_case_id` cannot be equal in similar incident table

# Search and Query Requirements
- list alerts by case
- list observables by case
- filter observables by type and tracking status
- list similar incidents by case
- list child observables by parent observable
- find all cases containing an observable

# Performance Expectations
- linking an observable to a case should remain efficient under many historical cases
- add appropriate indexes:
  - `case_observables(case_id)`
  - `case_observables(observable_id)`
  - `similar_security_incidents(case_id)`
  - `child_observables(parent_observable_id)`
- tests must cover similar incident recomputation cost on realistic fixture volumes

# Implementation Guidance for Agents
1. Implement normalization library per observable type.
2. Implement canonical create-or-get logic with uniqueness controls.
3. Implement case linking workflow with transaction boundaries.
4. Implement similar incident updater as domain logic or worker inside this service.
5. Add full tests for duplicate links, normalization, and similar-incident updates.
6. Publish audit events only after successful persistence.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). Full layout with gRPC server; add `api/` only if REST is needed for this service.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   ├── alert.go
│   ├── observable.go
│   ├── case_observable.go
│   └── similar_incident.go
├── api/
│   ├── api.go
│   └── (resource)_handlers.go   # optional, if REST exposed
├── service/
│   ├── service.go
│   ├── alert.go
│   ├── observable.go
│   ├── case_observable.go
│   └── similar_incident.go
├── repository/
│   └── repository.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── alert_observable_service.proto
│   ├── generate_proto.sh
│   └── alert_observable_service/
│       ├── alert_observable_service.pb.go
│       └── alert_observable_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `handler/`: Group by resource (alert, observable, case_observable, similar_incident); delegate to Service; map errors to gRPC status codes.
- `service/`: Business logic; normalization, create-or-get observable, link-to-case (idempotent), similar-incident recomputation; call Repository; publish audit events after persist.
- `repository/`: alert_rules, alerts, observables, case_observables, child_observables, similar_security_incidents; indexes per Performance Expectations.
- `proto/alert_observable_service.proto`: `go_package = "<module>/proto/alert_observable_service"`; define all Commands and Queries per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/alert-observable-service` |
| Proto | `alert_observable_service.proto`; go_package `<module>/proto/alert_observable_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health + optional REST) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint; graceful shutdown; gRPC (and REST if present) error mapping; timeouts; idempotent link observable to case; similar-incident recomputation on link/unlink; audit events after successful persistence; no secrets in logs.

## Agent Checklist

- [ ] Set module name and proto go_package; run `proto/generate_proto.sh alert_observable_service go`.
- [ ] Implement repository with migrations and indexes (case_observables, observables, similar_security_incidents, child_observables).
- [ ] Implement normalization and create-or-get observable; case linking with transaction; similar-incident updater.
- [ ] Implement handler (and optional api) per resource; wire event publisher.
- [ ] main.go: gRPC + HTTP, timeouts, graceful shutdown.
- [ ] Tests: duplicate links, normalization, similar-incident updates, handler tests.
- [ ] Update README.

# Do Not Do
- Do not trigger external enrichment tools.
- Do not store threat lookup results here.
- Do not assume fuzzy similarity in v1.
- Do not duplicate case lifecycle fields here except foreign references.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE alert_rules (
  id UUID PRIMARY KEY,
  rule_name VARCHAR(255) NOT NULL,
  rule_type VARCHAR(100),
  source_system VARCHAR(100),
  external_rule_key VARCHAR(255),
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE observables (
  id UUID PRIMARY KEY,
  observable_type VARCHAR(50) NOT NULL,
  observable_value VARCHAR(1000) NOT NULL,
  normalized_value VARCHAR(1000),
  first_seen_time TIMESTAMPTZ,
  last_seen_time TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (observable_type, normalized_value)
);

CREATE TABLE alerts (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  alert_rule_id UUID REFERENCES alert_rules(id),
  source_system VARCHAR(100) NOT NULL,
  source_alert_id VARCHAR(255),
  title VARCHAR(500),
  description TEXT,
  event_occurred_time TIMESTAMPTZ,
  event_received_time TIMESTAMPTZ,
  severity VARCHAR(20),
  raw_payload JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE case_observables (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  observable_id UUID NOT NULL REFERENCES observables(id),
  role_in_case VARCHAR(50),
  tracking_status VARCHAR(50),
  is_primary BOOLEAN NOT NULL DEFAULT false,
  accuracy VARCHAR(50),
  determination VARCHAR(100),
  impact VARCHAR(50),
  added_by_user_id UUID,
  added_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (case_id, observable_id)
);

CREATE TABLE child_observables (
  id UUID PRIMARY KEY,
  parent_observable_id UUID NOT NULL REFERENCES observables(id),
  child_observable_id UUID NOT NULL REFERENCES observables(id),
  relationship_type VARCHAR(100) NOT NULL,
  relationship_direction VARCHAR(20),
  confidence NUMERIC(5,2),
  source_name VARCHAR(100),
  source_record_id VARCHAR(255),
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (parent_observable_id, child_observable_id, relationship_type)
);

CREATE TABLE similar_security_incidents (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  similar_case_id UUID NOT NULL,
  match_reason VARCHAR(100) NOT NULL DEFAULT 'shared_observable',
  shared_observable_count INTEGER NOT NULL DEFAULT 1,
  shared_observable_ids JSONB NOT NULL,
  shared_observable_values JSONB,
  similarity_score NUMERIC(10,2),
  last_computed_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (case_id, similar_case_id)
);

CREATE INDEX idx_case_observables_case_id ON case_observables(case_id);
CREATE INDEX idx_case_observables_observable_id ON case_observables(observable_id);
CREATE INDEX idx_similar_security_incidents_case_id ON similar_security_incidents(case_id);
CREATE INDEX idx_child_observables_parent_observable_id ON child_observables(parent_observable_id);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package alert_observable_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/alert_observable_service";

service AlertObservableService {
  rpc CreateAlertRule(CreateAlertRuleRequest) returns (CreateAlertRuleResponse);
  rpc CreateAlert(CreateAlertRequest) returns (CreateAlertResponse);
  rpc CreateOrGetObservable(CreateOrGetObservableRequest) returns (CreateOrGetObservableResponse);
  rpc LinkObservableToCase(LinkObservableToCaseRequest) returns (LinkObservableToCaseResponse);
  rpc UpdateCaseObservable(UpdateCaseObservableRequest) returns (UpdateCaseObservableResponse);
  rpc CreateChildObservableRelation(CreateChildObservableRelationRequest) returns (CreateChildObservableRelationResponse);
  rpc RecomputeSimilarIncidentsForCase(RecomputeSimilarIncidentsForCaseRequest) returns (RecomputeSimilarIncidentsForCaseResponse);
  rpc GetAlert(GetAlertRequest) returns (GetAlertResponse);
  rpc ListCaseAlerts(ListCaseAlertsRequest) returns (ListCaseAlertsResponse);
  rpc GetObservable(GetObservableRequest) returns (GetObservableResponse);
  rpc ListCaseObservables(ListCaseObservablesRequest) returns (ListCaseObservablesResponse);
  rpc ListChildObservables(ListChildObservablesRequest) returns (ListChildObservablesResponse);
  rpc ListSimilarIncidents(ListSimilarIncidentsRequest) returns (ListSimilarIncidentsResponse);
  rpc FindCasesByObservable(FindCasesByObservableRequest) returns (FindCasesByObservableResponse);
}

message CreateAlertRuleRequest {
  string rule_name = 1;
  string rule_type = 2;
  string source_system = 3;
  string external_rule_key = 4;
  string description = 5;
  bool is_active = 6;
}

message CreateAlertRuleResponse {
  string id = 1;
}

message CreateAlertRequest {
  string case_id = 1;
  string alert_rule_id = 2;
  string source_system = 3;
  string source_alert_id = 4;
  string title = 5;
  string description = 6;
  google.protobuf.Timestamp event_occurred_time = 7;
  google.protobuf.Timestamp event_received_time = 8;
  string severity = 9;
  string raw_payload_json = 10;
}

message CreateAlertResponse {
  string id = 1;
}

message CreateOrGetObservableRequest {
  string observable_type = 1;
  string observable_value = 2;
  string normalized_value = 3;
  google.protobuf.Timestamp first_seen_time = 4;
  google.protobuf.Timestamp last_seen_time = 5;
}

message CreateOrGetObservableResponse {
  Observable observable = 1;
  bool created = 2;
}

message LinkObservableToCaseRequest {
  string case_id = 1;
  string observable_type = 2;
  string observable_value = 3;
  string normalized_value = 4;
  string role_in_case = 5;
  string tracking_status = 6;
  bool is_primary = 7;
  string added_by_user_id = 8;
}

message LinkObservableToCaseResponse {
  CaseObservable case_observable = 1;
}

message UpdateCaseObservableRequest {
  string case_observable_id = 1;
  optional string tracking_status = 2;
  optional string accuracy = 3;
  optional string determination = 4;
  optional string impact = 5;
}

message UpdateCaseObservableResponse {
  CaseObservable case_observable = 1;
}

message CreateChildObservableRelationRequest {
  string parent_observable_id = 1;
  string child_observable_id = 2;
  string relationship_type = 3;
  string relationship_direction = 4;
  double confidence = 5;
  string source_name = 6;
  string source_record_id = 7;
  string metadata_json = 8;
}

message CreateChildObservableRelationResponse {
  ChildObservable relation = 1;
}

message RecomputeSimilarIncidentsForCaseRequest {
  string case_id = 1;
}

message RecomputeSimilarIncidentsForCaseResponse {}

message GetAlertRequest {
  string alert_id = 1;
}

message GetAlertResponse {
  Alert alert = 1;
}

message ListCaseAlertsRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListCaseAlertsResponse {
  repeated Alert items = 1;
  string next_page_token = 2;
}

message GetObservableRequest {
  string observable_id = 1;
}

message GetObservableResponse {
  Observable observable = 1;
}

message ListCaseObservablesRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional string observable_type = 4;
  optional string tracking_status = 5;
}

message ListCaseObservablesResponse {
  repeated CaseObservable items = 1;
  string next_page_token = 2;
}

message ListChildObservablesRequest {
  string parent_observable_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListChildObservablesResponse {
  repeated ChildObservable items = 1;
  string next_page_token = 2;
}

message ListSimilarIncidentsRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListSimilarIncidentsResponse {
  repeated SimilarIncident items = 1;
  string next_page_token = 2;
}

message FindCasesByObservableRequest {
  string observable_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message FindCasesByObservableResponse {
  repeated string case_ids = 1;
  string next_page_token = 2;
}

message AlertRule {
  string id = 1;
  string rule_name = 2;
  string rule_type = 3;
  string source_system = 4;
  string external_rule_key = 5;
  string description = 6;
  bool is_active = 7;
  google.protobuf.Timestamp created_at = 8;
  google.protobuf.Timestamp updated_at = 9;
}

message Alert {
  string id = 1;
  string case_id = 2;
  string alert_rule_id = 3;
  string source_system = 4;
  string source_alert_id = 5;
  string title = 6;
  string description = 7;
  google.protobuf.Timestamp event_occurred_time = 8;
  google.protobuf.Timestamp event_received_time = 9;
  string severity = 10;
  string raw_payload_json = 11;
  google.protobuf.Timestamp created_at = 12;
  google.protobuf.Timestamp updated_at = 13;
}

message Observable {
  string id = 1;
  string observable_type = 2;
  string observable_value = 3;
  string normalized_value = 4;
  google.protobuf.Timestamp first_seen_time = 5;
  google.protobuf.Timestamp last_seen_time = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

message CaseObservable {
  string id = 1;
  string case_id = 2;
  string observable_id = 3;
  string role_in_case = 4;
  string tracking_status = 5;
  bool is_primary = 6;
  string accuracy = 7;
  string determination = 8;
  string impact = 9;
  string added_by_user_id = 10;
  google.protobuf.Timestamp added_at = 11;
  google.protobuf.Timestamp updated_at = 12;
}

message ChildObservable {
  string id = 1;
  string parent_observable_id = 2;
  string child_observable_id = 3;
  string relationship_type = 4;
  string relationship_direction = 5;
  double confidence = 6;
  string source_name = 7;
  string source_record_id = 8;
  string metadata_json = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}

message SimilarIncident {
  string id = 1;
  string case_id = 2;
  string similar_case_id = 3;
  string match_reason = 4;
  int32 shared_observable_count = 5;
  string shared_observable_ids_json = 6;
  string shared_observable_values_json = 7;
  double similarity_score = 8;
  google.protobuf.Timestamp last_computed_at = 9;
  google.protobuf.Timestamp created_at = 10;
  google.protobuf.Timestamp updated_at = 11;
}
```
