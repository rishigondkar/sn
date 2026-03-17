ALTER TABLE observables DROP CONSTRAINT IF EXISTS observables_finding_check;
ALTER TABLE observables DROP COLUMN IF EXISTS finding;
ALTER TABLE observables DROP COLUMN IF EXISTS incident_count;
