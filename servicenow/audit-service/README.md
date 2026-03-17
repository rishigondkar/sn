# Audit Service

Centralized, append-only audit trail for the security operations platform. Consumes audit events from the event bus (pub/sub), persists them idempotently, and exposes gRPC query APIs for filtering and timeline retrieval.

## Features

- **Event consumption**: Subscribes to the platform audit event bus; validates the shared audit envelope; deduplicates by `event_id`; retries on failure with configurable dead-letter behavior.
- **gRPC query APIs** (service-to-service, port 50051):
  - `ListAuditEventsByCase(case_id, pagination, time range, sort)`
  - `ListAuditEventsByObservable(observable_id, ...)`
  - `ListAuditEventsByEntity(entity_type, entity_id, ...)`
  - `ListAuditEventsByActor(actor_user_id, ...)`
  - `ListAuditEventsByCorrelationId(correlation_id, ...)`
- **Health**: `GET /health` on HTTP port 8080; returns 200 when up and DB is reachable, 503 if DB is unavailable.
- **Graceful shutdown**: On SIGTERM/SIGINT, stops the consumer, drains gRPC and HTTP, then exits (30s timeout).

REST surface for the UI is provided by the API Gateway (e.g. `GET /api/v1/cases/{caseId}/audit-events`); the gateway calls this service via gRPC.

## Repository layout

```
├── cmd/main.go
├── handler/           # gRPC query handlers
├── service/           # Business logic
├── repository/        # audit_events persistence
├── consumer/          # Event bus subscriber, validation, idempotent insert
├── config/
├── logging/
├── proto/             # audit_service.proto + generated Go
├── migrations/        # PostgreSQL DDL (001_audit_events.up.sql)
├── go.mod, go.sum
├── Dockerfile
├── .gitlab-ci.yml
└── README.md
```

## Prerequisites

- Go 1.24+
- PostgreSQL (for `audit_events` table)
- `protoc` with Go plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) for proto generation

## Setup

1. Clone and install dependencies:
   ```bash
   go mod download
   ```

2. Generate gRPC stubs:
   ```bash
   cd proto && ./generate_proto.sh audit_service go && cd ..
   ```

3. Run migrations (create `audit_events` and indexes):
   ```bash
   psql "$DATABASE_URL" -f migrations/001_audit_events.up.sql
   ```

4. Configure (environment):
   - `DATABASE_URL` – PostgreSQL connection string (default: `postgres://localhost/audit?sslmode=disable`)
   - `GRPC_PORT` – gRPC server port (default: 50051)
   - `HTTP_PORT` – HTTP server port for health (default: 8080)
   - `CONSUMER_MAX_RETRIES`, `CONSUMER_RETRY_DELAY`, `CONSUMER_BATCH_SIZE` – consumer behavior

5. Run the service:
   ```bash
   go run ./cmd
   ```
   Or build and run:
   ```bash
   go build -o audit-service ./cmd && ./audit-service
   ```

## Event bus (POC)

The service expects a **MessageSource** that delivers audit event messages. The message body must be the **audit event envelope JSON** defined in the platform contract (no wrapper).

- **POC / testing**: A channel-based source is included (`consumer.ChannelSource`); production should wire a real transport.
- **Production**: Use a single topic/queue (e.g. `soc-audit-events`) on SNS/SQS, Kafka, or another at-least-once pub/sub system. Document the chosen transport and topic/queue name in the deployment runbook. The consumer is idempotent by `event_id`; failed messages should go to a dead-letter queue after the configured retries.

## Tests

- Unit tests (no DB): `go test ./consumer/... ./handler/...`
- Repository tests (require PostgreSQL): set `TEST_DATABASE_URL` or `DATABASE_URL`, then `go test ./repository/...`

Run all (repository tests skip if no DB):

```bash
go test ./...
```

## Ports

| Port  | Purpose        |
|-------|----------------|
| 50051 | gRPC (queries) |
| 8080  | HTTP (health)  |

## Agent checklist (from PRD)

- [x] Set module name and proto go_package; run proto generation (query RPCs only).
- [x] Implement repository (audit_events table, indexes); migrations.
- [x] Implement consumer: subscribe, validate envelope, idempotent insert, retry/dead-letter.
- [x] Implement service and handler for query RPCs.
- [x] main.go: start gRPC server + consumer; graceful shutdown for both.
- [x] Tests: duplicate event idempotency, malformed event handling, query filters.
- [x] Update README.
