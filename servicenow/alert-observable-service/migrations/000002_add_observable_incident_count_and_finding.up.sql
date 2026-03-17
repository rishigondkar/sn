-- Add incident_count and finding to observables table.
-- Finding: -- None --, Unknown, Malicious, Suspicious, Clean
ALTER TABLE observables
  ADD COLUMN IF NOT EXISTS incident_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS finding VARCHAR(50);

ALTER TABLE observables
  ADD CONSTRAINT observables_finding_check
  CHECK (finding IS NULL OR finding IN ('-- None --', 'Unknown', 'Malicious', 'Suspicious', 'Clean'));
