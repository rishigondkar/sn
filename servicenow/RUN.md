# Running the platform

To run the entire project with the gateway calling real services, start each service with the **default gRPC ports** below so the gateway can reach them on one host.

## Quick checklist (you’re good when)

- [x] PostgreSQL running (user `cursor`, password `qwerty`)
- [x] Six databases created (`case_db`, `alert_observable`, `enrichment_threat`, `assignment_ref`, `attachments`, `audit`)
- [ ] **Migrations applied once** (see below)
- [ ] Start the six domain services, then the gateway
- [ ] Test via gateway (e.g. `curl` to create a case)

## Default gRPC ports (single-host)

| Service | Default gRPC port | Default HTTP port |
|---------|-------------------|-------------------|
| Case Service | **50051** | 8080 |
| Alert & Observable Service | **50052** | 8080 |
| Enrichment & Threat Lookup Service | **50053** | 8080 |
| Assignment & Reference Service | **50054** | 8080 |
| Attachment Service | **50055** | 8080 |
| Audit Service | **50056** | 8080 |
| API Gateway | — | **8080** |

Each domain service listens on its own gRPC port; the gateway uses `CASE_SERVICE_ADDR=localhost:50051`, `OBSERVABLE_SERVICE_ADDR=localhost:50052`, etc. (defaults in gateway config).

## Run migrations once (before first run)

Apply each service’s migrations to its database. From repo root, with `psql` on your PATH and `PGPASSWORD=qwerty`:

```bash
export PGP="PGPASSWORD=qwerty psql -U cursor -h localhost"

# Case Service → case_db
$PGP -d case_db -f case-service/migrations/000001_cases.up.sql
$PGP -d case_db -f case-service/migrations/000002_case_worknotes.up.sql
$PGP -d case_db -f case-service/migrations/000003_case_number_sequence.up.sql
$PGP -d case_db -f case-service/migrations/000004_case_incident_fields.up.sql
$PGP -d case_db -f case-service/migrations/000005_incident_detail_fields.up.sql

# Alert & Observable → alert_observable
$PGP -d alert_observable -f alert-observable-service/migrations/000001_init_schema.up.sql

# Enrichment & Threat → enrichment_threat
$PGP -d enrichment_threat -f enrichment-threat-service/migrations/001_initial_schema.up.sql
$PGP -d enrichment_threat -f enrichment-threat-service/migrations/002_dedupe_indexes.up.sql

# Assignment & Reference → assignment_ref
$PGP -d assignment_ref -f assignment-reference-service/migrations/001_initial_schema.sql
$PGP -d assignment_ref -f assignment-reference-service/migrations/002_seed_poc.sql
$PGP -d assignment_ref -f assignment-reference-service/migrations/003_assignment_groups_soc_csirt.sql

# Attachment → attachments
$PGP -d attachments -f attachment-service/migrations/000001_create_attachments.up.sql

# Audit → audit
$PGP -d audit -f audit-service/migrations/001_audit_events.up.sql
```

If a relation already exists, you may see errors for that file; the rest are still applied. Re-run is safe for idempotent migrations (e.g. `CREATE TABLE IF NOT EXISTS`).

## Order and commands

1. **PostgreSQL** must be running. Create the six databases once (see **Scripts** below). Each service is configured to use its own database with user `cursor` by default; override with env vars if needed.

   **Create databases** (from repo root):
   ```bash
   cd case-service && go run ../scripts/create_dbs.go
   ```
   Or with psql: `psql -U cursor -h localhost -d postgres -f scripts/create_dbs.sql`

   **Databases:** `case_db`, `alert_observable`, `enrichment_threat`, `assignment_ref`, `attachments`, `audit`. Default connection: `cursor`/`qwerty`@localhost:5432 (see each service’s `config/config.go`).

2. **Start domain services** (order not critical; gateway connects on first request):
   ```bash
   # Terminal 1 – Case Service (50051). Set AUDIT_SERVICE_ADDR so case changes are recorded for Audit History.
   cd case-service && AUDIT_SERVICE_ADDR=localhost:50056 go run ./cmd

   # Terminal 2 – Alert & Observable (50052)
   cd alert-observable-service && go run ./cmd

   # Terminal 3 – Enrichment & Threat (50053)
   cd enrichment-threat-service && go run ./cmd

   # Terminal 4 – Assignment & Reference (50054 gRPC; use HTTP_ADDR=:8081 to avoid conflict with gateway on 8080)
   cd assignment-reference-service && HTTP_ADDR=:8081 go run ./cmd

   # Terminal 5 – Attachment (50055)
   cd attachment-service && go run ./cmd

   # Terminal 6 – Audit (50056)
   cd audit-service && go run ./cmd
   ```

3. **Start the gateway** (calls Case Service over gRPC when Case endpoints are used; others use stubs until real clients are wired):
   ```bash
   cd api-gateway-bff && go run ./cmd
   ```

4. **Call the API** (e.g. create a case via gateway):
   ```bash
   curl -X POST http://localhost:8080/api/v1/cases \
     -H "Content-Type: application/json" \
     -H "X-User-Id: user-1" \
     -d '{"title":"Test case","priority":"P2"}'
   ```

## Restarting services after code changes

**Whenever you change code in any service, restart that service and any dependents so the new code is loaded.**

- **Domain service changed** (e.g. case-service, alert-observable-service, assignment-reference-service): stop that process, then start it again with the same command (see **Order and commands** above). If the gateway calls it, **restart the gateway** as well so it uses a fresh connection.
- **Gateway changed** (api-gateway-bff): restart the gateway. No need to restart domain services unless their code also changed.
- **Frontend changed**: restart or refresh the frontend dev server (Vite); the browser will hot-reload for many edits.

**Dependency:** The API Gateway depends on all domain services (case, observable, reference, enrichment, attachment, audit). So after changing any of those, restart the modified service(s) and then the gateway.

## Overriding ports

If you run a domain service on a different port, set the matching gateway env var, e.g.:

- `CASE_SERVICE_ADDR=localhost:15051`
- `OBSERVABLE_SERVICE_ADDR=localhost:15052`
- etc.

For the domain service, set `GRPC_PORT` (or `GRPC_ADDR` where applicable) to the same value.

## Audit history empty

For the case **Audit History** tab to show changes (create, update, worknotes, assign, close), the case service must send events to the audit service:

1. **Audit service** must be running (Terminal 6, port 50056) and its migrations applied to the `audit` database.
2. **Case service** must be started with `AUDIT_SERVICE_ADDR=localhost:50056` so it uses the gRPC audit publisher instead of the no-op. Example: `cd case-service && AUDIT_SERVICE_ADDR=localhost:50056 go run ./cmd`.

After that, new case creates, updates, worknotes, assignments, and closes will appear in Audit History. Existing cases created before this was set will not have historical events.

## Assignment group dropdown empty

If the Assignment Group dropdown shows only "— None —" and no error:

1. **Run migrations** so `assignment_ref` has seed data: `go run ./scripts/run_migrations.go`
2. **Restart assignment-reference-service** so it uses the same DB (or so it self-seeds on startup). The service seeds POC data (Triage, SOC L1, SOC L2, CSIRT) when `assignment_groups` is empty.
3. Ensure **assignment-reference-service** uses the same `DATABASE_URL` as where you ran migrations (default `postgres://cursor:qwerty@localhost:5432/assignment_ref?sslmode=disable`).

Verify DB has groups: `go run ./scripts/verify_assignment_groups.go`

## E2E tests (frontend)

With the gateway and frontend running:

```bash
cd frontend && npm run test:e2e
```

Tests require gateway on port 8080 and (for assignment groups) assignment-reference-service with seeded DB. Start services first, then run tests.

## Integration status

- **Case Service:** Gateway uses **real gRPC client** when `CASE_SERVICE_ADDR` is set (default localhost:50051). Outgoing metadata (x-user-id, x-request-id, x-correlation-id) is set from the HTTP request context.
- **Reference / Assignment:** Gateway uses **real gRPC client** for ListUsers/ListGroups (assignment-reference-service on 50054). Assignment group dropdown and user dropdown require this service and a seeded `assignment_ref` DB.
- **Other services:** Gateway uses real gRPC clients for observable, enrichment, attachment, audit.
- **Audit events:** Producers (e.g. Case Service) use an audit publisher (NoopPublisher by default). Connect a real event bus and implement publishers so the Audit Service consumer receives events.
