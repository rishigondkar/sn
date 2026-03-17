-- POC seed: at least one user and one assignment group for Case/Alert services to reference.
-- Idempotent: use ON CONFLICT DO NOTHING so re-running is safe.

INSERT INTO users (id, username, email, display_name, is_active, created_at, updated_at)
VALUES (
  'a0000000-0000-4000-8000-000000000001',
  'soc-admin',
  'soc-admin@example.com',
  'SOC Admin',
  true,
  NOW(),
  NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO assignment_groups (id, group_name, description, is_active, created_at, updated_at)
VALUES
  ('b0000000-0000-4000-8000-000000000001', 'Triage', 'Default triage assignment group', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000002', 'SOC L1', 'Security Operations Center Level 1', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000003', 'SOC L2', 'Security Operations Center Level 2', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000004', 'CSIRT', 'Computer Security Incident Response Team', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO group_members (id, group_id, user_id, member_role, created_at)
VALUES
  ('c0000000-0000-4000-8000-000000000001', 'b0000000-0000-4000-8000-000000000001', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000002', 'b0000000-0000-4000-8000-000000000002', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000003', 'b0000000-0000-4000-8000-000000000003', 'a0000000-0000-4000-8000-000000000001', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000004', 'b0000000-0000-4000-8000-000000000004', 'a0000000-0000-4000-8000-000000000001', 'member', NOW())
ON CONFLICT (group_id, user_id) DO NOTHING;
