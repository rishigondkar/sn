-- Unique indexes for idempotent upsert (natural key dedupe)

-- Enrichment: one row per (case_id, observable_id, enrichment_type, source_name, source_record_id)
-- Use sentinel UUID for COALESCE so NULLs participate in uniqueness
CREATE UNIQUE INDEX idx_enrichment_results_dedupe ON enrichment_results (
  COALESCE(case_id, '00000000-0000-0000-0000-000000000000'::uuid),
  COALESCE(observable_id, '00000000-0000-0000-0000-000000000000'::uuid),
  enrichment_type,
  source_name,
  COALESCE(source_record_id, '')
);

-- Threat lookup: one row per (observable_id, lookup_type, source_name, source_record_id)
CREATE UNIQUE INDEX idx_threat_lookup_results_dedupe ON threat_lookup_results (
  observable_id,
  lookup_type,
  source_name,
  COALESCE(source_record_id, '')
);
