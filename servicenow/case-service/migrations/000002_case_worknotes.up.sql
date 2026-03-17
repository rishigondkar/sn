CREATE TABLE IF NOT EXISTS case_worknotes (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL REFERENCES cases(id),
  note_text TEXT NOT NULL,
  note_type VARCHAR(30) NOT NULL DEFAULT 'worknote',
  created_by_user_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS idx_case_worknotes_case_id ON case_worknotes(case_id);
