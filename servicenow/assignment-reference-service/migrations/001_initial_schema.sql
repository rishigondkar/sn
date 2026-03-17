-- Assignment Reference Service: initial schema (Appendix A)
-- Run against PostgreSQL.

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY,
  username VARCHAR(100) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  display_name VARCHAR(255) NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active);
CREATE INDEX IF NOT EXISTS idx_users_display_name ON users (display_name);
CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS assignment_groups (
  id UUID PRIMARY KEY,
  group_name VARCHAR(255) NOT NULL UNIQUE,
  description TEXT,
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_assignment_groups_is_active ON assignment_groups (is_active);
CREATE INDEX IF NOT EXISTS idx_assignment_groups_group_name ON assignment_groups (group_name);

CREATE TABLE IF NOT EXISTS group_members (
  id UUID PRIMARY KEY,
  group_id UUID NOT NULL REFERENCES assignment_groups(id),
  user_id UUID NOT NULL REFERENCES users(id),
  member_role VARCHAR(50),
  created_at TIMESTAMPTZ NOT NULL,
  UNIQUE (group_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_group_members_group_id ON group_members (group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members (user_id);
