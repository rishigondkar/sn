-- Attachment Service: attachments table (Appendix A)
CREATE TABLE IF NOT EXISTS attachments (
  id UUID PRIMARY KEY,
  case_id UUID NOT NULL,
  file_name VARCHAR(500) NOT NULL,
  file_size_bytes BIGINT NOT NULL,
  content_type VARCHAR(255),
  storage_provider VARCHAR(50) NOT NULL DEFAULT 's3',
  storage_key VARCHAR(1000) NOT NULL,
  storage_bucket VARCHAR(255),
  uploaded_by_user_id UUID NOT NULL,
  uploaded_at TIMESTAMPTZ NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT false,
  deleted_at TIMESTAMPTZ,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_attachments_case_id ON attachments(case_id);
CREATE INDEX IF NOT EXISTS idx_attachments_is_deleted ON attachments(is_deleted);
