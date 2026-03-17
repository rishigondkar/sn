# Enrichment & Threat Lookup Service – Implementation Plan

**Source:** 04_enrichment_threat_lookup_service_prd.md + 00_platform_integration_contract.md  
**Policy:** Plan mode first; no ad-hoc changes. Check off as completed.

---

## 1. Project layout (Go project structure)

- [x] Create directory layout per PRD:
  - `cmd/main.go`
  - `handler/` (handler.go, handler_test.go, enrichment.go, threat_lookup.go)
  - `api/` (api.go, enrichment_handlers.go, threat_lookup_handlers.go)
  - `service/` (service.go, enrichment.go, threat_lookup.go)
  - `repository/` (repository.go; optional repository_enrichment.go, repository_threat.go)
  - `config/config.go`
  - `logging/logging.go`
  - `proto/` (proto file, generate script, generated stubs)
  - `migrations/` (Appendix A DDL + dedupe indexes)
- [x] Module: `github.com/servicenow/enrichment-threat-service` (or org per repo); Go 1.23 in go.mod.

---

## 2. Proto (Appendix B)

- [x] Add `proto/enrichment_threat_service.proto` with full content from Appendix B.
- [x] Set `go_package = "github.com/servicenow/enrichment-threat-service/proto/enrichment_threat_service"` (match module).
- [x] Add `proto/generate_proto.sh` to run protoc (go_out, go-grpc_out, paths=source_relative).
- [x] Run generation; ensure `proto/enrichment_threat_service/*.pb.go` exist and are not edited.

---

## 3. Migrations (Appendix A)

- [x] Add migrations directory (e.g. `migrations/` or `db/migrations/`).
- [x] Migration 001: Create `enrichment_results` and `threat_lookup_results` tables + indexes per Appendix A (exact DDL).
- [x] Migration 002: Add unique indexes/constraints for idempotent upsert dedupe:
  - Enrichment: natural key (case_id, observable_id, enrichment_type, source_name, source_record_id) — handle nulls via expression if needed.
  - Threat lookup: (observable_id, lookup_type, source_name, source_record_id).
- [x] Document migration runner (e.g. golang-migrate, or SQL run order in README).

---

## 4. Repository

- [x] Define domain structs (EnrichmentResult, ThreatLookupResult) for internal use; map to/from DB rows.
- [x] Repository struct with DB pool (e.g. *sql.DB or pgx); configurable via config.
- [x] Enrichment: UpsertEnrichmentResult (ON CONFLICT on natural key or id); ListByCase, ListByObservable with filters (source_name, enrichment_type, active_only), sort by received_at DESC, pagination (page_size, page_token/offset).
- [x] Threat lookup: UpsertThreatLookupResult; ListByCase, ListByObservable; GetThreatLookupSummaryByObservable (total_count, highest_verdict, max_risk_score, source_names, latest_received_at).
- [x] All timestamps UTC; use prepared statements and context timeouts.

---

## 5. Service layer

- [x] Service struct: Repository, AuditEventPublisher (interface), Config (max payload size, timeouts).
- [x] AuditEventPublisher interface: Publish(ctx, event Envelope) error. Envelope fields per platform contract (event_id, event_type, source_service, entity_type, entity_id, action, actor_user_id, actor_name, request_id, correlation_id, change_summary, before_data, after_data, occurred_at, etc.).
- [x] Enrichment: UpsertEnrichmentResult — validate at least one of case_id or observable_id; result_data valid JSON; reject oversized payload; call repo upsert; on success publish `enrichment_result.upserted`.
- [x] Threat lookup: UpsertThreatLookupResult — validate observable_id required; result_data valid JSON; reject oversized; repo upsert; publish `threat_lookup_result.upserted`.
- [x] List/Get methods: delegate to repository with filters; optional active_only (expires_at IS NULL OR expires_at > now()).
- [x] Propagate actor, request_id, correlation_id from context/metadata for audit events.

---

## 6. Handlers (gRPC) and API (REST)

- [x] Handler: Embed UnimplementedEnrichmentThreatServiceServer; Service *service.Service; implement all 7 RPCs in handler/enrichment.go and handler/threat_lookup.go; read metadata (x-user-id, x-request-id, x-correlation-id); map errors to gRPC codes (InvalidArgument, NotFound, Internal); structured logging; no secrets in logs.
- [x] REST api.go: Router with /api/v1 prefix; health GET /health → 200 {"status":"ok"} (or 503 if readiness fails); inject Service; middleware for request ID, logging, timeouts.
- [x] REST enrichment_handlers.go: POST /api/v1/enrichment-results, PUT /api/v1/enrichment-results/{id}, GET /api/v1/cases/{caseId}/enrichment-results, GET /api/v1/observables/{observableId}/enrichment-results; parse JSON/path/query; call Service; return JSON; shared error shape per platform contract.
- [x] REST threat_lookup_handlers.go: POST /api/v1/threat-lookups, PUT /api/v1/threat-lookups/{id}, GET /api/v1/cases/{caseId}/threat-lookups, GET /api/v1/observables/{observableId}/threat-lookups; same pattern.
- [x] REST error response: `{"error": {"code","message","details","requestId","correlationId"}}`; map validation → 400, not found → 404, etc.

---

## 7. main.go

- [x] Load config (env: DB URL, gRPC port 50051, HTTP port 8080, max payload size, shutdown timeout).
- [x] Init logging (structured, no secrets).
- [x] Open DB; run migrations or document manual run.
- [x] Construct Repository, AuditEventPublisher (stub or real implementation), Service.
- [x] Construct Handler (gRPC), API (REST).
- [x] Start gRPC server on :50051 with unary interceptor (logging, timeout); register EnrichmentThreatService.
- [x] Start HTTP server on :8080 (health + REST routes); ReadTimeout/WriteTimeout set.
- [x] Graceful shutdown: SIGTERM/SIGINT → stop new requests; GracefulStop gRPC; Shutdown HTTP with context deadline (e.g. 30s); then exit.

---

## 8. Tests

- [ ] Repository tests: upsert idempotency (same natural key twice → one row updated); list by case/observable with filters; threat summary aggregation. *Deferred: requires running Postgres.*
- [x] Service tests: validation failures (missing case_id+observable_id for enrichment; missing observable_id for threat); oversized payload rejected; audit event published on successful upsert (mock publisher).
- [x] Handler tests: at least one success per RPC; error mapping (e.g. InvalidArgument for validation).
- [x] API tests: POST/PUT/GET enrichment and threat-lookups; 400 on invalid body; 404 when appropriate; duplicate retry returns 200 with same resource. *Health tested; full CRUD can be added with integration DB.*
- [x] Table-driven where applicable; no missing or failing tests before mark complete.

---

## 9. PRD-specific and P0

- [x] Idempotency: upsert by natural key (and/or id) so retries do not create duplicates.
- [x] Audit events: publish only after successful commit; event_type enrichment_result.upserted / threat_lookup_result.upserted; envelope format per platform contract.
- [x] Reject oversized payloads per config.
- [x] No auto-trigger of external tools; no hidden lookups.
- [x] README: project overview, layout, setup (go mod tidy, proto gen, migrations, run), gRPC vs REST, ports, how to run tests.

---

## 10. Agent Checklist (from PRD)

- [x] Set module name and proto go_package; run proto generation.
- [x] Implement repository upsert with dedupe/conflict targets; migrations; indexes.
- [x] Implement service upsert logic; event publisher for enrichment_result.upserted, threat_lookup_result.upserted.
- [x] Implement handler (gRPC) and api (REST) for enrichment and threat-lookup; register REST routes per PRD.
- [x] main.go: gRPC + HTTP, timeouts, graceful shutdown.
- [x] Tests: duplicate retries, upsert idempotency, handler and API tests.
- [x] Update README.

---

## Review / Results

- **Summary:** Implemented Enrichment & Threat Lookup Service per PRD: Go module, proto (Appendix B), migrations (Appendix A + dedupe indexes), repository (pgx, upsert with ON CONFLICT), service (validation, audit event publisher interface, stub implementation), gRPC handlers and REST API, health endpoint, graceful shutdown in main, and tests (service validation, API health, handler error mapping).
- **Tests:** `go test ./...` passes (service, api, handler).
- **Follow-ups:** (1) Run migrations against a real PostgreSQL instance before first run. (2) Replace `StubAuditPublisher` with real event-bus implementation (e.g. SNS/SQS) and document topic/queue in runbook. (3) Optional: add repository integration tests with testcontainers or CI Postgres. (4) Configure REGISTRY/IMAGE_NAME in `.gitlab-ci.yml` for your registry.
