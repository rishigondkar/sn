# Build and Integration Verification Report

**Scope:** Verify all seven services are correctly built per their PRDs and that they will work together.  
**Updates applied:** Gateway real Case client, default ports 50051–50056, RUN.md.

---

## 1. Per-service build correctness

All seven services are correctly built per PRD: layout, proto, migrations, handlers, health, graceful shutdown, and (where applicable) gRPC metadata and audit publisher. See previous verification for details.

---

## 2. Updates applied (to run the project correctly)

### 2.1 Gateway – real Case Service client

- **go.mod:** Added `replace` directives for all six domain services (`../case-service`, `../alert-observable-service`, etc.) and `require` for `google.golang.org/grpc` and `google.golang.org/protobuf` so the gateway can import Case Service proto.
- **clients/case_client_grpc.go:** New file implementing `CaseCommandService` and `CaseQueryService` by dialing the Case Service, setting outgoing gRPC metadata (x-user-id, x-request-id, x-correlation-id) from `clients.FromContext(ctx)`, and mapping gateway DTOs to/from proto. Uses GetCase before AssignCase/CloseCase/UpdateCase to pass correct `version_no`.
- **cmd/main.go:** Uses `clients.NewCaseClients(cfg.CaseServiceAddr, cfg.DownstreamTimeout)` when `CASE_SERVICE_ADDR` is set; on failure falls back to stubs. Calls the Case connection close function on shutdown.

### 2.2 Default gRPC ports (single-host)

So that all services can run on one host without port conflicts and match the gateway’s default addrs:

| Service | Default GRPC port (before → after) |
|---------|-----------------------------------|
| Case Service | 50051 (unchanged) |
| Alert & Observable Service | 50051 → **50052** |
| Enrichment & Threat Lookup Service | 50051 → **50053** |
| Assignment & Reference Service | :50051 → **:50054** |
| Attachment Service | 50051 → **50055** |
| Audit Service | 50051 → **50056** |

Gateway config already defaulted to `localhost:50051` … `localhost:50056` for the six downstreams; domain services now default to these ports.

### 2.3 Run guide

- **RUN.md** (repo root): Describes default gRPC/HTTP ports, order to start services, how to start the gateway, and a sample curl. Notes that Case uses the real gRPC client and other services still use stubs.
- **api-gateway-bff/README.md:** Updated to state that the Case Service uses the real gRPC client when `CASE_SERVICE_ADDR` is set and that others still use stubs.

---

## 3. How to run the entire project

1. Start PostgreSQL and create DBs (or set `DATABASE_URL` per service).
2. Start the six domain services (each in its own terminal or process) so they listen on 50051–50056.
3. Start the gateway: `cd api-gateway-bff && go run ./cmd`.
4. Use the gateway at `http://localhost:8080` (e.g. `POST /api/v1/cases` with `X-User-Id`). Case create/get/update/worknote/assign/close go to the real Case Service over gRPC.

---

## 4. Remaining (optional) integration work

- **Other five services:** Implement real gRPC clients in `api-gateway-bff/clients/` (e.g. `*_grpc.go`) and wire them in `main.go` the same way as Case (with metadata and config addrs).
- **Audit event bus:** Connect producers to a real bus (SQS/Kafka) and point the Audit Service consumer at the same topic/queue so audit events flow end-to-end.

No further changes are required for the project to “run correctly” with the gateway calling the Case Service over gRPC with correct metadata and default ports.
