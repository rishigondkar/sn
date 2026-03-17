#!/bin/sh
set -e

# If PGHOST is set (e.g. by docker-compose), ensure DBs exist and run migrations
if [ -n "${PGHOST}" ]; then
  echo "Waiting for Postgres at $PGHOST..."
  until psql -h "$PGHOST" -p "${PGPORT:-5432}" -U "${PGUSER:-cursor}" -d postgres -c 'select 1' 2>/dev/null; do
    sleep 1
  done
  echo "Creating databases if missing..."
  PGPASSWORD="${PGPASSWORD:-qwerty}" psql -h "$PGHOST" -p "${PGPORT:-5432}" -U "${PGUSER:-cursor}" -d postgres -f /app/scripts/create_dbs.sql || true
  echo "Ensuring assignment_ref exists..."
  PGPASSWORD="${PGPASSWORD:-qwerty}" psql -h "$PGHOST" -p "${PGPORT:-5432}" -U "${PGUSER:-cursor}" -d postgres -c "CREATE DATABASE assignment_ref" 2>/dev/null || true
  echo "Creating assignment groups and users (SOC L1, SOC L2, CSIRT, Triage)..."
  /app/scripts/run-assignment-ref-only.sh || echo "Warning: assignment-ref-only failed; full migrations will retry assignment_ref."
  echo "Running remaining migrations..."
  /app/scripts/run-migrations.sh
fi

# Service addresses (gateway talks to these on localhost)
export CASE_SERVICE_ADDR="${CASE_SERVICE_ADDR:-localhost:50051}"
export OBSERVABLE_SERVICE_ADDR="${OBSERVABLE_SERVICE_ADDR:-localhost:50052}"
export ENRICHMENT_SERVICE_ADDR="${ENRICHMENT_SERVICE_ADDR:-localhost:50053}"
export REFERENCE_SERVICE_ADDR="${REFERENCE_SERVICE_ADDR:-localhost:50054}"
export ATTACHMENT_SERVICE_ADDR="${ATTACHMENT_SERVICE_ADDR:-localhost:50055}"
export AUDIT_SERVICE_ADDR="${AUDIT_SERVICE_ADDR:-localhost:50056}"

# Keep 8080 for API gateway only. Each backend uses a dedicated HTTP port.
export HTTP_ADDR="${HTTP_ADDR:-:8081}"

# Start audit-service first so case-service (and any other publisher) can connect when they start
( export DATABASE_URL="${AUDIT_DATABASE_URL:-$DATABASE_URL}"; export HTTP_PORT=8084; exec /app/bin/audit-service ) &
echo "Waiting for audit-service (gRPC 50056) to be ready..."
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
  if curl -sS -o /dev/null -w "%{http_code}" http://localhost:8084/health 2>/dev/null | grep -q 200; then
    echo "Audit-service ready."
    break
  fi
  if [ "$i" = "15" ]; then
    echo "Warning: audit-service health check did not return 200 after 15 attempts; continuing anyway."
  fi
  sleep 1
done

# Per-service DB URLs and HTTP ports. AUDIT_SERVICE_ADDR is exported above so all backends that publish audit events receive it.
( export DATABASE_URL="${CASE_DATABASE_URL:-$DATABASE_URL}"; export HTTP_PORT=8083; exec /app/bin/case-service ) &
( export HTTP_PORT=8082; exec /app/bin/alert-observable-service ) &
( export DATABASE_URL="${ENRICHMENT_DATABASE_URL:-$DATABASE_URL}"; export HTTP_PORT=8086; exec /app/bin/enrichment-threat-service ) &
( export DATABASE_URL="${ASSIGNMENT_DATABASE_URL:-$DATABASE_URL}"; export HTTP_ADDR="${HTTP_ADDR:-:8081}"; exec /app/bin/assignment-reference-service ) &
( export DATABASE_URL="${ATTACHMENT_DATABASE_URL:-$DATABASE_URL}"; export HTTP_PORT=8085; exec /app/bin/attachment-service ) &

# Wait for all backends to be reachable (HTTP health) so the gateway does not get "connection refused" on 50051-50056
echo "Waiting for backends (case 8083, observable 8082, enrichment 8086, reference 8081, attachment 8085)..."
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
  ok=0
  curl -sS -o /dev/null -w "%{http_code}" http://localhost:8081/health 2>/dev/null | grep -q 200 && ok=$((ok+1)) || true
  curl -sS -o /dev/null -w "%{http_code}" http://localhost:8082/health 2>/dev/null | grep -q 200 && ok=$((ok+1)) || true
  curl -sS -o /dev/null -w "%{http_code}" http://localhost:8083/health 2>/dev/null | grep -q 200 && ok=$((ok+1)) || true
  curl -sS -o /dev/null -w "%{http_code}" http://localhost:8085/health 2>/dev/null | grep -q 200 && ok=$((ok+1)) || true
  curl -sS -o /dev/null -w "%{http_code}" http://localhost:8086/health 2>/dev/null | grep -q 200 && ok=$((ok+1)) || true
  if [ "$ok" = "5" ]; then
    echo "All backends ready."
    break
  fi
  if [ "$i" = "20" ]; then
    echo "Warning: not all backends returned 200 after 20 attempts; continuing anyway."
  fi
  sleep 1
done

/app/bin/api-gateway &
sleep 1

# Wait for gateway to be ready, then seed presentation data (50 cases, 200 observables)
echo "Waiting for API gateway..."
for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
  if curl -sS -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q 200; then
    echo "Gateway ready. Seeding presentation data..."
    out=$(curl -sS -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/seed/presentation)
    code=$(echo "$out" | tail -n1)
    body=$(echo "$out" | sed '$d')
    if [ "$code" = "200" ]; then
      echo "Presentation seed OK: $body"
    else
      echo "Presentation seed returned HTTP $code: $body"
    fi
    break
  fi
  sleep 2
done

exec caddy run --config /app/Caddyfile
