-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- ======================
-- Accounts
-- ======================
CREATE TABLE accounts(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- ======================
-- Profiles (many–1 with accounts)
-- ======================
CREATE TABLE profiles(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES accounts(id)
ON DELETE CASCADE,
  first_name VARCHAR(255) NOT NULL,
  last_name VARCHAR(255) NOT NULL,
  display_name VARCHAR(255),
  avatar_url VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- ======================
-- Account security (1–1 with accounts)
-- ======================
CREATE TABLE account_securities(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL UNIQUE REFERENCES accounts(id)
ON DELETE CASCADE,
  is_email_verified BOOLEAN NOT NULL DEFAULT FALSE,
  email_verified_at TIMESTAMPTZ,
  two_factor_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  two_factor_enabled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- ======================
-- OAuth identities (many–1 with accounts)
-- ======================
CREATE TABLE oauths(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES accounts(id)
ON DELETE CASCADE,
  provider VARCHAR(255) NOT NULL,
  subject VARCHAR(255) NOT NULL,
  token VARCHAR(255) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(provider,
  subject)
);
-- ======================
-- Sessions (many-1 with accounts)
-- ======================
CREATE TABLE sessions(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  account_id UUID NOT NULL REFERENCES accounts(id)
ON DELETE CASCADE,
  refresh_token VARCHAR(255) NOT NULL UNIQUE,
  user_agent VARCHAR(255) NOT NULL,
  ip_addr VARCHAR(255) NOT NULL,
  device_id VARCHAR(255) NOT NULL,
  last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- ======================
-- Indexes for FK performance
-- ======================
CREATE INDEX idx_profiles_account_id
ON profiles(account_id);
CREATE INDEX idx_account_securities_account_id
ON account_securities(account_id);
CREATE INDEX idx_oauths_account_id
ON oauths(account_id);
CREATE INDEX idx_sessions_account_id
ON sessions(account_id);
