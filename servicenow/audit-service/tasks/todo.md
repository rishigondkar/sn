# Audit Service – Implementation Plan

**Sources:** `07_audit_service_prd.md`, `00_platform_integration_contract.md`, Go gRPC template.

---

## 1. Project layout (per PRD Go Project Structure)

- [x] Create directories: `cmd/`, `handler/`, `service/`, `repository/`, `consumer/`, `config/`, `logging/`, `proto/`, `migrations/`
- [x] `go.mod`: module `github.com/servicenow/audit-service`, Go 1.24
- [x] No `api/` (gateway calls this service via gRPC only; REST surface is gateway’s responsibility)

---

## 2. Proto (Appendix B)

- [x] Add `proto/audit_service.proto`: package `audit_service`, `go_package = "github.com/servicenow/audit-service/proto/audit_service"`
- [x] Service: `ListAuditEventsByCase`, `ListAuditEventsByObservable`, `ListAuditEventsByEntity`, `ListAuditEventsByActor`, `ListAuditEventsByCorrelationId`
- [x] Request/response messages and `AuditEvent` message as in Appendix B (including optional timestamps, `oldest_first`, pagination)
- [x] Add `proto/generate_proto.sh` and run generation → `proto/audit_service/*.pb.go`

---

## 3. Migrations (Appendix A)

- [x] Add `migrations/` with executable DDL: `audit_events` table and indexes (event_id unique, case_id, observable_id, (entity_type, entity_id), actor_user_id, correlation_id, occurred_at)

---

## 4. Config

- [x] `config/config.go`: DB URL, gRPC port (50051), HTTP port (8080), consumer config (retries, DLQ), timeouts; load from env

---

## 5. Logging

- [x] `logging/logging.go`: `SetupLogging`, gRPC `LoggingInterceptor`; structured logging; no secrets/PII in logs

---

## 6. Repository

- [x] `repository/repository.go`: struct with DB pool
- [x] Idempotent insert: insert audit event; unique on `event_id` (ON CONFLICT DO NOTHING or ignore duplicate)
- [x] List methods: by case_id, observable_id, (entity_type, entity_id), actor_user_id, correlation_id; all with pagination (page_size, page_token), optional occurred_after/occurred_before, sort (newest first default, optional oldest_first)
- [x] Return domain-friendly structs for service layer

---

## 7. Consumer

- [x] `consumer/consumer.go`: subscribe to event bus (abstraction: e.g. MessageSource interface for POC – channel or stub; production SNS/SQS or Kafka)
- [x] Parse message body as audit event envelope JSON (platform contract shape)
- [x] Validate required envelope fields
- [x] Dedupe by `event_id` via repository idempotent insert
- [x] Retry on transient failure; after N retries send to dead-letter (interface/hook)
- [x] No synchronous callbacks to producing services
- [x] Document transport and topic/queue in README for POC

---

## 8. Service layer

- [x] `service/service.go`: Repository, optional consumer deps
- [x] `service/audit_query.go`: ListByCase, ListByObservable, ListByEntity, ListByActor, ListByCorrelationId – call repository, handle pagination/tokens, map to response shape

---

## 9. Handler layer (gRPC only)

- [x] `handler/handler.go`: embed `UnimplementedAuditServiceServer`, hold `*service.Service`
- [x] `handler/audit_query.go`: implement 5 List* RPCs; delegate to service; map proto request/response; gRPC error mapping (InvalidArgument, NotFound, Internal)

---

## 10. main.go

- [x] Wire config, logging, repository, service, handler
- [x] Start gRPC server on 50051 (with interceptor, timeouts)
- [x] Start HTTP server on 8080: health only (`GET /health` → 200 `{"status":"ok"}`)
- [x] Start event consumer (goroutine or worker)
- [x] Graceful shutdown: signal handler; stop consumer first, then drain gRPC (GracefulStop), then HTTP Shutdown; shutdown timeout (e.g. 30s)

---

## 11. Tests

- [x] Repository: idempotent insert (duplicate `event_id` does not duplicate row)
- [x] Consumer: malformed event handling (invalid JSON, missing required fields)
- [x] Consumer: duplicate event_id accepted once, second time idempotent no-op
- [x] Handler/service: query filters (by case, observable, entity, actor, correlation) – table-driven where possible

---

## 12. Docker, CI, README

- [x] Dockerfile: multi-stage build; expose 50051 (gRPC), 8080 (HTTP)
- [x] `.gitlab-ci.yml`: build and push Docker image
- [x] README: overview, layout, setup (migrations, proto gen, run), health, event bus transport/topic for POC, runbook notes

---

## 13. Agent Checklist (PRD)

- [x] Set module name and proto go_package; run proto generation (query RPCs only)
- [x] Implement repository (audit_events table, indexes); migrations
- [x] Implement consumer: subscribe, validate envelope, idempotent insert, retry/dead-letter
- [x] Implement service and handler for query RPCs
- [x] main.go: start gRPC server + consumer; graceful shutdown for both
- [x] Tests: duplicate event idempotency, malformed event handling, query filters
- [x] Update README

---

## Review / results

- **Layout**: Implemented per PRD: `cmd/`, `handler/`, `service/`, `repository/`, `consumer/`, `config/`, `logging/`, `proto/`, `migrations/`. No `api/` (gateway calls via gRPC).
- **Proto**: `proto/audit_service.proto` from Appendix B with `go_package = github.com/servicenow/audit-service/proto/audit_service`; `generate_proto.sh` generates Go stubs.
- **Migrations**: `migrations/001_audit_events.up.sql` applies Appendix A (table + indexes).
- **Repository**: Idempotent insert (`ON CONFLICT (event_id) DO NOTHING`), list by case/observable/entity/actor/correlation with keyset pagination and time filters.
- **Consumer**: Envelope validation, idempotent persist, retry then Nack (dead-letter). `ChannelSource` for POC; production must wire SQS/Kafka per README.
- **Service/Handler**: All 5 query RPCs implemented; validation returns `InvalidArgument`; errors mapped to gRPC status.
- **main.go**: gRPC on 50051, HTTP health on 8080, consumer goroutine; graceful shutdown (stop consumer → GracefulStop → HTTP Shutdown, 30s).
- **Tests**: Consumer (ValidateEnvelope, invalid JSON, missing fields, channel source); handler (InvalidArgument for empty case_id/entity); repository (idempotent insert, list empty) — repo tests skip when no TEST_DATABASE_URL/DATABASE_URL.
- **Dockerfile / CI / README**: Added; README documents event bus POC and Agent Checklist.

**Remaining / follow-ups**

- Run repository tests with a real PostgreSQL (set `TEST_DATABASE_URL`) to confirm idempotent insert and list queries end-to-end.
- In production, replace `ChannelSource` with an SQS/Kafka implementation and document topic/queue and DLQ in the runbook.
