# Alert & Observable Service

Source of truth for alerts, observables, case-observable links, child observable relationships, and similar incident relationships. Part of the SOC case management platform.

## Scope

- **In scope:** Alert rules, alerts, observables, linking observables to cases, tracking status, child observable relations, similar-incident computation, audit event publishing.
- **Out of scope:** Case lifecycle, worknotes, threat lookup/enrichment persistence, attachment persistence, audit event storage (consumed by Audit Service).

## Ports

| Port  | Purpose        |
|-------|----------------|
| 50051 | gRPC server    |
| 8080  | HTTP (health)  |

## Configuration (environment)

| Variable     | Default       | Description        |
|-------------|---------------|--------------------|
| GRPC_PORT   | 50051         | gRPC listen port   |
| HTTP_PORT   | 8080          | Health HTTP port   |
| DB_HOST     | localhost     | PostgreSQL host    |
| DB_PORT     | 5432          | PostgreSQL port    |
| DB_USER     | postgres      | Database user      |
| DB_PASSWORD | (empty)       | Database password  |
| DB_NAME     | alert_observable | Database name   |
| DB_SSLMODE  | disable       | SSL mode           |

## Database

PostgreSQL. Run migrations before starting:

```bash
# Apply migrations in order (e.g. using psql)
psql "$DATABASE_URL" -f migrations/000001_init_schema.up.sql
psql "$DATABASE_URL" -f migrations/000002_add_observable_incident_count_and_finding.up.sql
psql "$DATABASE_URL" -f migrations/000003_add_observable_notes.up.sql
```

Schema: see `migrations/000001_init_schema.up.sql` and PRD Appendix A. **000002** adds `incident_count` and `finding` on `observables` (required for GetObservable and INSERT). **000003** adds `notes` (required for the observable detail page).

## Build and run

```bash
go build -o bin/server ./cmd
./bin/server
```

Health check:

```bash
curl http://localhost:8080/health
```

## Proto

Regenerate Go stubs after editing the proto:

```bash
cd proto && ./generate_proto.sh alert_observable_service go
```

## Tests

```bash
go test ./...
```

Repository tests that hit PostgreSQL run only when `TEST_DB_DSN` is set (e.g. `postgres://user:pass@localhost/alert_observable?sslmode=disable`).

## gRPC API

See `03_alert_observable_service_prd.md` and `proto/alert_observable_service.proto` for:

- **Commands:** CreateAlertRule, CreateAlert, CreateOrGetObservable, LinkObservableToCase, UpdateCaseObservable, CreateChildObservableRelation, RecomputeSimilarIncidentsForCase
- **Queries:** GetAlert, ListCaseAlerts, GetObservable, ListCaseObservables, ListChildObservables, ListSimilarIncidents, FindCasesByObservable

Actor and tracing: set gRPC metadata `x-user-id`, `x-request-id`, `x-correlation-id` (and optional `x-user-name`).

## Audit events

The service publishes audit events after successful writes (event types: `alert.created`, `observable.created`, `observable.linked.to_case`, `observable.updated`, `child_observable.created`, `similar_incident.linked`). By default a no-op publisher is used; replace with a real publisher (e.g. to SNS/SQS) when the event bus is configured.

## Module

`github.com/org/alert-observable-service` — replace `org` with your organization in `go.mod` and proto `go_package` if needed.
