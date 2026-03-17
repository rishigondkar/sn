# Assignment Reference Service

Source of truth for users, assignment groups, and group membership for the SOC platform. Exposes gRPC (service-to-service) and REST (UI/gateway) for queries and optional write administration.

## Features

- **gRPC** (port 50051): GetUser, ListUsers, GetGroup, ListGroups, ListGroupMembers, ValidateUserExists, ValidateGroupExists; CreateUser, UpdateUser, CreateGroup, UpdateGroup, AddGroupMember, RemoveGroupMember.
- **REST** (port 8080): `GET /api/v1/reference/users`, `GET /api/v1/reference/groups`, `GET /api/v1/reference/groups/{groupId}/members` (pagination, active_only, filters).
- **Health**: `GET /health` returns `{"status":"ok"}`.

## Layout

```
├── cmd/main.go
├── handler/          # gRPC handlers
├── api/              # REST handlers
├── service/          # Business logic
├── repository/       # Data access
├── config/
├── logging/
├── proto/
├── migrations/
├── go.mod
├── Dockerfile
└── README.md
```

## Prerequisites

- Go 1.22+
- PostgreSQL
- `protoc` and Go plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) for proto generation

## Setup

1. Clone and enter the repo.
2. `go mod tidy`
3. Generate proto: from repo root, `./proto/generate_proto.sh assignment_reference_service go` (ensure `$GOPATH/bin` or plugin path is in `PATH`).
4. Create a PostgreSQL database and run migrations in order:
   - `psql -f migrations/001_initial_schema.sql`
   - `psql -f migrations/002_seed_poc.sql`
5. Set `DATABASE_URL` (e.g. `postgres://user:pass@localhost:5432/dbname?sslmode=disable`). Optional: `GRPC_ADDR=:50051`, `HTTP_ADDR=:8080`.

## Run

```bash
go run ./cmd
```

gRPC on `:50051`, HTTP (REST + health) on `:8080`.

## Seed data (POC)

After running `002_seed_poc.sql` you get:

- User: `soc-admin` (id `a0000000-0000-4000-8000-000000000001`).
- Group: `Triage` (id `b0000000-0000-4000-8000-000000000001`).
- One group member linking the seed user to the seed group.

Case and Alert services can reference these IDs.

## Tests

```bash
go test ./...
```

Unit tests for repository, service, and handler; no running server required for unit tests. For gRPC handler tests that call the server, start the server separately or use in-process testing as needed.

## REST examples

- List users (active only, page size 10): `GET /api/v1/reference/users?page_size=10`
- List groups: `GET /api/v1/reference/groups`
- List members of a group: `GET /api/v1/reference/groups/{groupId}/members`

## gRPC

Use the generated client from `proto/assignment_reference_service`. Pass actor and tracing via metadata: `x-user-id`, `x-request-id`, `x-correlation-id`.
