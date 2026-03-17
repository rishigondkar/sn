CREATE TABLE IF NOT EXISTS cases (
  id UUID PRIMARY KEY,
  case_number VARCHAR(50) NOT NULL UNIQUE,
  short_description VARCHAR(500) NOT NULL,
  description TEXT,
  state VARCHAR(50) NOT NULL,
  priority VARCHAR(20) NOT NULL,
  severity VARCHAR(20) NOT NULL,
  opened_by_user_id UUID NOT NULL,
  opened_time TIMESTAMPTZ NOT NULL,
  event_occurred_time TIMESTAMPTZ,
  event_received_time TIMESTAMPTZ,
  affected_user_id UUID,
  assigned_user_id UUID,
  assignment_group_id UUID,
  alert_rule_id UUID,
  active_duration_seconds BIGINT NOT NULL DEFAULT 0,
  accuracy VARCHAR(50),
  determination VARCHAR(100),
  impact VARCHAR(50),
  closure_code VARCHAR(50),
  closure_reason TEXT,
  closed_by_user_id UUID,
  closed_time TIMESTAMPTZ,
  is_active BOOLEAN NOT NULL DEFAULT true,
  version_no INTEGER NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_cases_state ON cases(state);
CREATE INDEX IF NOT EXISTS idx_cases_opened_time ON cases(opened_time DESC);
CREATE INDEX IF NOT EXISTS idx_cases_assigned_user ON cases(assigned_user_id);
CREATE INDEX IF NOT EXISTS idx_cases_assignment_group ON cases(assignment_group_id);
