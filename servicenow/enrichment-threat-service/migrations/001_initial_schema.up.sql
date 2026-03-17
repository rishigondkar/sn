-- Enrichment & Threat Lookup Service – initial schema (Appendix A)

CREATE TABLE enrichment_results (
  id UUID PRIMARY KEY,
  case_id UUID,
  observable_id UUID,
  enrichment_type VARCHAR(100) NOT NULL,
  source_name VARCHAR(100) NOT NULL,
  source_record_id VARCHAR(255),
  status VARCHAR(30) NOT NULL,
  summary VARCHAR(1000),
  result_data JSONB NOT NULL,
  score NUMERIC(10,2),
  confidence NUMERIC(5,2),
  requested_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ,
  last_updated_by VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  CONSTRAINT enrichment_results_case_or_observable CHECK (case_id IS NOT NULL OR observable_id IS NOT NULL)
);

CREATE TABLE threat_lookup_results (
  id UUID PRIMARY KEY,
  case_id UUID,
  observable_id UUID NOT NULL,
  lookup_type VARCHAR(100) NOT NULL,
  source_name VARCHAR(100) NOT NULL,
  source_record_id VARCHAR(255),
  verdict VARCHAR(50),
  risk_score NUMERIC(10,2),
  confidence_score NUMERIC(5,2),
  tags JSONB,
  matched_indicators JSONB,
  summary VARCHAR(1000),
  result_data JSONB NOT NULL,
  queried_at TIMESTAMPTZ,
  received_at TIMESTAMPTZ NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_enrichment_results_case_id ON enrichment_results(case_id);
CREATE INDEX idx_enrichment_results_observable_id ON enrichment_results(observable_id);
CREATE INDEX idx_enrichment_results_source_name ON enrichment_results(source_name);
CREATE INDEX idx_enrichment_results_received_at ON enrichment_results(received_at);

CREATE INDEX idx_threat_lookup_results_case_id ON threat_lookup_results(case_id);
CREATE INDEX idx_threat_lookup_results_observable_id ON threat_lookup_results(observable_id);
CREATE INDEX idx_threat_lookup_results_source_name ON threat_lookup_results(source_name);
CREATE INDEX idx_threat_lookup_results_received_at ON threat_lookup_results(received_at);
