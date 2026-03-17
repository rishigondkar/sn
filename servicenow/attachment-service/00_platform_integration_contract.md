# Platform Integration Contract PRD
Version: 1.0  
Status: Build baseline  
Primary language: Golang  
Audience: LLMs and coding agents building services in independent chats

## Purpose
This document defines the shared contracts, non-functional requirements, cross-service patterns, and rules that every service must follow so that independently built services work cohesively without runtime contract mismatches.


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

## Project layout
Follow [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md) for repository structure and conventions. Each service PRD specifies its layout and any deviations (e.g. API Gateway has no gRPC server; Audit Service adds an event consumer). Proto definitions live per-service under each repo's `proto/` directory; API Gateway consumes other services' protos or generated client code (e.g. vendored or shared module). Service discovery and peer addresses are via config or environment; do not hardcode.


# System Overview

The platform is a SOC case management system composed of these services:

1. API Gateway / BFF
2. Case Service
3. Alert & Observable Service
4. Enrichment & Threat Lookup Service
5. Assignment & Reference Service
6. Attachment Service
7. Audit Service

## Communication Rules
- UI to backend uses REST over HTTPS only.
- Service-to-service communication uses gRPC only.
- Cross-service audit uses pub/sub only.
- Services own their own data stores. No direct cross-service SQL access.

## Runtime Topology
- All backend services run on ECS Fargate.
- Public ingress terminates at API Gateway / BFF behind an ALB.
- Internal services are private and reachable only inside the VPC.
- Event bus is used for audit and derived workflows.
- Attachments are stored in object storage such as S3.
- Each service uses its own logical database/schema ownership boundary.

## POC Scope Exclusions
- Response tasks (subtasks under a case) are out of scope; Case Service owns the case record and worknotes only.
- Dedicated Ingestion Service (import/transform pipelines) is out of scope; case creation is via API Gateway or direct service calls.
- Playbooks, risk scoring engine, and standalone Analytics/MTTR service are out of scope for this POC.

# Canonical Domain Concepts

## Identifiers
Every primary entity uses two identifiers where appropriate:
- `id`: UUID, internal primary key.
- `*_number` or external-friendly key where needed, such as `case_number`.

### UUID Rules
- Use UUIDv4 or UUIDv7 consistently within a service.
- Never expose sequential database IDs.
- Cross-service references use UUID strings.

### Tenancy (POC)
- POC is single-tenant; individual service schemas do not require `tenant_id`. Multi-tenancy can be introduced in a later phase.

## Timestamps
- Store all timestamps in UTC.
- Use RFC3339 in REST and protobuf `Timestamp` in gRPC.
- Required fields must not be nullable unless explicitly allowed.

## Actor Metadata
Every write request must propagate:
- actor user id
- actor display name if available
- request id
- correlation id

# Shared Error Model

## REST Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "short human readable message",
    "details": [
      {
        "field": "short_description",
        "issue": "required"
      }
    ],
    "requestId": "req-123",
    "correlationId": "corr-456"
  }
}
```

## gRPC Error Rules
- Use canonical gRPC status codes.
- Put machine-readable error metadata in trailers where useful.
- Validation failures map to `InvalidArgument`.
- Missing resources map to `NotFound`.
- optimistic concurrency failures map to `Aborted` or `FailedPrecondition`.
- duplicate/idempotency violations map to `AlreadyExists`.

# Shared Security and Auth Rules
- Assume upstream authentication exists before API Gateway.
- API Gateway forwards trusted identity headers or tokens to downstream services.
- Internal gRPC calls use service-to-service auth and mTLS where available.
- Every write endpoint must enforce authorization checks.
- Every service must log actor, request id, and correlation id for writes.

## Identity Propagation (POC)
- REST (between client and API Gateway):
  - Use headers such as `X-User-Id`, `X-Request-Id`, and `X-Correlation-Id` to carry actor and tracing information.
  - API Gateway is responsible for mapping upstream auth to these headers; backend services must not trust these headers directly from untrusted clients.
- gRPC (between services):
  - Use metadata keys `x-user-id`, `x-request-id`, and `x-correlation-id` to propagate actor and tracing information.
  - API Gateway MUST inject these metadata keys on outbound gRPC calls; all services MUST read/write the same keys.

# Shared Observability Rules
- Structured logs only.
- Emit metrics for request count, latency, error rate, DB latency, outbound dependency latency, queue lag where relevant.
- Emit traces for REST, gRPC, DB, and event publication/consumption.
- Include request id and correlation id in logs, metrics dimensions when supported, and trace context propagation.

# Shared Idempotency Rules
- All externally callable create/update endpoints that may be retried must support idempotency.
- Accept an `Idempotency-Key` header for REST write operations where duplicate submissions are possible.
- Store dedupe records or natural unique keys to prevent duplicate side effects.
- Event consumers must be idempotent using `event_id`.

# Shared Audit Event Contract

All services publish audit events after successful business state changes.

## Event Envelope
```json
{
  "event_id": "uuid-or-ulid",
  "event_type": "case.updated",
  "source_service": "case-service",
  "entity_type": "case",
  "entity_id": "uuid",
  "parent_entity_type": "case",
  "parent_entity_id": "uuid",
  "case_id": "uuid",
  "observable_id": "uuid",
  "action": "update",
  "actor_user_id": "uuid",
  "actor_name": "Jane Analyst",
  "request_id": "req-123",
  "correlation_id": "corr-456",
  "change_summary": "priority changed from p3 to p2",
  "before_data": {},
  "after_data": {},
  "metadata": {},
  "occurred_at": "2026-03-14T15:20:00Z"
}
```

## Event Publication Rules
- Publish only after successful commit of business state.
- Prefer transactional outbox pattern.
- Events are append-only facts.
- Do not publish partial or speculative state.
- Do not mutate previously published events.
- Use stable `event_type` names in dot notation.

## Audit Event Bus (POC)
- The event bus is technology-agnostic but MUST support:
  - Publishing a single audit event envelope (JSON) as the message body.
  - At-least-once delivery of each event to the Audit Service consumer.
  - Idempotent consumption in Audit Service based on `event_id`.
- Message body on the wire MUST be exactly the audit event envelope JSON (no additional wrapper), so all producers and the Audit Service agree on the format.
- For the POC, the chosen transport (for example, SNS/SQS, Kafka, or another pub/sub system) and the topic/queue naming (for example, a single topic such as `soc-audit-events`) MUST be documented in the deployment README or runbook.

## Required Event Types by Domain
- Case: `case.created`, `case.updated`, `case.state.changed`, `case.assigned`, `case.closed`, `worknote.created`
- Observable: `alert.created`, `observable.created`, `observable.linked.to_case`, `observable.updated`, `child_observable.created`, `similar_incident.linked`
- Enrichment: `enrichment_result.upserted`, `threat_lookup_result.upserted`
- Attachment: `attachment.uploaded`, `attachment.deleted`, `attachment.metadata.updated`
- Reference: `user.updated`, `group.updated` only if those changes are audited

# Shared REST Conventions
- Base path prefix: `/api/v1`
- Use plural nouns.
- Use JSON request and response bodies.
- Support pagination with `page_size`, `page_token`, or offset/limit consistently within a service.
- Support filtering and sorting only where specified.
- Return `201 Created` for creates, `200 OK` for updates/reads, `204 No Content` for deletes where no body is returned.

# Shared gRPC Conventions
- Keep protobuf field names stable and additive.
- Never reuse field numbers.
- One proto package per service.
- Prefer explicit request and response messages, not primitive returns.
- Include `request_id`, `correlation_id`, and actor fields in metadata or message headers.
- Distinguish read APIs from command APIs.

# Shared Database Rules
- No cross-service foreign keys.
- Cross-service references are stored as UUID values and validated via service APIs where needed.
- Every table must have `created_at` and `updated_at` unless append-only.
- Use optimistic concurrency where simultaneous updates are plausible.
- JSONB is allowed only for flexible payloads explicitly defined in this contract, such as raw source payloads or enrichment result data.

# Shared Enum Conventions (POC)
- Case state (Case Service):
  - Allowed values: `new`, `triage`, `in_progress`, `pending`, `resolved`, `closed`.
- Case priority / severity:
  - POC MAY use a simple set such as priorities `P1`–`P4` and severities `critical`, `high`, `medium`, `low`. Each service that exposes these fields MUST document its exact allowed values in its own PRD and API contract.
- Observable type (Alert & Observable Service, v1):
  - Allowed values: `ip`, `domain`, `hash`, `url`, `email`, `file`. Additional types MAY be added in later iterations but MUST be documented before use.
- Observable tracking status (Alert & Observable Service):
  - Allowed values: `new`, `under_review`, `enriched`, `dismissed`, `confirmed`.

# Shared Testing Requirements
Every service must include:
- unit tests for domain logic
- repository tests for persistence logic
- transport tests for REST or gRPC handlers
- contract tests against the shared schema
- event publication/consumption tests where relevant
- at least one happy path and multiple failure-path tests per write endpoint

# Shared Delivery Checklist
Before marking a service complete:
- schema created and migration tested
- REST endpoints validated
- gRPC server/client contracts validated
- event publishing/consumption validated
- logs/metrics/traces present
- auth and authorization checks present
- idempotency behavior verified
- retries and timeouts configured
- README and runbook written
- `tasks/todo.md` reviewed and completed
- `tasks/lessons.md` updated if needed
