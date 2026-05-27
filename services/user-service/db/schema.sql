CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- Helper Functions
-- =====================================================

-- Automatically update updated_at timestamp on row changes
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Users
-- =====================================================
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email varchar(255) UNIQUE NOT NULL,
    status varchar(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'inactive', 'deleted')),
    password_hash varchar(255),
    is_email_verified bool NOT NULL DEFAULT false,
    two_factor_enabled bool NOT NULL DEFAULT false,
    two_factor_enabled_at timestamptz,
    email_verified_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    deleted_at timestamptz
);

COMMENT on COLUMN users.status IS 'The status of the user (active, inactive, deleted)';

CREATE TRIGGER set_updated_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Use Case: Find users by status
CREATE INDEX idx_users_status ON users(status);
-- Use Case: Find users by email
CREATE INDEX idx_users_email ON users(email);
-- Use Case: Find users by is_email_verified
CREATE INDEX idx_users_is_email_verified ON users(is_email_verified);
-- Use Case: Find users by two_factor_enabled
CREATE INDEX idx_users_two_factor_enabled ON users(two_factor_enabled);

-- ============================================
-- Profiles
-- ============================================
CREATE TABLE profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    avatar_url text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER set_updated_at_profiles
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Use Case: Find profiles by user_id
CREATE INDEX idx_profiles_user_id ON profiles(user_id);

-- ============================================
-- OAuths
-- ============================================
CREATE TABLE oauths (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider varchar(255) NOT NULL,
    provider_id varchar(255) NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER set_updated_at_oauths
    BEFORE UPDATE ON oauths
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Use Case: Find oauths by user_id
CREATE INDEX idx_oauths_user_id ON oauths(user_id);
-- Use Case: Find oauths by provider and provider_id
CREATE INDEX idx_oauths_provider_provider_id ON oauths(provider, provider_id);

-- ============================================
-- Sessions
-- ============================================
CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash varchar(255) NOT NULL,
    user_agent varchar(255) NOT NULL,
    ip_addr varchar(255) NOT NULL,
    device_id varchar(255) NOT NULL,
    last_login_at timestamptz DEFAULT now(),
    expires_at timestamptz NOT NULL,
    is_revoked bool NOT NULL DEFAULT false,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TRIGGER set_updated_at_sessions
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Use Case: Find sessions by user_id
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
-- Use Case: Find sessions by refresh_token_hash
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
-- Use Case: Find sessions by is_revoked
CREATE INDEX idx_sessions_is_revoked ON sessions(is_revoked);
-- Use Case: Find sessions by expires_at
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
-- Use Case: Find sessions by last_login_at
CREATE INDEX idx_sessions_last_login ON sessions(last_login_at DESC);
