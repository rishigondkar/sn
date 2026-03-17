-- Alert & Observable Service schema (Appendix A)
CREATE TABLE alert_rules (
  id UUID PRIMARY KEY,
  rule_name VARCHAR(255) NOT NULL,
  rule_type VARCHAR(100),
  source_system VARCHAR(100),
  external_rule_key VARCHAR(255),
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE observables (
  id UUID PRIMARY KEY,
  observable_type VARCHAR(50) NOT NULL,
  observable_value VARCHAR(1000) NOT NULL,
  normalized_value VARCHAR(1000),
  first_seen_time TIMESTAMPTZ,
  last_seen_time TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (observable_type, normalized_value)
);

CREATE TABLE alerts (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  alert_rule_id UUID REFERENCES alert_rules(id),
  source_system VARCHAR(100) NOT NULL,
  source_alert_id VARCHAR(255),
  title VARCHAR(500),
  description TEXT,
  event_occurred_time TIMESTAMPTZ,
  event_received_time TIMESTAMPTZ,
  severity VARCHAR(20),
  raw_payload JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE case_observables (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  observable_id UUID NOT NULL REFERENCES observables(id),
  role_in_case VARCHAR(50),
  tracking_status VARCHAR(50),
  is_primary BOOLEAN NOT NULL DEFAULT false,
  accuracy VARCHAR(50),
  determination VARCHAR(100),
  impact VARCHAR(50),
  added_by_user_id UUID,
  added_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (case_id, observable_id)
);

CREATE TABLE child_observables (
  id UUID PRIMARY KEY,
  parent_observable_id UUID NOT NULL REFERENCES observables(id),
  child_observable_id UUID NOT NULL REFERENCES observables(id),
  relationship_type VARCHAR(100) NOT NULL,
  relationship_direction VARCHAR(20),
  confidence NUMERIC(5,2),
  source_name VARCHAR(100),
  source_record_id VARCHAR(255),
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (parent_observable_id, child_observable_id, relationship_type)
);

CREATE TABLE similar_security_incidents (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  similar_case_id UUID NOT NULL,
  match_reason VARCHAR(100) NOT NULL DEFAULT 'shared_observable',
  shared_observable_count INTEGER NOT NULL DEFAULT 1,
  shared_observable_ids JSONB NOT NULL,
  shared_observable_values JSONB,
  similarity_score NUMERIC(10,2),
  last_computed_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  UNIQUE (case_id, similar_case_id)
);

CREATE INDEX idx_case_observables_case_id ON case_observables(case_id);
CREATE INDEX idx_case_observables_observable_id ON case_observables(observable_id);
CREATE INDEX idx_similar_security_incidents_case_id ON similar_security_incidents(case_id);
CREATE INDEX idx_child_observables_parent_observable_id ON child_observables(parent_observable_id);
