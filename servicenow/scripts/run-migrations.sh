#!/bin/sh
# Run all service migrations. Uses PGHOST, PGUSER, PGPASSWORD (defaults: cursor/qwerty).
# Idempotent: safe to re-run (ON CONFLICT DO NOTHING / CREATE TABLE IF NOT EXISTS).
# Assignment_ref (users + assignment groups) is run FIRST so groups always exist even if other migrations fail.
set -e
PGHOST="${PGHOST:-localhost}"
PGUSER="${PGUSER:-cursor}"
PGPORT="${PGPORT:-5432}"
export PGPASSWORD="${PGPASSWORD:-qwerty}"

run_file() {
  echo "  Applying $(basename "$2") to $1..."
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$1" -f "$2"
}

M="/app/migrations"

# Run assignment_ref FIRST so users and assignment groups (SOC L1, SOC L2, CSIRT, Triage) always exist
echo "Migrating assignment_ref (users, assignment groups, group_members)..."
for f in "$M/assignment-reference-service"/001_initial_schema.sql "$M/assignment-reference-service"/002_seed_poc.sql \
         "$M/assignment-reference-service"/003_assignment_groups_soc_csirt.sql \
         "$M/assignment-reference-service"/004_seed_assignment_users_and_groups.sql; do
  [ -f "$f" ] && run_file assignment_ref "$f"
done
echo "Assignment groups and users ready."

echo "Migrating case_db..."
for f in "$M/case-service"/000001_cases.up.sql "$M/case-service"/000002_case_worknotes.up.sql \
         "$M/case-service"/000003_case_number_sequence.up.sql "$M/case-service"/000004_case_incident_fields.up.sql \
         "$M/case-service"/000005_incident_detail_fields.up.sql; do
  [ -f "$f" ] && run_file case_db "$f"
done

echo "Migrating alert_observable..."
for f in "$M/alert-observable-service"/000001_init_schema.up.sql "$M/alert-observable-service"/000002_add_observable_incident_count_and_finding.up.sql \
         "$M/alert-observable-service"/000003_add_observable_notes.up.sql; do
  [ -f "$f" ] && run_file alert_observable "$f"
done

echo "Migrating enrichment_threat..."
for f in "$M/enrichment-threat-service"/001_initial_schema.up.sql "$M/enrichment-threat-service"/002_dedupe_indexes.up.sql; do
  [ -f "$f" ] && run_file enrichment_threat "$f"
done

echo "Migrating attachments..."
[ -f "$M/attachment-service/000001_create_attachments.up.sql" ] && run_file attachments "$M/attachment-service/000001_create_attachments.up.sql"

echo "Migrating audit..."
[ -f "$M/audit-service/001_audit_events.up.sql" ] && run_file audit "$M/audit-service/001_audit_events.up.sql"

echo "Migrations completed."
