# API Gateway / BFF PRD
Version: 1.0  
Service name: api-gateway  
Language: Golang  
Primary responsibility: public REST entry point and response aggregation for UI and external clients

## Mission
Provide a stable public REST API for the web UI and approved external clients. Normalize auth context, validate request structure, route requests to internal services, aggregate read responses where necessary, and hide all internal gRPC topology from the UI.


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
- Public REST API under `/api/v1`
- Auth context extraction and propagation
- Request validation at the boundary
- Request/response shaping for UI
- Aggregated case detail endpoint
- Mapping between REST JSON and internal gRPC contracts
- Idempotency key handling pass-through or enforcement where appropriate
- Request ID and correlation ID generation/propagation

# Out of Scope
- Owning business data
- Direct database writes for domain entities
- Direct S3 access for attachment persistence
- Publishing audit events for domain mutations on behalf of downstream services
- Implementing domain logic that belongs to another service

# Responsibilities
1. Terminate public requests.
2. Authenticate and authorize at coarse-grained route level.
3. Forward writes to the owning service by gRPC or internal REST adapter if needed during transition.
4. Aggregate reads for the case details screen.
5. Normalize error responses to the shared REST error model.
6. Apply timeouts, retries, and circuit breaking on downstream calls.
7. Preserve backward compatibility for the public API.

# Public REST API

## Core Routes
- `POST /api/v1/cases`
- `GET /api/v1/cases/{caseId}`
- `PATCH /api/v1/cases/{caseId}`
- `POST /api/v1/cases/{caseId}/worknotes`
- `GET /api/v1/cases/{caseId}/worknotes`
- `POST /api/v1/cases/{caseId}/assign`
- `POST /api/v1/cases/{caseId}/close`
- `POST /api/v1/cases/{caseId}/observables`
- `GET /api/v1/cases/{caseId}/observables`
- `GET /api/v1/cases/{caseId}/alerts`
- `GET /api/v1/cases/{caseId}/enrichment-results`
- `GET /api/v1/observables/{observableId}/threat-lookups`
- `POST /api/v1/attachments`
- `GET /api/v1/cases/{caseId}/attachments`
- `GET /api/v1/cases/{caseId}/audit-events`
- `GET /api/v1/reference/users`
- `GET /api/v1/reference/groups`

## Aggregated Read Endpoint
### `GET /api/v1/cases/{caseId}/detail`
This endpoint should compose:
- case core data from Case Service
- worknotes from Case Service
- alerts from Alert & Observable Service
- observables from Alert & Observable Service
- child observable graph summary from Alert & Observable Service
- similar incidents from Alert & Observable Service
- enrichment results from Enrichment & Threat Lookup Service
- threat lookup summaries from Enrichment & Threat Lookup Service
- attachment metadata from Attachment Service
- audit timeline from Audit Service
- reference expansions from Assignment & Reference Service

### Response Shape
Return a coherent UI view model. Do not leak gRPC-specific details or internal service errors directly.

### Downstream RPCs for GET /cases/{caseId}/detail
Call these gRPC methods (per service PRDs) so domain services implement the right queries. Use parallel fan-out where possible; tolerate optional section failure.
- **Case Service:** GetCase, ListWorknotes
- **Alert & Observable Service:** ListCaseAlerts, ListCaseObservables, ListChildObservables (for graph summary), ListSimilarIncidents
- **Enrichment & Threat Lookup Service:** ListEnrichmentResultsByCase, ListThreatLookupResultsByCase (or per-observable summaries as needed)
- **Attachment Service:** ListAttachmentsByCase
- **Audit Service:** ListAuditEventsByCase
- **Assignment & Reference Service:** GetUser/GetGroup (or batch) for expanding assigned_user_id, assignment_group_id, opened_by_user_id, etc.

# Downstream Dependencies
- Case Service: required
- Alert & Observable Service: required
- Enrichment & Threat Lookup Service: required for case detail view
- Assignment & Reference Service: required
- Attachment Service: required
- Audit Service: required for audit timeline views

# gRPC Client Contracts To Implement
- `CaseCommandService`
- `CaseQueryService`
- `ObservableCommandService`
- `ObservableQueryService`
- `EnrichmentQueryService`
- `ThreatLookupQueryService`
- `ReferenceQueryService`
- `AttachmentQueryService`
- `AuditQueryService`

The gateway must keep client interfaces separate by domain to reduce accidental misuse.

# Auth and Context Rules
- Read actor identity from trusted upstream auth.
- Generate `request_id` if absent.
- Generate or forward `correlation_id`.
- Inject actor metadata into downstream gRPC metadata using the canonical keys from the platform contract (Identity Propagation): `x-user-id`, `x-request-id`, `x-correlation-id`.
- Never trust raw client-supplied internal actor fields unless verified by auth middleware.

# Validation Rules
- Validate JSON shape and required fields before calling downstream services.
- Reject unknown enum values if the route contract requires known values.
- Enforce pagination parameter bounds.
- Enforce file upload size metadata validation before invoking Attachment Service.
- For browser-based clients, set CORS headers (including allowed origins and methods) on REST responses; preflight and credentials behavior must be documented or configurable.

# Error Handling
- Normalize all downstream errors into the shared REST error envelope.
- Distinguish validation, auth, not found, dependency unavailable, and conflict cases.
- Do not expose stack traces.
- Preserve request IDs and correlation IDs in errors.

# Retry and Timeout Policy
- Apply short timeouts to all downstream calls.
- Retry only safe idempotent reads by default.
- Do not retry non-idempotent writes unless the idempotency contract is in place.
- Surface partial failure clearly for aggregated endpoints if some optional dependencies fail.

# Aggregation Strategy
For `GET /cases/{caseId}/detail`:
1. Fetch case core first.
2. Fan out in parallel to supporting services.
3. Tolerate optional section failures where the screen can still render.
4. Return explicit section-level degradation flags for UI if needed.
5. Log which downstream service degraded the response.

# Data It Must Never Own
- case tables
- observables
- threat lookups
- enrichment rows
- attachments
- audit event store

# Non-Functional Requirements
- p95 latency target for non-aggregated endpoints: under 200 ms excluding network variance
- p95 latency target for aggregated detail endpoint: under 700 ms under normal dependency health
- graceful degradation when optional dependencies fail
- structured request logs
- route-level metrics and tracing

# Testing Requirements
- route handler tests
- auth middleware tests
- downstream client contract tests
- error mapping tests
- aggregation partial-failure tests
- idempotency key pass-through tests
- pagination and validation tests

# Build Instructions for Agents
1. Create public REST routes first.
2. Define response DTOs and request validators.
3. Define downstream client interfaces second.
4. Implement orchestration layer per endpoint.
5. Add middleware for auth, request ID, correlation ID, logging, and recovery.
6. Add tests before declaring any route complete.
7. Verify aggregated endpoint against mocked downstream services.

# Go Project Structure

Follow the canonical template for layout and conventions where applicable: [PRD_GO_GRPC_SERVICE_TEMPLATE.md](../golang_prd/PRD_GO_GRPC_SERVICE_TEMPLATE.md). This service **deviates** from the full template: it has no gRPC server, no domain repository, and no own proto.

## Repository Layout

```
<project-root>/
├── cmd/
│   └── main.go
├── api/
│   ├── api.go
│   ├── case_handlers.go
│   ├── observable_handlers.go
│   ├── attachment_handlers.go
│   ├── audit_handlers.go
│   └── reference_handlers.go
├── orchestrator/
│   ├── orchestrator.go
│   └── case_detail.go
├── clients/
│   ├── clients.go
│   ├── case_client.go
│   ├── observable_client.go
│   ├── enrichment_client.go
│   ├── reference_client.go
│   ├── attachment_client.go
│   └── audit_client.go
├── config/
│   └── config.go
├── logging/
│   └── logging.go
├── go.mod
├── go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

- **No** `handler/` (no gRPC server). **No** `repository/` for domain data. **No** `proto/` for this service (use other services' protos or generated client code; document where client stubs come from, e.g. vendored or shared proto repo).
- `cmd/main.go`: Wire HTTP server only (REST + health), config, logging, graceful shutdown. No gRPC server.
- `api/`: Router, middleware (auth, request ID, correlation ID, logging, recovery), route handlers under `/api/v1`; handlers call orchestrator.
- `orchestrator/`: Holds gRPC client interfaces; composes responses (e.g. case detail aggregation); no DB.
- `clients/`: gRPC client constructors and wrappers; dial with timeouts; propagate actor, request_id, correlation_id in metadata.

## Module, Ports, P0

| Item | Value |
|------|--------|
| Module | `github.com/<org>/api-gateway` |
| Port | HTTP only, e.g. **8080** (REST + health) |
| Go version | 1.22+ (platform contract); use 1.23 in go.mod if following template |

**P0 (production):** Auth middleware on REST; normalize errors to shared REST envelope; timeouts and retries/circuit breaking on downstream gRPC; graceful shutdown (drain in-flight requests); health endpoint `GET /health` (200 when up; optional 503 if critical downstream unreachable); no secrets in logs.

## Agent Checklist

- [ ] Set module name in go.mod and imports.
- [ ] Implement routes and DTOs; register under `/api/v1`.
- [ ] Define downstream client interfaces (Case, Observable, Enrichment, Reference, Attachment, Audit).
- [ ] Implement orchestrator per endpoint; case detail: fan-out and partial-failure handling.
- [ ] Add middleware: auth, request ID, correlation ID, logging, recovery.
- [ ] Wire health at `GET /health`; graceful shutdown on SIGTERM/SIGINT.
- [ ] Add tests (handlers, auth, error mapping, aggregation partial-failure).
- [ ] Update README (structure, routes, how to run).

### Proto Dependencies and Build Order (POC)

- API Gateway consumes other services' gRPC APIs; it MUST vendor or import proto definitions (or generated Go stubs) for:
  - Case, Alert/Observable, Enrichment & Threat Lookup, Assignment & Reference, Attachment, and Audit services.
- Recommended build order:
  - Define and generate protos for services 02–07 first.
  - Gateway then depends on those protos or generated clients (for example, via shared proto module or vendored generated code).
- The Gateway `README.md` MUST document:
  - Where client stubs come from (shared module vs vendored files).
  - How to regenerate them when upstream protos change.
- Request/response message shapes for each downstream service are defined in that service PRD's **Appendix B: Proto contract**.

# Do Not Do
- Do not embed business rules that belong to domain services.
- Do not write directly to any domain database.
- Do not call the event bus for case or observable changes.
- Do not expose internal gRPC error internals in REST responses.
- Do not silently swallow dependency failures.
