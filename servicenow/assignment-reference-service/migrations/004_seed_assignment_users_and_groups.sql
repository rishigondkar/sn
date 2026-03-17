-- Seed dummy assignment users and ensure the three assignment groups (SOC L1, SOC L2, CSIRT) exist.
-- Idempotent: ON CONFLICT DO NOTHING so re-running is safe.

-- Ensure the three assignment groups exist (in case 002/003 were skipped or differ)
INSERT INTO assignment_groups (id, group_name, description, is_active, created_at, updated_at)
VALUES
  ('b0000000-0000-4000-8000-000000000002', 'SOC L1', 'Security Operations Center Level 1', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000003', 'SOC L2', 'Security Operations Center Level 2', true, NOW(), NOW()),
  ('b0000000-0000-4000-8000-000000000004', 'CSIRT', 'Computer Security Incident Response Team', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Dummy assignment users
INSERT INTO users (id, username, email, display_name, is_active, created_at, updated_at)
VALUES
  ('a0000000-0000-4000-8000-000000000002', 'soc-l1-alice', 'alice@example.com', 'Alice (SOC L1)', true, NOW(), NOW()),
  ('a0000000-0000-4000-8000-000000000003', 'soc-l1-bob', 'bob@example.com', 'Bob (SOC L1)', true, NOW(), NOW()),
  ('a0000000-0000-4000-8000-000000000004', 'soc-l2-carol', 'carol@example.com', 'Carol (SOC L2)', true, NOW(), NOW()),
  ('a0000000-0000-4000-8000-000000000005', 'soc-l2-dave', 'dave@example.com', 'Dave (SOC L2)', true, NOW(), NOW()),
  ('a0000000-0000-4000-8000-000000000006', 'csirt-eve', 'eve@example.com', 'Eve (CSIRT)', true, NOW(), NOW()),
  ('a0000000-0000-4000-8000-000000000007', 'csirt-frank', 'frank@example.com', 'Frank (CSIRT)', true, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Assign users to the three groups (SOC L1, SOC L2, CSIRT)
-- SOC L1: alice, bob
INSERT INTO group_members (id, group_id, user_id, member_role, created_at)
VALUES
  ('c0000000-0000-4000-8000-000000000010', 'b0000000-0000-4000-8000-000000000002', 'a0000000-0000-4000-8000-000000000002', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000011', 'b0000000-0000-4000-8000-000000000002', 'a0000000-0000-4000-8000-000000000003', 'member', NOW())
ON CONFLICT (group_id, user_id) DO NOTHING;

-- SOC L2: carol, dave
INSERT INTO group_members (id, group_id, user_id, member_role, created_at)
VALUES
  ('c0000000-0000-4000-8000-000000000012', 'b0000000-0000-4000-8000-000000000003', 'a0000000-0000-4000-8000-000000000004', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000013', 'b0000000-0000-4000-8000-000000000003', 'a0000000-0000-4000-8000-000000000005', 'member', NOW())
ON CONFLICT (group_id, user_id) DO NOTHING;

-- CSIRT: eve, frank
INSERT INTO group_members (id, group_id, user_id, member_role, created_at)
VALUES
  ('c0000000-0000-4000-8000-000000000014', 'b0000000-0000-4000-8000-000000000004', 'a0000000-0000-4000-8000-000000000006', 'member', NOW()),
  ('c0000000-0000-4000-8000-000000000015', 'b0000000-0000-4000-8000-000000000004', 'a0000000-0000-4000-8000-000000000007', 'member', NOW())
ON CONFLICT (group_id, user_id) DO NOTHING;
