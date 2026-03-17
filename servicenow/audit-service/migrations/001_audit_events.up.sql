CREATE TABLE audit_events (
  id UUID PRIMARY KEY,
  event_id VARCHAR(100) NOT NULL UNIQUE,
  event_type VARCHAR(100) NOT NULL,
  source_service VARCHAR(100) NOT NULL,
  entity_type VARCHAR(100) NOT NULL,
  entity_id UUID NOT NULL,
  parent_entity_type VARCHAR(100),
  parent_entity_id UUID,
  case_id UUID,
  observable_id UUID,
  action VARCHAR(50) NOT NULL,
  actor_user_id UUID,
  actor_name VARCHAR(255),
  request_id VARCHAR(100),
  correlation_id VARCHAR(100),
  change_summary VARCHAR(1000),
  before_data JSONB,
  after_data JSONB,
  metadata JSONB,
  occurred_at TIMESTAMPTZ NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_audit_events_case_id ON audit_events(case_id);
CREATE INDEX idx_audit_events_observable_id ON audit_events(observable_id);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_actor_user_id ON audit_events(actor_user_id);
CREATE INDEX idx_audit_events_correlation_id ON audit_events(correlation_id);
CREATE INDEX idx_audit_events_occurred_at ON audit_events(occurred_at);
