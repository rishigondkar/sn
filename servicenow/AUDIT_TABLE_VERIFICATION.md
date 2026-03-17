# Audit table: how it's populated and how to check it

## Write path (populates the table)

1. Case-service, alert-observable-service, attachment-service, enrichment-threat-service, and assignment-reference-service call their **audit publisher** after each change (e.g. create case, add worknote, link observable, update case).
2. The publisher sends the event to **audit-service** via gRPC **IngestEvent** (audit-service listens on **50056**).
3. audit-service handler `IngestEvent` → service `IngestEvent` → repository `InsertEvent` → **INSERT** into database **audit**, table **audit_events** (`ON CONFLICT (event_id) DO NOTHING`).

So if the audit table is not updating, either:
- Events are not reaching the audit-service (publish fails, or `AUDIT_SERVICE_ADDR` not set for a backend).
- The gRPC connection to the audit-service went bad after the first few requests (e.g. after the seed burst). Each publisher now **retries once** on connection error (clears client, reconnects, then sends again), so later operations (worknote, state change, observable link) should still be written after a rebuild.
- audit-service is not writing (check audit-service logs).
- You are querying a different database than the one audit-service uses (should be DB **audit** in the same Postgres instance).

## Read path (fetch for UI)

1. Frontend calls **GET /api/v1/cases/{caseId}/audit-events**.
2. API gateway calls audit-service gRPC **ListAuditEventsByCase** (gateway dials **AUDIT_SERVICE_ADDR**, default localhost:50056).
3. audit-service handler **ListAuditEventsByCase** → service **ListByCase** → repository **ListByCaseID** → **SELECT** from **audit_events** `WHERE case_id = $1 ORDER BY occurred_at DESC`.

So the same **audit** database and **audit_events** table are used for both writes and reads. If the table has rows but the UI shows nothing, the fetch path (gateway → audit-service or case_id format) may be the issue; if the table is empty, the write path is the issue.

---

## Commands: check audit table from Docker

**Last 20 records (from host, using Postgres container):**
```bash
docker compose exec postgres psql -U cursor -d audit -c "SELECT event_id, event_type, source_service, entity_type, action, case_id, occurred_at FROM audit_events ORDER BY occurred_at DESC LIMIT 20;"
```

**From inside the app container** (app can reach Postgres at host `postgres`):
```bash
docker compose exec app sh
```
Then:
```bash
PGPASSWORD=qwerty psql -h postgres -p 5432 -U cursor -d audit -c "SELECT event_id, event_type, source_service, entity_type, action, case_id, occurred_at FROM audit_events ORDER BY occurred_at DESC LIMIT 20;"
```

**Row count and latest timestamps:**
```bash
docker compose exec postgres psql -U cursor -d audit -c "SELECT COUNT(*) AS total, MAX(occurred_at) AS latest_occurred, MAX(ingested_at) AS latest_ingested FROM audit_events;"
```

**If the table is empty or not updating**, check app logs for audit publish/ingest errors:
```bash
docker compose logs app 2>&1 | grep -i "audit publish\|audit ingest\|audit lazy"
```
