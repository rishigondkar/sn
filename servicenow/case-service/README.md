# Case Service

Source of truth for case lifecycle and worknotes in the security operations platform. Owns case creation, updates, assignment, state transitions, closure, worknotes, and active duration tracking.

## Features

- **gRPC** (port 50051): service-to-service API
  - Commands: CreateCase, UpdateCase, AssignCase, CloseCase, AddWorknote
  - Queries: GetCase, ListCases, ListWorknotes
- **REST** (port 8080): internal/admin API under `/api/v1/`
  - `POST /api/v1/cases`, `GET /api/v1/cases`, `GET /api/v1/cases/{id}`, `PATCH /api/v1/cases/{id}`
  - `POST /api/v1/cases/{id}/assign`, `POST /api/v1/cases/{id}/close`
  - `POST /api/v1/cases/{id}/worknotes`, `GET /api/v1/cases/{id}/worknotes`
- **Health**: `GET /health` returns 200 with `{"status":"ok"}`

## Layout

```
├── cmd/main.go
├── handler/          # gRPC handlers
├── api/              # REST handlers
├── service/          # Business logic
├── repository/       # DB access
├── domain/           # Types and validation
├── audit/            # Event envelope and publisher
├── config/
├── logging/
├── proto/
│   ├── case_service.proto
│   └── case_service/  # Generated
├── migrations/
└── tasks/
```

## Prerequisites

- Go 1.23+
- PostgreSQL (for migrations and runtime)
- `protoc` and Go plugins for proto generation (see below)

## Setup

1. Clone and enter the repo:
   ```bash
   cd case-service
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Generate gRPC stubs:
   ```bash
   ./proto/generate_proto.sh case_service go
   ```
   Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on PATH (e.g. `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` and `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`).

4. Run migrations (PostgreSQL):
   - Apply `migrations/000001_cases.up.sql`, `000002_case_worknotes.up.sql`, `000003_case_number_sequence.up.sql` in order (e.g. with `psql` or [golang-migrate](https://github.com/golang-migrate/migrate)).

5. Run the service:
   ```bash
   export DATABASE_URL="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
   go run ./cmd
   ```
   Or build and run:
   ```bash
   go build -o case-service ./cmd
   ./case-service
   ```

## Configuration

| Env | Description | Default |
|-----|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | required |
| `GRPC_PORT` | gRPC server port | 50051 |
| `HTTP_PORT` | HTTP server port | 8080 |
| `AUDIT_EVENT_TOPIC` | Audit event bus topic (optional) | - |

## Tests

```bash
go test ./...
```

Tests include domain validation, service logic (with mocked repo/publisher), and handler error mapping. Repository and integration tests require a test database when added.

## Audit events

The service publishes audit events after successful writes (case.created, case.updated, case.assigned, case.closed, worknote.created). Event bus integration is via the `audit.Publisher` interface; a no-op implementation is used when no topic is configured.

## License

Internal use.
