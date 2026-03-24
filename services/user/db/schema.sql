CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    -- core
    email varchar(255) UNIQUE NOT NULL,

    -- security
    password_hash varchar(255), -- NOTE: nullable for only oauthed user
    is_email_verified bool NOT NULL DEFAULT false,
    two_factor_enabled bool NOT NULL DEFAULT false,

    -- timestamptz
    two_factor_enabled_at timestamptz,
    email_verified_at timestamptz,

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE, -- NOTE: 1:1 relation with user

    first_name varchar(255) NOT NULL,
    last_name varchar(255) NOT NULL,
    avatar_url text,

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE oauths (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    provider varchar(255) NOT NULL,
    provider_id varchar(255) NOT NULL,

    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- NOTE: Many:1 relation with user

    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- NOTE: Many:1 relation with user

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

    -- indexes
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE UNIQUE INDEX idx_sessions_refresh_token ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_active ON sessions(user_id) WHERE is_revoked = false; -- NOTE: “active sessions”
CREATE INDEX idx_sessions_expires ON sessions(expires_at); -- NOTE: “expired sessions” and "cleanup jobs"
CREATE INDEX idx_sessions_last_login ON sessions(last_login_at DESC); -- NOTE: “recent devices”
