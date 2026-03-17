#!/bin/sh
# Run only assignment_ref migrations so users and assignment groups (SOC L1, SOC L2, CSIRT, Triage) always exist.
# Called first from start.sh so groups are created even if other migrations fail later.
set -e
PGHOST="${PGHOST:-localhost}"
PGUSER="${PGUSER:-cursor}"
PGPORT="${PGPORT:-5432}"
export PGPASSWORD="${PGPASSWORD:-qwerty}"
M="/app/migrations"

run_file() {
  echo "  Applying $(basename "$2") to assignment_ref..."
  psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d assignment_ref -f "$2"
}

echo "Ensuring assignment_ref DB and assignment groups (users, SOC L1, SOC L2, CSIRT)..."
for f in "$M/assignment-reference-service"/001_initial_schema.sql \
         "$M/assignment-reference-service"/002_seed_poc.sql \
         "$M/assignment-reference-service"/003_assignment_groups_soc_csirt.sql \
         "$M/assignment-reference-service"/004_seed_assignment_users_and_groups.sql; do
  [ -f "$f" ] && run_file assignment_ref "$f"
done
echo "Assignment groups ready."
