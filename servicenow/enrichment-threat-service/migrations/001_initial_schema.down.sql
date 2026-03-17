DROP INDEX IF EXISTS idx_threat_lookup_results_received_at;
DROP INDEX IF EXISTS idx_threat_lookup_results_source_name;
DROP INDEX IF EXISTS idx_threat_lookup_results_observable_id;
DROP INDEX IF EXISTS idx_threat_lookup_results_case_id;
DROP TABLE IF EXISTS threat_lookup_results;

DROP INDEX IF EXISTS idx_enrichment_results_received_at;
DROP INDEX IF EXISTS idx_enrichment_results_source_name;
DROP INDEX IF EXISTS idx_enrichment_results_observable_id;
DROP INDEX IF EXISTS idx_enrichment_results_case_id;
DROP TABLE IF EXISTS enrichment_results;
