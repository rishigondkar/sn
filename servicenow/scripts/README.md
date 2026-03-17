# Scripts

## create_dbs.sql

Creates the six PostgreSQL databases used by the services (if running manually with psql):

- `case_db` – Case Service  
- `alert_observable` – Alert & Observable Service  
- `enrichment_threat` – Enrichment & Threat Lookup Service  
- `assignment_ref` – Assignment & Reference Service  
- `attachments` – Attachment Service  
- `audit` – Audit Service  

**With psql:**

```bash
psql -U cursor -h localhost -d postgres -f scripts/create_dbs.sql
```

(If a database already exists, that `CREATE DATABASE` will error; the rest still run.)

## create_dbs.go

Same as above, but runnable without psql. Uses `cursor`/`qwerty` by default (override with `DATABASE_URL`).

**Run from case-service** (has pgx dependency):

```bash
cd case-service && go run ../scripts/create_dbs.go
```

Defaults: `postgres://cursor:qwerty@localhost:5432/postgres?sslmode=disable`. Each service’s config uses the corresponding database name; see each service’s `config/config.go`.
