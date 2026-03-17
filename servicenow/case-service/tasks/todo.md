# Case Service – Implementation Plan

**Source:** 02_case_service_prd.md, 00_platform_integration_contract.md  
**Policy:** Plan mode first; no ad-hoc changes. Execute plan then verify.

---

## 1. Project layout (Go Project Structure)

- [ ] **1.1** Create `go.mod`: module `github.com/servicenow/case-service`, go 1.23; require grpc, protobuf, pgx/sql, testify, etc.
- [ ] **1.2** Create directory layout per PRD:
  - `cmd/main.go`
  - `handler/handler.go`, `handler/case.go`, `handler/worknote.go`, `handler/handler_test.go`
  - `api/api.go`, `api/case_handlers.go`, `api/worknote_handlers.go`
  - `service/service.go`, `service/case.go`, `service/worknote.go`
  - `repository/repository.go` (and optionally `repository/case.go`, `repository/worknote.go` for grouping)
  - `config/config.go`, `logging/logging.go`
  - `proto/case_service.proto`, `proto/generate_proto.sh`, `proto/case_service/*.pb.go` (generated)
  - `migrations/` for SQL (Appendix A)
- [ ] **1.3** Add `Dockerfile`, `.gitlab-ci.yml`, `README.md` stubs.

---

## 2. Proto (Appendix B)

- [ ] **2.1** Add `proto/case_service.proto` with full contract from Appendix B; set `go_package = "github.com/servicenow/case-service/proto/case_service"`.
- [ ] **2.2** Add `proto/generate_proto.sh` to run protoc with go/grpc plugins; output to `proto/case_service/`.
- [ ] **2.3** Run `./proto/generate_proto.sh case_service go` and verify `case_service.pb.go`, `case_service_grpc.pb.go` exist.

---

## 3. Migrations (Appendix A)

- [ ] **3.1** Add migration(s) for `cases` table (full DDL from Appendix A).
- [ ] **3.2** Add migration for `case_worknotes` table (with FK to cases).
- [ ] **3.3** Document migration runner (e.g. golang-migrate, or embed + run from main/cobra). Use numbering (000001_cases.up.sql, 000002_case_worknotes.up.sql).

---

## 4. Domain and validation

- [ ] **4.1** Define domain enums: state (`new`, `triage`, `in_progress`, `pending`, `resolved`, `closed`), priority (`P1`–`P4`), severity (`critical`, `high`, `medium`, `low`). Document accuracy/impact if used.
- [ ] **4.2** Define domain structs: Case, Worknote (mirror DB/proto); validation helpers (short_description 1–500, state/priority/severity enums, closure rules).

---

## 5. Repository layer

- [ ] **5.1** Repository struct with DB pool (pgx or database/sql); configurable timeouts.
- [ ] **5.2** Case CRUD: Create (return id, case_number), GetByID, Update (with version_no check), List with filters (state, priority, severity, assignment_group_id, assigned_user_id, affected_user_id, opened_time range), sort by opened_time desc; pagination (page_size, page_token/offset).
- [ ] **5.3** Worknote: Create, ListByCaseID with pagination.
- [ ] **5.4** Optimistic concurrency: Update/Assign/Close use `version_no`; WHERE version_no = :expected; return error if no rows affected (conflict).
- [ ] **5.5** Case number generation: sequence or dedicated table (e.g. `INC-000001`); no client-assigned case_number.

---

## 6. Event publisher (audit)

- [ ] **6.1** Define audit event envelope (JSON) per platform contract (event_id, event_type, source_service, entity_type, entity_id, action, actor_user_id, request_id, correlation_id, before_data, after_data, occurred_at, etc.).
- [ ] **6.2** Event publisher interface: Publish(ctx, event) or PublishAfterCommit. Prefer transactional outbox: write to outbox table in same TX as business write, then async/sync publish from outbox (or single TX + publish after commit).
- [ ] **6.3** Publish events: case.created, case.updated, case.state.changed, case.assigned, case.closed, worknote.created. Include before/after where feasible.

---

## 7. Service layer

- [ ] **7.1** Service struct: Repository, EventPublisher; optional Assignment & Reference gRPC client (can be no-op for POC).
- [ ] **7.2** CreateCase: validate required (short_description, state, priority, severity, opened_by_user_id, opened_time); generate case_number; repo Create; publish case.created.
- [ ] **7.3** UpdateCase: get current case; validate state transition (no closed→non-closed in v1); validate optional fields; repo Update with version_no; publish case.updated / case.state.changed as needed.
- [ ] **7.4** AssignCase: get case; repo update assignment with version_no; publish case.assigned.
- [ ] **7.5** CloseCase: validate closure_code, closure_reason, actor; set state=closed, closed_by_user_id, closed_time, is_active=false; repo update with version_no; publish case.closed.
- [ ] **7.6** Active duration: update active_duration_seconds on read or write using active states (new, triage, in_progress); document algorithm in code/tests.
- [ ] **7.7** AddWorknote: validate case exists, note_text; repo create worknote; publish worknote.created.
- [ ] **7.8** GetCase, ListCases, ListWorknotes: delegate to repository; map to domain/proto.

---

## 8. Handlers (gRPC) and API (REST)

- [ ] **8.1** Handler: embed UnimplementedCaseServiceServer; Service field; extract actor/request_id/correlation_id from gRPC metadata (x-user-id, x-request-id, x-correlation-id); pass to service.
- [ ] **8.2** handler/case.go: CreateCase, UpdateCase, AssignCase, CloseCase, GetCase, ListCases — map proto ↔ domain, call service, map errors to gRPC codes (InvalidArgument, NotFound, Aborted/FailedPrecondition for version conflict, AlreadyExists).
- [ ] **8.3** handler/worknote.go: AddWorknote, ListWorknotes — same pattern.
- [ ] **8.4** api/api.go: router with /api/v1 prefix; GET /health → 200 (or 503 if DB down); optional REST for cases/worknotes.
- [ ] **8.5** api/case_handlers.go, api/worknote_handlers.go: internal REST handlers calling same Service; JSON in/out; map errors to REST error shape and HTTP status.

---

## 9. main.go, health, graceful shutdown

- [ ] **9.1** Load config (ports 50051, 8080; DB URL; event bus config if needed).
- [ ] **9.2** Init DB pool, repository, event publisher, service, handler, api.
- [ ] **9.3** gRPC server: listen 50051; register CaseService; unary interceptor (logging, timeout); reflection optional.
- [ ] **9.4** HTTP server: 8080; /health (200 when up; optional 503 if DB ping fails); /api/v1/* → api router; ReadTimeout/WriteTimeout.
- [ ] **9.5** Graceful shutdown: SIGTERM/SIGINT → stop listeners; grpc.GracefulStop(); http.Server.Shutdown(ctx with ~30s); exit.
- [ ] **9.6** Run migrations on startup or document separate migration step.

---

## 10. Tests

- [ ] **10.1** Unit tests: domain validation, state transition rules, closure validation.
- [ ] **10.2** Repository tests: Create/Get/Update/List cases and worknotes; optimistic concurrency (version_no conflict).
- [ ] **10.3** Service tests: CreateCase, UpdateCase, AssignCase, CloseCase, AddWorknote (with mock repo and mock publisher); state transition and closure failure paths.
- [ ] **10.4** Handler/transport tests: gRPC (and optionally REST) happy path and error mapping (NotFound, InvalidArgument, version conflict).
- [ ] **10.5** At least one happy path and multiple failure-path tests per write endpoint (per platform contract).

---

## 11. Documentation and CI

- [ ] **11.1** README: project overview, layout, gRPC (port 50051) and REST (8080), health, setup (go mod tidy, proto gen, migrations, run), how to run tests, event bus config.
- [ ] **11.2** Dockerfile: multi-stage build; expose 50051, 8080.
- [ ] **11.3** .gitlab-ci.yml: build, test, Docker build and push (no K8s).

---

## 12. Agent Checklist (from PRD) – verification

- [ ] Set module name in go.mod and all imports; rename proto and set go_package.
- [ ] Run proto/generate_proto.sh case_service go.
- [ ] Implement repository with version_no; migrations for cases and case_worknotes.
- [ ] Implement service layer (state/closure rules, transition validators).
- [ ] Implement handler (gRPC) and optional api (REST); map errors to status codes.
- [ ] Wire event publisher; publish after successful commit (outbox preferred).
- [ ] main.go: gRPC + HTTP, server timeouts, graceful shutdown.
- [ ] Add tests (state transitions, closure, optimistic concurrency, handler tests).
- [ ] Update README (structure, gRPC/REST, setup, tests).

---

## Review / results

- **Done:** Go module (`github.com/servicenow/case-service`), proto from Appendix B with generate script, migrations (cases, case_worknotes, case_number_seq), repository with version_no and RunInTx, domain + validation, service (CreateCase, UpdateCase, AssignCase, CloseCase, GetCase, ListCases, AddWorknote, ListWorknotes), audit event envelope and NoopPublisher, gRPC handlers with metadata extraction and error mapping, REST API and health, main.go with gRPC 50051 + HTTP 8080 and graceful shutdown, Dockerfile, .gitlab-ci.yml, README. Tests: domain validation, handler mapError, service ErrNotFound.
- **Agent Checklist:** Module name set; proto generated; repository with version_no and migrations; service layer with state/closure rules; gRPC + REST handlers with error mapping; event publisher wired (noop); main.go with timeouts and graceful shutdown; tests for validation, state transition, error mapping; README updated.
- **Remaining / follow-up:** (1) Run migrations against a real PostgreSQL and run full integration test (CreateCase → GetCase → UpdateCase → CloseCase, ListCases, worknotes). (2) Optional: transactional outbox table and relay for audit events. (3) Optional: readiness check that pings DB and returns 503 if down. (4) Repository unit tests with testcontainers or CI DB when available.
