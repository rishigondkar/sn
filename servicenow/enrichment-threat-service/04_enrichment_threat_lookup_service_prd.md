# Enrichment & Threat Lookup Service PRD
Version: 1.0  
Service name: enrichment-threat-service  
Language: Golang  
Primary responsibility: passive persistence and retrieval of enrichment results and threat lookup results

## Mission
Provide a stable service for storing, updating, and querying enrichment and threat lookup data linked to observables and optionally cases. This service does not automatically call external tools. It only accepts data through APIs and returns stored results.


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
- store enrichment results
- store threat lookup results
- query enrichment by case or observable
- query threat lookups by case or observable
- idempotent upsert behavior for external integrations
- publish audit events for upserts

# Out of Scope
- running external enrichment jobs automatically
- orchestrating tool execution
- owning observables
- owning cases
- child observable management
- attachments
- audit event storage

# Owned Data Model

## `enrichment_results`
- `id` UUID PK
- `case_id` UUID nullable
- `observable_id` UUID nullable
- `enrichment_type` VARCHAR(100) not null
- `source_name` VARCHAR(100) not null
- `source_record_id` VARCHAR(255) nullable
- `status` VARCHAR(30) not null
- `summary` VARCHAR(1000) nullable
- `result_data` JSONB not null
- `score` NUMERIC(10,2) nullable
- `confidence` NUMERIC(5,2) nullable
- `requested_at` TIMESTAMPTZ nullable
- `received_at` TIMESTAMPTZ not null
- `expires_at` TIMESTAMPTZ nullable
- `last_updated_by` VARCHAR(255) nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Constraint:
- at least one of `case_id` or `observable_id` must be non-null

## `threat_lookup_results`
- `id` UUID PK
- `case_id` UUID nullable
- `observable_id` UUID not null
- `lookup_type` VARCHAR(100) not null
- `source_name` VARCHAR(100) not null
- `source_record_id` VARCHAR(255) nullable
- `verdict` VARCHAR(50) nullable
- `risk_score` NUMERIC(10,2) nullable
- `confidence_score` NUMERIC(5,2) nullable
- `tags` JSONB nullable
- `matched_indicators` JSONB nullable
- `summary` VARCHAR(1000) nullable
- `result_data` JSONB not null
- `queried_at` TIMESTAMPTZ nullable
- `received_at` TIMESTAMPTZ not null
- `expires_at` TIMESTAMPTZ nullable
- `created_at` TIMESTAMPTZ not null
- `updated_at` TIMESTAMPTZ not null

Executable PostgreSQL DDL for these tables: see **Appendix A: Table schemas (DDL)**.

# Business Rules

## Passive Only
This service must not initiate enrichment or threat lookup calls on its own.
No cron-triggered lookups.
No event-triggered outbound lookups.
No hidden side effects.

## Upsert Semantics
Integrations may retry. The service must support idempotent upserts.
Recommended natural key for dedupe:
- observable_id or case_id
- type
- source_name
- source_record_id if available
- queried_at/received_at depending on source semantics

If a source sends repeated updates for the same logical record, update the existing row rather than creating duplicates.

## Expiry Semantics
- `expires_at` indicates staleness/TTL if provided by source or policy
- service does not auto-refresh stale records
- query APIs may optionally filter active/non-expired records

## Distinction Between Enrichment and Threat Lookup
- enrichment is general contextual data
- threat lookup is threat-intelligence-oriented with verdicts, risk, tags, indicator matches
- do not merge these tables in v1

# REST and Integration API Surface
These routes are especially important for external clients that post results.

## Required Routes
- `POST /api/v1/enrichment-results`
- `PUT /api/v1/enrichment-results/{id}`
- `GET /api/v1/cases/{caseId}/enrichment-results`
- `GET /api/v1/observables/{observableId}/enrichment-results`
- `POST /api/v1/threat-lookups`
- `PUT /api/v1/threat-lookups/{id}`
- `GET /api/v1/cases/{caseId}/threat-lookups`
- `GET /api/v1/observables/{observableId}/threat-lookups`

# gRPC API Contracts

## Commands
- `UpsertEnrichmentResult`
- `UpsertThreatLookupResult`

## Queries
- `ListEnrichmentResultsByCase`
- `ListEnrichmentResultsByObservable`
- `ListThreatLookupResultsByCase`
- `ListThreatLookupResultsByObservable`
- `GetThreatLookupSummaryByObservable`

# Integration Dependencies
- Alert & Observable Service for optional validation of observable existence
- Case Service for optional validation of case existence
- Audit via event bus

Validation may be strict or eventual, but behavior must be documented and consistent.

# Audit Events To Publish
- `enrichment_result.upserted`
- `threat_lookup_result.upserted`

# Validation Rules
- `result_data` required and must be valid JSON object/array according to source schema requirements
- `observable_id` required for threat lookups
- for enrichment, at least one of `case_id` or `observable_id` required
- enum validation for `status` and `verdict` where defined
- reject oversized payloads beyond configured limits

# Query Behavior
- list by case
- list by observable
- optional filters: source_name, type, verdict, active_only
- sort newest first by `received_at`

# Non-Functional Requirements
- reliable ingestion under retry
- support payload sizes large enough for lookup results within configured guardrails
- strong indexing on `observable_id`, `case_id`, `source_name`, and timestamps

# Implementation Guidance for Agents
1. Implement ingestion DTOs and dedupe strategy first.
2. Define repository upsert methods with explicit conflict targets.
3. Preserve raw source payload in `result_data`.
4. Add tests for duplicate retries and partial updates.
5. Keep source-specific parsing out of the core service unless standardized adapters are explicitly added later.

# Go Project Structure

Follow the canonical template: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). This service **requires both gRPC and REST**: gRPC for gateway and internal callers; REST for external clients posting enrichment and threat-lookup results.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── handler/
│   ├── handler.go
│   ├── handler_test.go
│   ├── enrichment.go
│   └── threat_lookup.go
├── api/
│   ├── api.go
│   ├── enrichment_handlers.go
│   └── threat_lookup_handlers.go
├── service/
│   ├── service.go
│   ├── enrichment.go
│   └── threat_lookup.go
├── repository/
│   └── repository.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── proto/
│   ├── enrichment_threat_service.proto
│   ├── generate_proto.sh
│   └── enrichment_threat_service/
│       ├── enrichment_threat_service.pb.go
│       └── enrichment_threat_service_grpc.pb.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- `handler/`: gRPC for UpsertEnrichmentResult, UpsertThreatLookupResult and all Query RPCs; delegate to Service.
- `api/`: REST for POST/PUT/GET enrichment-results and threat-lookups per "REST and Integration API Surface"; same Service layer; use shared REST error model.
- `service/`: Idempotent upsert logic; dedupe by natural key; publish audit events on upsert.
- `repository/`: enrichment_results, threat_lookup_results; upsert with conflict targets; indexes on observable_id, case_id, source_name, timestamps.
- `proto/enrichment_threat_service.proto`: `go_package = "<module>/proto/enrichment_threat_service"`; Commands and Queries per gRPC API Contracts above. Full proto definition including request/response messages: see **Appendix B: Proto contract**.

## Module, Ports, Proto, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/enrichment-threat-service` |
| Proto | `enrichment_threat_service.proto`; go_package `<module>/proto/enrichment_threat_service` |
| gRPC port | **50051** |
| HTTP port | **8080** (health + REST for ingestion and queries) |
| Go version | 1.22+ (platform contract); 1.23 in go.mod if using template |

**P0:** Health endpoint; graceful shutdown; gRPC and REST error mapping; timeouts; idempotent upserts; audit events for upserts; no secrets in logs; reject oversized payloads per config.

## Agent Checklist

- [ ] Set module name and proto go_package; run proto generation.
- [ ] Implement repository upsert with dedupe/conflict targets; migrations; indexes.
- [ ] Implement service upsert logic; event publisher for enrichment_result.upserted, threat_lookup_result.upserted.
- [ ] Implement handler (gRPC) and api (REST) for enrichment and threat-lookup; register REST routes per PRD.
- [ ] main.go: gRPC + HTTP, timeouts, graceful shutdown.
- [ ] Tests: duplicate retries, upsert idempotency, handler and API tests.
- [ ] Update README.

# Do Not Do
- Do not auto-trigger external tools.
- Do not normalize away important raw source details.
- Do not put observables or cases in this service's database beyond foreign reference IDs.
- Do not conflate enrichment and threat lookup records.

---

# Appendix A: Table schemas (DDL)

```sql
CREATE TABLE enrichment_results (
  id UUID PRIMARY KEY,
  case_id UUID,
  observable_id UUID,
  enrichment_type VARCHAR(100) NOT NULL,
  source_name VARCHAR(100) NOT NULL,
  source_record_id VARCHAR(255),
  status VARCHAR(30) NOT NULL,
  summary VARCHAR(1000),
  result_data JSONB NOT NULL,
  score NUMERIC(10,2),
  confidence NUMERIC(5,2),
  requested_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ,
  last_updated_by VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT enrichment_results_case_or_observable CHECK (case_id IS NOT NULL OR observable_id IS NOT NULL)
);

CREATE TABLE threat_lookup_results (
  id UUID PRIMARY KEY,
  case_id UUID,
  observable_id UUID NOT NULL,
  lookup_type VARCHAR(100) NOT NULL,
  source_name VARCHAR(100) NOT NULL,
  source_record_id VARCHAR(255),
  verdict VARCHAR(50),
  risk_score NUMERIC(10,2),
  confidence_score NUMERIC(5,2),
  tags JSONB,
  matched_indicators JSONB,
  summary VARCHAR(1000),
  result_data JSONB NOT NULL,
  queried_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_enrichment_results_case_id ON enrichment_results(case_id);
CREATE INDEX idx_enrichment_results_observable_id ON enrichment_results(observable_id);
CREATE INDEX idx_enrichment_results_source_name ON enrichment_results(source_name);
CREATE INDEX idx_enrichment_results_received_at ON enrichment_results(received_at);

CREATE INDEX idx_threat_lookup_results_case_id ON threat_lookup_results(case_id);
CREATE INDEX idx_threat_lookup_results_observable_id ON threat_lookup_results(observable_id);
CREATE INDEX idx_threat_lookup_results_source_name ON threat_lookup_results(source_name);
CREATE INDEX idx_threat_lookup_results_received_at ON threat_lookup_results(received_at);
```

---

# Appendix B: Proto contract

Actor and request/correlation ids are passed via gRPC metadata per platform contract; they are not repeated in request messages below.

```protobuf
syntax = "proto3";

package enrichment_threat_service;

import "google/protobuf/timestamp.proto";

option go_package = "<module>/proto/enrichment_threat_service";

service EnrichmentThreatService {
  rpc UpsertEnrichmentResult(UpsertEnrichmentResultRequest) returns (UpsertEnrichmentResultResponse);
  rpc UpsertThreatLookupResult(UpsertThreatLookupResultRequest) returns (UpsertThreatLookupResultResponse);
  rpc ListEnrichmentResultsByCase(ListEnrichmentResultsByCaseRequest) returns (ListEnrichmentResultsByCaseResponse);
  rpc ListEnrichmentResultsByObservable(ListEnrichmentResultsByObservableRequest) returns (ListEnrichmentResultsByObservableResponse);
  rpc ListThreatLookupResultsByCase(ListThreatLookupResultsByCaseRequest) returns (ListThreatLookupResultsByCaseResponse);
  rpc ListThreatLookupResultsByObservable(ListThreatLookupResultsByObservableRequest) returns (ListThreatLookupResultsByObservableResponse);
  rpc GetThreatLookupSummaryByObservable(GetThreatLookupSummaryByObservableRequest) returns (GetThreatLookupSummaryByObservableResponse);
}

message UpsertEnrichmentResultRequest {
  string id = 1;
  optional string case_id = 2;
  optional string observable_id = 3;
  string enrichment_type = 4;
  string source_name = 5;
  optional string source_record_id = 6;
  string status = 7;
  optional string summary = 8;
  string result_data_json = 9;
  optional double score = 10;
  optional double confidence = 11;
  optional google.protobuf.Timestamp requested_at = 12;
  google.protobuf.Timestamp received_at = 13;
  optional google.protobuf.Timestamp expires_at = 14;
  optional string last_updated_by = 15;
}

message UpsertEnrichmentResultResponse {
  EnrichmentResult result = 1;
}

message UpsertThreatLookupResultRequest {
  string id = 1;
  optional string case_id = 2;
  string observable_id = 3;
  string lookup_type = 4;
  string source_name = 5;
  optional string source_record_id = 6;
  optional string verdict = 7;
  optional double risk_score = 8;
  optional double confidence_score = 9;
  optional string tags_json = 10;
  optional string matched_indicators_json = 11;
  optional string summary = 12;
  string result_data_json = 13;
  optional google.protobuf.Timestamp queried_at = 14;
  google.protobuf.Timestamp received_at = 15;
  optional google.protobuf.Timestamp expires_at = 16;
}

message UpsertThreatLookupResultResponse {
  ThreatLookupResult result = 1;
}

message ListEnrichmentResultsByCaseRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional string source_name = 4;
  optional string enrichment_type = 5;
  optional bool active_only = 6;
}

message ListEnrichmentResultsByCaseResponse {
  repeated EnrichmentResult items = 1;
  string next_page_token = 2;
}

message ListEnrichmentResultsByObservableRequest {
  string observable_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional string source_name = 4;
  optional string enrichment_type = 5;
  optional bool active_only = 6;
}

message ListEnrichmentResultsByObservableResponse {
  repeated EnrichmentResult items = 1;
  string next_page_token = 2;
}

message ListThreatLookupResultsByCaseRequest {
  string case_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional string source_name = 4;
  optional string lookup_type = 5;
  optional string verdict = 6;
  optional bool active_only = 7;
}

message ListThreatLookupResultsByCaseResponse {
  repeated ThreatLookupResult items = 1;
  string next_page_token = 2;
}

message ListThreatLookupResultsByObservableRequest {
  string observable_id = 1;
  int32 page_size = 2;
  string page_token = 3;
  optional string source_name = 4;
  optional string lookup_type = 5;
  optional string verdict = 6;
  optional bool active_only = 7;
}

message ListThreatLookupResultsByObservableResponse {
  repeated ThreatLookupResult items = 1;
  string next_page_token = 2;
}

message GetThreatLookupSummaryByObservableRequest {
  string observable_id = 1;
}

message GetThreatLookupSummaryByObservableResponse {
  string observable_id = 1;
  int32 total_count = 2;
  optional string highest_verdict = 3;
  optional double max_risk_score = 4;
  repeated string source_names = 5;
  google.protobuf.Timestamp latest_received_at = 6;
}

message EnrichmentResult {
  string id = 1;
  string case_id = 2;
  string observable_id = 3;
  string enrichment_type = 4;
  string source_name = 5;
  string source_record_id = 6;
  string status = 7;
  string summary = 8;
  string result_data_json = 9;
  double score = 10;
  double confidence = 11;
  google.protobuf.Timestamp requested_at = 12;
  google.protobuf.Timestamp received_at = 13;
  google.protobuf.Timestamp expires_at = 14;
  string last_updated_by = 15;
  google.protobuf.Timestamp created_at = 16;
  google.protobuf.Timestamp updated_at = 17;
}

message ThreatLookupResult {
  string id = 1;
  string case_id = 2;
  string observable_id = 3;
  string lookup_type = 4;
  string source_name = 5;
  string source_record_id = 6;
  string verdict = 7;
  double risk_score = 8;
  double confidence_score = 9;
  string tags_json = 10;
  string matched_indicators_json = 11;
  string summary = 12;
  string result_data_json = 13;
  google.protobuf.Timestamp queried_at = 14;
  google.protobuf.Timestamp received_at = 15;
  google.protobuf.Timestamp expires_at = 16;
  google.protobuf.Timestamp created_at = 17;
  google.protobuf.Timestamp updated_at = 18;
}
```
