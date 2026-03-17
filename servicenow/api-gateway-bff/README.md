# API Gateway / BFF

Public REST entry point and response aggregation for the SOC platform UI and approved external clients.

## Responsibilities

- Terminate public REST requests under `/api/v1`
- Extract and propagate auth context (X-User-Id, X-Request-Id, X-Correlation-Id)
- Validate request structure and normalize errors to the shared REST envelope
- Route requests to internal services via gRPC client interfaces
- Aggregate case detail (GET `/api/v1/cases/{caseId}/detail`) with partial-failure tolerance

## Project structure

```
cmd/main.go           # HTTP server, config, logging, graceful shutdown
api/                  # Router, middleware, REST handlers under /api/v1
orchestrator/         # Client interfaces and response composition (e.g. case detail)
clients/              # gRPC client interfaces and stub implementations
config/               # Configuration from environment
logging/              # Structured logging
```

- **No gRPC server** in this service.
- **No repository or proto** in this repo; the gateway consumes other services’ gRPC APIs.

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness (200 when up) |
| POST | `/api/v1/cases` | Create case |
| GET | `/api/v1/cases/{caseId}` | Get case |
| PATCH | `/api/v1/cases/{caseId}` | Update case |
| POST | `/api/v1/cases/{caseId}/worknotes` | Add worknote |
| GET | `/api/v1/cases/{caseId}/worknotes` | List worknotes |
| POST | `/api/v1/cases/{caseId}/assign` | Assign case |
| POST | `/api/v1/cases/{caseId}/close` | Close case |
| POST | `/api/v1/cases/{caseId}/observables` | Link observable |
| GET | `/api/v1/cases/{caseId}/observables` | List observables |
| GET | `/api/v1/cases/{caseId}/alerts` | List alerts |
| GET | `/api/v1/cases/{caseId}/enrichment-results` | List enrichment results |
| GET | `/api/v1/cases/{caseId}/attachments` | List attachments |
| GET | `/api/v1/cases/{caseId}/audit-events` | List audit events |
| GET | `/api/v1/cases/{caseId}/detail` | Aggregated case detail |
| GET | `/api/v1/observables/{observableId}/threat-lookups` | Threat lookups for observable |
| POST | `/api/v1/attachments` | Create attachment (metadata) |
| GET | `/api/v1/reference/users` | List users |
| GET | `/api/v1/reference/groups` | List groups |

## Configuration (environment)

| Variable | Default | Description |
|----------|---------|-------------|
| HTTP_PORT | 8080 | HTTP listen port |
| SHUTDOWN_TIMEOUT | 30s | Graceful shutdown timeout |
| DOWNSTREAM_TIMEOUT | 15s | Timeout for downstream gRPC calls |
| CASE_SERVICE_ADDR | localhost:50051 | Case Service gRPC address |
| OBSERVABLE_SERVICE_ADDR | localhost:50052 | Alert & Observable Service |
| ENRICHMENT_SERVICE_ADDR | localhost:50053 | Enrichment & Threat Lookup Service |
| REFERENCE_SERVICE_ADDR | localhost:50054 | Assignment & Reference Service |
| ATTACHMENT_SERVICE_ADDR | localhost:50055 | Attachment Service |
| AUDIT_SERVICE_ADDR | localhost:50056 | Audit Service |
| CORS_ALLOWED_ORIGINS | * | Comma-separated origins |
| CORS_ALLOWED_METHODS | GET,POST,... | Comma-separated methods |

## Running

```bash
go run ./cmd
# or
go build -o api-gateway ./cmd && ./api-gateway
```

Server listens on `:8080` by default. Send SIGTERM or SIGINT for graceful shutdown.

## Downstream clients

- **Case Service:** When `CASE_SERVICE_ADDR` is set (default `localhost:50051`), the gateway uses a **real gRPC client** (`clients.NewCaseClients`) that dials the Case Service and forwards x-user-id, x-request-id, x-correlation-id in outgoing metadata. If the Case Service is unreachable at startup, the gateway falls back to stub implementations.
- **Other services:** The `clients/` package defines **interfaces** and **stub implementations** so that:

- The gateway builds and runs without those services (stubs return empty or placeholder data).
- Tests can use mocks that implement the same interfaces.

**To wire real gRPC clients:**

1. Add the proto definitions (or generated Go stubs) for each downstream service, e.g. from a shared proto module or by copying generated code into this repo.
2. Implement the interfaces in `clients/` (e.g. `CaseQueryService`, `CaseCommandService`) with types that dial the service and call the generated client, mapping responses to the DTOs in `clients/types.go`.
3. Replace the stub constructors in `cmd/main.go` with constructors that dial and return the real implementations.

Regeneration of client stubs should follow the process defined for the shared proto repo or the individual service repos (see their README or Appendix B in each service PRD).

## Tests

```bash
go test ./...
```

- `api/`: route handler tests, validation, error responses.
- `api/middleware_test.go`: request ID, correlation ID, auth context.
- `orchestrator/`: case detail aggregation and partial-failure behavior.

## References

- **Platform contract:** `00_platform_integration_contract.md`
- **Service PRD:** `01_api_gateway_bff_prd.md`
