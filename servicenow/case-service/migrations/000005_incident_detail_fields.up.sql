-- Incident detail page fields (requested by, environment, PDN, notification time, MTTR, counts, VIP, documents).

ALTER TABLE cases ADD COLUMN IF NOT EXISTS requested_by_user_id UUID;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS environment_level VARCHAR(50);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS environment_type VARCHAR(50);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS pdn VARCHAR(255);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS impacted_object VARCHAR(100);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS notification_time TIMESTAMPTZ;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS mttr VARCHAR(100);
ALTER TABLE cases ADD COLUMN IF NOT EXISTS reassignment_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS assigned_to_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS is_affected_user_vip BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS engineering_document TEXT;
ALTER TABLE cases ADD COLUMN IF NOT EXISTS response_document TEXT;
