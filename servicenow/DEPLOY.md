# Deploy: One-command setup with Docker Compose

Database (PostgreSQL), all backend services, API gateway, and frontend are managed for you. One command brings everything up.

## Quick start (recommended)

From the repo root:

```bash
docker compose up -d
```

This will:

1. Start **PostgreSQL 16** and create databases: `case_db`, `alert_observable`, `enrichment_threat`, `assignment_ref`, `attachments`, `audit`.
2. Start the **app** container (all 6 domain services + API gateway + Caddy).
3. On first run, the app waits for Postgres to be healthy, then runs all **migrations** automatically.
4. Expose the app on **http://localhost:8089** (frontend and `/api`, `/health`).

No env vars or external database needed. Default DB user: `cursor` / `qwerty` (change in `docker-compose.yml` if you like).

To stop:

```bash
docker compose down
```

To reset and start fresh (deletes DB data):

```bash
docker compose down -v
docker compose up -d
```

## Nginx (frontend at sn.sharedspace360.com)

1. Copy the Nginx config:
   ```bash
   sudo cp nginx-sn.sharedspace360.com.conf /etc/nginx/sites-available/sn.sharedspace360.com
   sudo ln -s /etc/nginx/sites-available/sn.sharedspace360.com /etc/nginx/sites-enabled/
   ```
2. On the server, run the stack with port 8089 published (default in `docker-compose.yml`). If you use another host port, change `proxy_pass http://127.0.0.1:8089` in the config.
3. Test and reload Nginx:
   ```bash
   sudo nginx -t && sudo systemctl reload nginx
   ```

The same server name serves both the frontend (/) and the API (/api, /health).

## Presentation seed (50 cases, 200 observables)

To populate the platform with demo data for presentations:

```bash
curl -X POST http://localhost:8089/api/v1/seed/presentation
```

This creates **50 cases** (with titles, priorities, severities, and assignment groups SOC L1/L2/CSIRT) and **200 observables** (4 per case: IP, domain, email, URL, MD5, SHA256). Response example: `{"message":"Presentation seed completed.","cases_created":50,"observables_created":200}`.

## Checking service status inside the container

If you see errors like `connection refused` to `127.0.0.1:50054` (assignment-reference), a backend may not be listening yet or may have exited. To inspect from the host:

**1. Get the app container name**
```bash
docker compose ps
# or: docker ps --format '{{.Names}}' | head -1
```

**2. Open a shell in the app container**
```bash
docker compose exec app sh
# or: docker exec -it <container_name> sh
```

**3. Inside the container – see what’s listening**
```bash
# List processes (all services run in the same container)
ps aux

# See which ports are in use (Alpine may not have ss; use netstat if available)
netstat -tlnp 2>/dev/null || true
# Or check by hitting health endpoints:
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health   # gateway
curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health   # assignment-reference (gRPC 50054)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8082/health   # alert-observable (50052)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/health   # case-service (50051)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8084/health   # audit-service (50056)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8085/health   # attachment (50055)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8086/health   # enrichment (50053)
```

**4. View logs (from the host)**
```bash
docker compose logs app
docker compose logs app --tail 200
docker compose logs app 2>&1 | grep -i "error\|refused\|50054"
```

**5. Audit-service logs only** (all services run in one container; filter by content):
```bash
# Server-side: IngestEvent / ListAuditEventsByCase handler and service layer
docker compose logs app 2>&1 | grep -E 'IngestEvent|ListAuditEventsByCase'

# All audit-related (including client "audit publish" / "audit ingest" from other services)
docker compose logs app 2>&1 | grep -iE 'IngestEvent|ListAuditEventsByCase|audit publish|audit ingest|audit lazy'
```

For **audit table** write/read flow and commands to check the last 20 records, see **AUDIT_TABLE_VERIFICATION.md**.

**Port map (gRPC → HTTP health):** 50051 case, 50052 observable, 50053 enrichment, **50054 assignment-reference**, 50055 attachment, 50056 audit. The startup script now waits for all backend health endpoints (8081, 8082, 8083, 8085, 8086) before starting the API gateway, which should prevent “connection refused” to 50054 after a fresh start.

## Optional: HTTPS

Use Let’s Encrypt (e.g. `certbot --nginx -d sn.sharedspace360.com`), then uncomment and adjust the HTTPS server block in `nginx-sn.sharedspace360.com.conf` and reload nginx.

---

## Running without Docker Compose (external database)

If you already have Postgres and want to run only the app container:

**Single shared database:**

```bash
docker run -d --name servicenow -p 8089:80 \
  -e DATABASE_URL="postgres://user:pass@dbhost:5432/servicenow?sslmode=disable" \
  -e DB_HOST=dbhost -e DB_USER=user -e DB_PASSWORD=pass -e DB_NAME=servicenow -e DB_SSLMODE=disable \
  servicenow-app:latest
```

**Separate database per service:** set `CASE_DATABASE_URL`, `ENRICHMENT_DATABASE_URL`, `ASSIGNMENT_DATABASE_URL`, `ATTACHMENT_DATABASE_URL`, `AUDIT_DATABASE_URL`, and for alert-observable use `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`. Do not set `PGHOST` so the app skips the migration step (run migrations yourself against your DBs).

Build the image first: `docker build -t servicenow-app:latest .`
