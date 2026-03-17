# Plan: Use Real Services Only (No Stubs)

This plan removes all stubs from the API Gateway BFF and wires real gRPC clients to every downstream service. The gateway will **fail to start** if any required service is unreachable.

---

## Current State

| Client / Interface | Current | Config Env | Backend Service | Proto |
|--------------------|---------|------------|-----------------|-------|
| Case (Command + Query) | Real when `CASE_SERVICE_ADDR` set; else stub | `CASE_SERVICE_ADDR` | case-service | case_service.proto |
| Observable (Command + Query) | Stub | `OBSERVABLE_SERVICE_ADDR` | alert-observable-service | alert_observable_service.proto |
| Enrichment | Stub | `ENRICHMENT_SERVICE_ADDR` | enrichment-threat-service | enrichment_threat_service.proto |
| ThreatLookup | Stub | (same as Enrichment) | enrichment-threat-service | enrichment_threat_service.proto |
| Reference | Stub | `REFERENCE_SERVICE_ADDR` | assignment-reference-service | assignment_reference_service.proto |
| Attachment (Query + Command) | Stub | `ATTACHMENT_SERVICE_ADDR` | attachment-service | attachment_service.proto |
| Audit | Stub | `AUDIT_SERVICE_ADDR` | audit-service | audit_service.proto |

---

## 1. Case Service

- **Status:** Real client exists (`case_client_grpc.go`).
- **Change:** In `cmd/main.go`, **require** `CASE_SERVICE_ADDR` and use real client only. If dial fails or env empty → exit with clear error (no stub fallback).

---

## 2. Assignment Reference Service

- **Backend:** `assignment-reference-service` (gRPC, e.g. port 50051 per its README; config uses 50054).
- **Proto:** `assignment_reference_service.proto` → `GetUser`, `ListUsers`, `GetGroup`, `ListGroups`.
- **Gateway interface:** `ReferenceQueryService` — `GetUser`, `GetGroup`, `ListUsers`, `ListGroups`.

**Tasks:**

1. Add dependency in `api-gateway-bff/go.mod`:
   - `require github.com/servicenow/assignment-reference-service v0.0.0`
   - `replace github.com/servicenow/assignment-reference-service => ../assignment-reference-service`
2. Create `clients/reference_client_grpc.go`:
   - Dial using `cfg.ReferenceServiceAddr`, same pattern as case (insecure credentials, shared conn).
   - Implement `ReferenceQueryService`: call proto `GetUser`/`GetGroup`/`ListUsers`/`ListGroups`, map proto messages to `clients.User` and `clients.Group`.
   - Use `withMetadata(ctx)` for x-user-id, x-request-id, x-correlation-id.
3. In `main.go`: require `REFERENCE_SERVICE_ADDR`, create reference client via new constructor, pass to orchestrator. On dial failure → log error and exit.
4. Track `closeRefConn func()` and call on shutdown.

---

## 3. Alert Observable Service

- **Backend:** `alert-observable-service` (gRPC).
- **Proto:** `alert_observable_service.proto` — discover RPCs for: link observable to case, list alerts by case, list observables by case, list child observables, list similar incidents.
- **Gateway interfaces:**
   - `ObservableCommandService`: `LinkObservableToCase(ctx, caseID, observableID)`.
   - `ObservableQueryService`: `ListCaseAlerts`, `ListCaseObservables`, `ListChildObservables`, `ListSimilarIncidents`.

**Tasks:**

1. Add dependency: `github.com/org/alert-observable-service` (replace with `../alert-observable-service`).
2. Create `clients/observable_client_grpc.go`:
   - Dial `cfg.ObservableServiceAddr`.
   - Map proto request/response types to gateway DTOs (`Alert`, `Observable`, `ChildObservable`, `SimilarIncident`).
   - Implement both interfaces; if the backend splits command/query, use same connection, two wrapper structs.
3. In `main.go`: require `OBSERVABLE_SERVICE_ADDR`, create observable clients, pass to orchestrator. On failure → exit.
4. Shutdown: close observable gRPC connection.

**Note:** Today case-service `LinkObservable` is used from `observable_handlers.go`. Confirm whether that should stay (case-service) or move to alert-observable-service; if the latter, implement `LinkObservableToCase` in the new gRPC client and keep handler calling `Orch.ObsCmd`.

---

## 4. Enrichment & Threat Lookup Service

- **Backend:** `enrichment-threat-service` (single service, two logical interfaces in gateway).
- **Proto:** `enrichment_threat_service.proto` — list enrichment results by case, list threat lookups by case / by observable.
- **Gateway interfaces:**
   - `EnrichmentQueryService`: `ListEnrichmentResultsByCase(ctx, caseID, pageSize, pageToken)`.
   - `ThreatLookupQueryService`: `ListThreatLookupResultsByCase`, `ListThreatLookupsByObservable`.

**Tasks:**

1. Add dependency: `github.com/servicenow/enrichment-threat-service` (replace with `../enrichment-threat-service`).
2. Create `clients/enrichment_client_grpc.go`:
   - Single dial to `cfg.EnrichmentServiceAddr`.
   - Two structs implementing `EnrichmentQueryService` and `ThreatLookupQueryService`, both wrapping the same gRPC client.
   - Map proto to `EnrichmentResult`, `ThreatLookupResult`.
3. In `main.go`: require `ENRICHMENT_SERVICE_ADDR`, create both clients, pass to orchestrator. On failure → exit.
4. Shutdown: close enrichment gRPC connection.

---

## 5. Attachment Service

- **Backend:** `attachment-service` (gRPC).
- **Proto:** `attachment_service.proto` — list attachments by case, create attachment (or equivalent).
- **Gateway interfaces:**
   - `AttachmentQueryService`: `ListAttachmentsByCase(ctx, caseID, pageSize, pageToken)`.
   - `AttachmentCommandService`: `CreateAttachment(ctx, *CreateAttachmentRequest)`.

**Tasks:**

1. Add dependency: `github.com/soc-platform/attachment-service` (replace with `../attachment-service`).
2. Create `clients/attachment_client_grpc.go`:
   - Dial `cfg.AttachmentServiceAddr`.
   - Implement list and create; map proto to `clients.Attachment` and `CreateAttachmentRequest`/response.
3. In `main.go`: require `ATTACHMENT_SERVICE_ADDR`, create attachment clients, pass to orchestrator. On failure → exit.
4. Shutdown: close attachment gRPC connection.

---

## 6. Audit Service

- **Backend:** `audit-service` (gRPC).
- **Proto:** `audit_service.proto` — list audit events by case (or equivalent).
- **Gateway interface:** `AuditQueryService`: `ListAuditEventsByCase(ctx, caseID, pageSize, pageToken)`.

**Tasks:**

1. Add dependency: `github.com/servicenow/audit-service` (replace with `../audit-service`).
2. Create `clients/audit_client_grpc.go`:
   - Dial `cfg.AuditServiceAddr`.
   - Call list RPC, map proto to `clients.AuditEvent`.
3. In `main.go`: require `AUDIT_SERVICE_ADDR`, create audit client, pass to orchestrator. On failure → exit.
4. Shutdown: close audit gRPC connection.

---

## 7. Main and Config

**Startup (no stubs):**

- Require all seven addrs to be non-empty: `CASE_SERVICE_ADDR`, `REFERENCE_SERVICE_ADDR`, `OBSERVABLE_SERVICE_ADDR`, `ENRICHMENT_SERVICE_ADDR`, `ATTACHMENT_SERVICE_ADDR`, `AUDIT_SERVICE_ADDR`.
- Optionally allow a single shared “service discovery” or default base (e.g. `SERVICES_HOST`) to build addrs if you prefer not to set six env vars by hand.
- For each service: dial at startup; on any dial failure, log which service failed and exit with non-zero code.
- Remove all `New*Stub()` calls and stub fallbacks.

**Shutdown:**

- Collect all `closeFn` from each real client and call them in reverse order (or in parallel) during graceful shutdown.

**Config:**

- Keep existing env vars in `config/config.go`. Consider validating in `Load()` that all required addrs are non-empty when running in “real-only” mode.

---

## 8. Dependencies and Proto Compatibility

- Each backend may use a different Go module path and proto `go_package`. The gateway will:
  - Add a `require` and local `replace` for each service in `go.mod`.
  - Import generated Go code from that service’s proto package (e.g. `pb "github.com/servicenow/assignment-reference-service/proto/assignment_reference_service"`).
- Ensure gateway’s `google.golang.org/grpc` and `google.golang.org/protobuf` versions are compatible with those used by each service to avoid duplicate or conflicting proto/grpc registrations.

---

## 9. Remove Stubs

After all real clients are implemented and wired:

- Delete stub files:
  - `clients/case_client_stub.go`
  - `clients/reference_client_stub.go`
  - `clients/observable_client_stub.go`
  - `clients/enrichment_client_stub.go`
  - `clients/attachment_client_stub.go`
  - `clients/audit_client_stub.go`
- Remove any test-only references to stubs; tests can use mocks or real in-process servers if needed.

---

## 10. Order of Implementation (Suggested)

1. **Case** — Enforce required addr and no stub fallback.
2. **Reference** — Small surface (4 methods), unblocks assignment group/user dropdowns.
3. **Audit** — Single method, good next step.
4. **Attachment** — Two interfaces, list + create.
5. **Enrichment + ThreatLookup** — One service, two interfaces.
6. **Observable** — Most RPCs; confirm LinkObservable ownership (case vs alert-observable).
7. **Remove all stubs** and add startup validation for required addrs.

---

## 11. Running the Stack

With real services only, you will need:

- All six backend services running and listening on the ports specified by the env vars.
- Their databases migrated and (for reference) seeded.
- Gateway started with all `*_SERVICE_ADDR` set; e.g.:

  ```bash
  export CASE_SERVICE_ADDR=localhost:50051
  export REFERENCE_SERVICE_ADDR=localhost:50054
  export OBSERVABLE_SERVICE_ADDR=localhost:50052
  export ENRICHMENT_SERVICE_ADDR=localhost:50053
  export ATTACHMENT_SERVICE_ADDR=localhost:50055
  export AUDIT_SERVICE_ADDR=localhost:50056
  ./api-gateway-bff
  ```

Document these in the gateway README and/or a top-level `docker-compose` or run script so the full stack can be started with one command.
