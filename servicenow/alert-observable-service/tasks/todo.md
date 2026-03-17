# Alert & Observable Service – Implementation Plan

Source: `03_alert_observable_service_prd.md`, `00_platform_integration_contract.md`.  
Plan-first per Agent Execution Policy; no ad-hoc changes.

---

## 1. Project layout (Go Project Structure)

- [x] Create Go module: `github.com/<org>/alert-observable-service` (use `alert-observable-service` for local path), Go 1.22+.
- [x] Directories: `cmd/`, `handler/`, `service/`, `repository/`, `config/`, `logging/`, `proto/`, `migrations/`. No `api/` (gRPC only per PRD).
- [x] Files: `cmd/main.go`; `handler/handler.go` + `alert.go`, `observable.go`, `case_observable.go`, `similar_incident.go` + `handler_test.go`; `service/service.go` + same resource files; `repository/repository.go`; `config/config.go`; `logging/logging.go`.

---

## 2. Proto (Appendix B)

- [x] Add `proto/alert_observable_service.proto` with full Appendix B content; set `go_package = "github.com/org/alert-observable-service/proto/alert_observable_service"`.
- [x] Add `proto/generate_proto.sh` to run `protoc` with Go plugin; output under `proto/alert_observable_service/` (`.pb.go`, `_grpc.pb.go`).
- [x] Run script; ensure generated code compiles.

---

## 3. Migrations (Appendix A)

- [ ] Add `migrations/` with executable PostgreSQL DDL from Appendix A (order: alert_rules, observables, alerts, case_observables, child_observables, similar_security_incidents).
- [ ] Include indexes: `idx_case_observables_case_id`, `idx_case_observables_observable_id`, `idx_similar_security_incidents_case_id`, `idx_child_observables_parent_observable_id`.

---

## 4. Repository layer

- [ ] Define interfaces and implementations for: AlertRule, Alert, Observable, CaseObservable, ChildObservable, SimilarIncident.
- [ ] Use `context.Context`, structured logging, timeouts; no cross-service DB access.
- [ ] CRUD/list with pagination (page_size, page_token/offset) where required by API.
- [ ] Uniqueness: enforce in DB and handle conflicts (AlreadyExists / idempotent behavior where specified).

---

## 5. Service layer

- [ ] **Normalization:** Implement per-type normalization (ip, domain, hash, url, email, file): canonical form; store `observable_value` + `normalized_value`; do not overwrite original.
- [ ] **CreateOrGetObservable:** Resolve by (observable_type, normalized_value); create if missing; idempotent.
- [ ] **LinkObservableToCase:** Validate case ref; create-or-resolve observable; upsert case_observables (unique case_id, observable_id); recompute similar incidents; publish audit events after successful commit (transaction boundary).
- [ ] **UpdateCaseObservable:** Update tracking_status, accuracy, determination, impact; publish `observable.updated` after persist.
- [ ] **Similar-incident recomputation:** On link/unlink (and RecomputeSimilarIncidentsForCase): find other cases sharing observable(s); upsert `similar_security_incidents` both directions; store count and IDs (and optional display values).
- [ ] **Audit event publisher:** Interface to publish audit envelope (event_id, event_type, source_service, entity_type, entity_id, actor, request_id, correlation_id, change_summary, before/after, occurred_at). Publish only after successful commit. Event types: `alert.created`, `observable.created`, `observable.linked.to_case`, `observable.updated`, `child_observable.created`, `similar_incident.linked`.

---

## 6. Handlers and main

- [ ] **Handlers:** gRPC only. Map validation → `InvalidArgument`, not found → `NotFound`, duplicates → `AlreadyExists`, concurrency → `Aborted`/`FailedPrecondition`. Extract actor, request_id, correlation_id from gRPC metadata (`x-user-id`, `x-request-id`, `x-correlation-id`).
- [ ] **Health:** HTTP server on port 8080 with health endpoint (e.g. `/health` or `/api/v1/health`); no REST API for domain resources.
- [ ] **main.go:** Start gRPC server (port 50051), HTTP server (8080); graceful shutdown (context cancellation, drain gRPC, shutdown HTTP); timeouts on outbound calls; config from env/file.

---

## 7. Tests

- [ ] Unit tests: normalization by type; create-or-get observable (existing vs new); duplicate case-observable link (idempotent); similar-incident recomputation (fixtures with multiple cases/observables).
- [ ] Repository tests: persistence for each entity; list/pagination; uniqueness constraints.
- [ ] Handler tests: at least one success and failure path per write RPC; error code mapping.
- [ ] No external services in unit tests; use mocks for event publisher and optional Case Service.

---

## 8. PRD-specific and checklist

- [ ] Event publisher wired; no consumer in this service (audit consumed elsewhere).
- [ ] Validation: observable_type in (ip, domain, hash, url, email, file); reject self-links for child observable; case_id != similar_case_id.
- [ ] README: how to run, migrate, test; ports 50051 (gRPC), 8080 (HTTP health).
- [ ] Agent Checklist: proto generated; repo + migrations + indexes; normalization + create-or-get + link + similar-incident; handlers + event publisher; main.go gRPC + HTTP + shutdown; tests; README.

---

## Review / results

- **Layout:** Implemented per PRD (cmd, handler, service, repository, config, logging, proto, migrations). No `api/` (gRPC only).
- **Proto:** Appendix B added; `generate_proto.sh` generates Go stubs under `proto/alert_observable_service/`.
- **Migrations:** `migrations/000001_init_schema.up.sql` with all tables and indexes from Appendix A.
- **Repository:** Postgres impl + Tx for link-to-case and similar-incident recomputation; adapters expose interfaces.
- **Service:** Normalization (ip, domain, hash, url, email, file), CreateOrGetObservable, LinkObservableToCase (idempotent, in tx with similar-incident recompute), UpdateCaseObservable, CreateChildObservableRelation, RecomputeSimilarIncidentsForCase; audit events published after commit (no-op publisher by default).
- **Handlers:** gRPC server; metadata (x-user-id, x-request-id, x-correlation-id); errors mapped to InvalidArgument, NotFound, AlreadyExists, Internal.
- **main:** gRPC (50051), HTTP health (8080), graceful shutdown (GracefulStop + Shutdown).
- **Tests:** Normalization and validation (service), mapErr/handler (handler), optional repo test when TEST_DB_DSN set.
- **Remaining:** Wire real AuditPublisher when event bus is available; run migrations against target DB; optional Dockerfile and .gitlab-ci.yml per template.
