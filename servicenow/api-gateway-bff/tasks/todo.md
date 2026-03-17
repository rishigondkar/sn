# API Gateway BFF – Implementation Plan

**Source:** 01_api_gateway_bff_prd.md, 00_platform_integration_contract.md  
**Policy:** Plan mode first; no ad-hoc changes. Check off as completed.

---

## 1. Project layout (per PRD Go Project Structure)

- [x] Create `go.mod` (module `github.com/servicenow/api-gateway`, Go 1.23).
- [x] Create directory layout:
  - `cmd/main.go` – HTTP server only (REST + health), config, logging, graceful shutdown; no gRPC server.
  - `api/` – Router, middleware, route handlers under `/api/v1`.
  - `orchestrator/` – Client interfaces and composition (e.g. case detail); no DB.
  - `clients/` – gRPC client constructors and wrappers; dial with timeouts; propagate `x-user-id`, `x-request-id`, `x-correlation-id` in metadata.
  - `config/`, `logging/` – Config and structured logging.
- [x] No `handler/`, no `repository/`, no `proto/` in this repo (gateway consumes other services’ protos or generated stubs; document stub source in README).

---

## 2. Proto / gRPC client stubs

- [x] API Gateway does **not** define its own proto (per PRD). It consumes Case, Observable, Enrichment, Reference, Attachment, Audit service APIs.
- [x] Define **downstream client interfaces** in clients for: CaseCommandService, CaseQueryService, ObservableCommandService, ObservableQueryService, EnrichmentQueryService, ThreatLookupQueryService, ReferenceQueryService, AttachmentQueryService, AuditQueryService (+ AttachmentCommandService).
- [x] Implement **stub clients** so the app builds and tests run without vendored protos. README documents where to get real stubs and how to wire them.

---

## 3. Migrations

- [x] **None.** This service does not own data (no Appendix A; no repository layer for domain data).

---

## 4. API layer (REST)

- [x] Register all routes under `/api/v1` (base path per platform contract): cases, worknotes, assign, close, observables, alerts, enrichment-results, attachments, audit-events, detail, threat-lookups, attachments, reference users/groups.
- [x] Request/response DTOs; validate JSON and required fields before calling downstream.
- [x] Middleware: auth (X-User-Id, X-User-Name), request ID, correlation ID, logging, recovery. Context keys set for downstream metadata.
- [x] Normalize errors to shared REST error envelope (code, message, details, requestId, correlationId).
- [x] Health: `GET /health` (200 when up). 503 on critical downstream not implemented in stubs.
- [x] CORS: configurable origins/methods in config and middleware.

---

## 5. Orchestrator

- [x] Orchestrator depends on client interfaces only (no gRPC types in API layer).
- [x] Per-route orchestration in handlers calling Orch.CaseCmd, Orch.CaseQuery, etc.
- [x] **GET /cases/{caseId}/detail:** Case core first; fan-out to worknotes, alerts, observables, enrichment, threat lookups, attachments, audit; phase 2 child observables; reference expansion; DegradedSections and logging.

---

## 6. Clients

- [x] Per-domain client files: case_client, observable_client, enrichment_client, reference_client, attachment_client, audit_client (+ _stub.go).
- [x] Stub constructors; real clients would dial with timeout and propagate metadata (documented in README).
- [x] Stub implementations compile without vendored protos; README documents how to wire real proto-generated clients.

---

## 7. Config and logging

- [x] Config: HTTP port, downstream addrs, timeouts, CORS from env.
- [x] Structured logging (slog); request_id/correlation_id from context in logging.FromContext.

---

## 8. Graceful shutdown and health

- [x] SIGTERM/SIGINT: signal.Notify, srv.Shutdown with configurable timeout.
- [x] Health: 200 when up; 503 for critical dependency can be added when real clients are wired.

---

## 9. Tests (per PRD)

- [x] Route handler tests (health, CreateCase validation/success, GetCase error).
- [x] Auth and RequestID middleware tests.
- [x] Orchestrator tests with mocks (GetCaseDetail required failure, success, empty caseId).
- [x] Error mapping: validation and downstream errors return REST envelope (CodeValidationError, CodeInternal).
- [ ] Idempotency key pass-through: header accepted and can be forwarded (stub clients don’t use it yet).
- [x] Pagination: parsePagination used in list handlers; validation in handlers.

---

## 10. PRD-specific and P0

- [ ] Idempotency-Key: CORS allows header; pass-through to downstream can be added when real clients are wired.
- [x] Config has DownstreamTimeout; retries/circuit breaking to be added when real gRPC clients are implemented.
- [x] No business rules or DB writes; no event bus publishing.
- [x] README: structure, routes, config, run, client stubs and regeneration.

---

## 11. Agent Checklist (from PRD)

- [x] Set module name in go.mod and imports.
- [x] Implement routes and DTOs; register under `/api/v1`.
- [x] Define downstream client interfaces (Case, Observable, Enrichment, Reference, Attachment, Audit).
- [x] Implement orchestrator per endpoint; case detail: fan-out and partial-failure handling.
- [x] Add middleware: auth, request ID, correlation ID, logging, recovery, CORS.
- [x] Wire health at `GET /health`; graceful shutdown on SIGTERM/SIGINT.
- [x] Add tests (handlers, auth, error mapping, aggregation partial-failure).
- [x] Update README (structure, routes, how to run, client stubs).

---

## Review / results

- **Build:** `go build ./...` and `go test ./...` pass.
- **Layout:** Matches PRD (cmd, api, orchestrator, clients, config, logging). No handler/, repository/, proto/.
- **Routes:** All core and aggregated routes registered under `/api/v1`; health at `GET /health`.
- **Stubs:** Client interfaces and stub implementations allow the gateway to run without downstream services; README documents wiring real gRPC clients.
- **Remaining / follow-up:**
  - Wire real gRPC clients when Case, Observable, Enrichment, Reference, Attachment, Audit protos are available; implement dial + metadata propagation in clients.
  - Add Idempotency-Key forwarding in write handlers and to downstream metadata when supported.
  - Optional: health 503 when critical downstream (e.g. Case Service) is unreachable.
  - Optional: retries and circuit breaking on downstream gRPC once real clients are in place.
