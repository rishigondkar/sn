# Enrichment & Threat Lookup Service

Passive persistence and retrieval of enrichment results and threat lookup results for the SOC platform. This service does not call external enrichment or threat-intel tools; it accepts data via gRPC and REST and returns stored results.

## Features

- **gRPC** (port 50051): service-to-service API
  - `UpsertEnrichmentResult` / `UpsertThreatLookupResult`
  - `ListEnrichmentResultsByCase` / `ListEnrichmentResultsByObservable`
  - `ListThreatLookupResultsByCase` / `ListThreatLookupResultsByObservable`
  - `GetThreatLookupSummaryByObservable`
- **REST** (port 8080): for external clients and UI
  - `POST/PUT /api/v1/enrichment-results`, `GET /api/v1/cases/{caseId}/enrichment-results`, `GET /api/v1/observables/{observableId}/enrichment-results`
  - `POST/PUT /api/v1/threat-lookups`, `GET /api/v1/cases/{caseId}/threat-lookups`, `GET /api/v1/observables/{observableId}/threat-lookups`
- **Health**: `GET /health` → `{"status":"ok"}`
- Idempotent upserts (dedupe by natural key); audit events published for upserts.

## Layout

```
├── cmd/main.go
├── handler/          # gRPC handlers
├── api/              # REST handlers
├── service/          # Business logic
├── repository/       # PostgreSQL access
├── config/
├── logging/
├── proto/            # Proto definition and generated Go
├── migrations/       # SQL migrations (Appendix A + dedupe indexes)
├── go.mod, go.sum
├── Dockerfile
└── README.md
```

## Prerequisites

- Go 1.23+
- PostgreSQL (for migrations and runtime)
- `protoc` and Go plugins (`protoc-gen-go`, `protoc-gen-go-grpc`) for proto generation

## Setup

1. Clone and enter the repo:
   ```bash
   cd enrichment-threat-service
   ```
2. Install dependencies and generate proto:
   ```bash
   go mod tidy
   cd proto && ./generate_proto.sh enrichment_threat_service go && cd ..
   ```
3. Run migrations (in order):
   - `migrations/001_initial_schema.up.sql`
   - `migrations/002_dedupe_indexes.up.sql`
   Use your migration runner (e.g. [golang-migrate](https://github.com/golang-migrate/migrate)) or run the SQL manually against your database.
4. Set environment (optional):
   - `DATABASE_URL` – PostgreSQL connection string (default: `postgres://localhost:5432/enrichment_threat?sslmode=disable`)
   - `GRPC_PORT` – default 50051
   - `HTTP_PORT` – default 8080
   - `MAX_PAYLOAD_BYTES` – max request body size (default 1MB)

## Run

```bash
go run ./cmd
```

Or run the built binary:

```bash
go build -o server ./cmd && ./server
```

## Tests

```bash
go test ./...
```

- Service: validation and error paths for upserts
- API: health endpoint
- Handler: gRPC error mapping (e.g. InvalidArgument for validation failures)

Repository and full integration tests require a running PostgreSQL instance; run migrations first and set `DATABASE_URL` for integration tests if added.

## gRPC vs REST

- **gRPC**: used by API Gateway and other backend services. Port 50051. Identity and tracing via metadata: `x-user-id`, `x-request-id`, `x-correlation-id`.
- **REST**: used by external clients and UI. Port 8080. Base path `/api/v1`. Error shape follows platform contract (`error.code`, `error.message`, `requestId`, `correlationId`).

## Deployment

Build the Docker image:

```bash
docker build -t enrichment-threat-service .
```

Run with `DATABASE_URL` and ports 50051 (gRPC) and 8080 (HTTP) exposed.
