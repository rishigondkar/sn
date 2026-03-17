# Assignment Reference Service – Implementation Plan

**Source:** `05_assignment_reference_service_prd.md` + `00_platform_integration_contract.md`

## 1. Project layout (per PRD Go Project Structure)

- [ ] **1.1** Create directory structure: `cmd/`, `handler/`, `api/`, `service/`, `repository/`, `config/`, `logging/`, `proto/`, `migrations/`.
- [ ] **1.2** Set module: `github.com/servicenow/assignment-reference-service` in `go.mod` (Go 1.22+; use 1.23 if template).
- [ ] **1.3** Add `go.mod` / `go.sum` with grpc, protobuf, testify; no extra APIs beyond PRD.

## 2. Proto (Appendix B)

- [ ] **2.1** Add `proto/assignment_reference_service.proto` with full Appendix B content; set `go_package = "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"`.
- [ ] **2.2** Add `proto/generate_proto.sh` (protoc with --go_out, --go-grpc_out, paths=source_relative).
- [ ] **2.3** Run proto generation; verify `proto/assignment_reference_service/*.pb.go` and `*_grpc.pb.go` exist.

## 3. Migrations (Appendix A)

- [ ] **3.1** Add migrations from Appendix A: `users`, `assignment_groups`, `group_members` (DDL as specified).
- [ ] **3.2** Add read-optimized indexes (e.g. active_only list queries: `is_active`, search fields as needed).
- [ ] **3.3** Seed at least one user and one assignment group per POC requirement; document in README.

## 4. Repository

- [ ] **4.1** `repository/repository.go`: struct with DB (e.g. `*sql.DB`), structured logging.
- [ ] **4.2** User: GetByID, List (pagination, active_only, filter_display_name, filter_username, filter_email), Create, Update; ValidateExists.
- [ ] **4.3** Group: GetByID, List (pagination, active_only, filter_group_name), Create, Update; ValidateExists.
- [ ] **4.4** GroupMember: ListByGroupID (pagination), Add (enforce unique group_id+user_id), Remove.
- [ ] **4.5** All with context propagation, timeouts; no cross-service DB access.

## 5. Service layer

- [ ] **5.1** `service/service.go`: Repository, optional audit publisher (for user.updated, group.updated).
- [ ] **5.2** User: GetUser, ListUsers, CreateUser, UpdateUser, ValidateUserExists; email validation, username uniqueness.
- [ ] **5.3** Group: GetGroup, ListGroups, CreateGroup, UpdateGroup, ValidateGroupExists; group name uniqueness.
- [ ] **5.4** GroupMember: ListGroupMembers, AddGroupMember, RemoveGroupMember; prevent duplicate membership.
- [ ] **5.5** Publish audit events (user.updated, group.updated) after successful commit per platform contract; event envelope format as specified.

## 6. Handlers (gRPC) and API (REST)

- [ ] **6.1** `handler/handler.go`: embed Unimplemented server, hold `*service.Service`.
- [ ] **6.2** `handler/user.go`: GetUser, ListUsers, ValidateUserExists, CreateUser, UpdateUser.
- [ ] **6.3** `handler/group.go`: GetGroup, ListGroups, ListGroupMembers, ValidateGroupExists, CreateGroup, UpdateGroup, AddGroupMember, RemoveGroupMember.
- [ ] **6.4** Map domain errors to gRPC codes (NotFound, InvalidArgument, AlreadyExists, etc.); read metadata (x-user-id, x-request-id, x-correlation-id).
- [ ] **6.5** REST: `api/api.go` + `api/user_handlers.go`, `api/group_handlers.go` for GET /api/v1/reference/users, GET /api/v1/reference/groups, GET /api/v1/reference/groups/{groupId}/members; same Service; platform error shape and status codes.

## 7. main.go, config, logging

- [ ] **7.1** `config/config.go`: DB URL, gRPC port (50051), HTTP port (8080), timeouts from env.
- [ ] **7.2** `logging/logging.go`: structured logging (slog), gRPC LoggingInterceptor; no secrets/PII in logs.
- [ ] **7.3** `cmd/main.go`: load config, init DB (run migrations or document manual step), repository → service → handler + api; start gRPC server (50051), HTTP server (8080) with health (GET /health) and REST routes; SIGTERM/SIGINT graceful shutdown (drain gRPC + HTTP, timeout e.g. 30s); server and outbound timeouts.

## 8. Tests

- [ ] **8.1** Repository tests: persistence, uniqueness, pagination, active_only filters.
- [ ] **8.2** Service tests: validation (email, duplicate username/group name, duplicate membership), query filters, pagination.
- [ ] **8.3** Handler/transport tests: happy path and failure paths per RPC; error code mapping.
- [ ] **8.4** At least one happy path and multiple failure-path tests per write endpoint.

## 9. PRD-specific

- [ ] **9.1** Audit: event publisher interface; publish user.updated / group.updated after Create/Update; envelope format per platform contract (event_id, event_type, source_service, entity_type, entity_id, actor_user_id, request_id, correlation_id, after_data, occurred_at, etc.).
- [ ] **9.2** No auth provider logic; no case-specific logic; no hard-delete of referenced records (deactivation only).

## 10. Deliverables

- [ ] **10.1** Dockerfile (multi-stage, Go 1.23, expose 50051 + 8080).
- [ ] **10.2** .gitlab-ci.yml: build and push image (no Kubernetes).
- [ ] **10.3** README: overview, layout, setup (clone, go mod tidy, proto gen, migrations, run), seed steps, how to run tests, REST vs gRPC.

---

## Execution log

- Plan created from PRD + platform contract.
- Implementation completed: proto, migrations, repository, service, handler, api, main, config, logging, Dockerfile, CI, README, tests.

---

## Review / Results

- **Build:** `go build ./...` succeeds.
- **Tests:** `go test ./...` passes (handler, repository, service, api).
- **Agent Checklist (PRD):** Module and proto set; proto generated. Repository and migrations implemented; service and handler (queries + commands) and optional REST implemented. main.go: gRPC + HTTP, timeouts, graceful shutdown. Tests: query/validation/error mapping/health. README updated.
- **Remaining / follow-up:** Migrations are run manually (psql -f migrations/*.sql). For production, wire a real AuditPublisher (e.g. SNS/SQS) in main. Optional: readiness probe that checks DB Ping on /health or /ready.
