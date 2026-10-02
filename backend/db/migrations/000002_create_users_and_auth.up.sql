-- =============================================================================
-- Migration: 000002_create_users_and_auth
-- Description: Core identity tables — users, refresh tokens, email verification,
--              password resets.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- users
-- The identity/account table. A user can be a member of multiple organizations.
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    email               CITEXT      NOT NULL,
    password_hash       TEXT        NOT NULL,
    first_name          VARCHAR(100) NOT NULL,
    last_name           VARCHAR(100) NOT NULL,
    phone               VARCHAR(30),
    avatar_url          TEXT,
    is_email_verified   BOOLEAN     NOT NULL DEFAULT false,
    is_super_admin      BOOLEAN     NOT NULL DEFAULT false,
    last_login_at       TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_created_at ON users (created_at DESC);

-- ---------------------------------------------------------------------------
-- refresh_tokens
-- Opaque refresh tokens stored as SHA-256 hashes. One-time use (rotated).
-- ---------------------------------------------------------------------------
CREATE TABLE refresh_tokens (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   TEXT        NOT NULL, -- SHA-256 of the actual token
    device_info  TEXT,
    ip_address   INET,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT refresh_tokens_token_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at)
    WHERE revoked_at IS NULL;

-- ---------------------------------------------------------------------------
-- email_verifications
-- ---------------------------------------------------------------------------
CREATE TABLE email_verifications (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_verifications_user_id ON email_verifications (user_id);

-- ---------------------------------------------------------------------------
-- password_resets
-- ---------------------------------------------------------------------------
CREATE TABLE password_resets (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_resets_user_id ON password_resets (user_id);

-- ---------------------------------------------------------------------------
-- updated_at trigger function (shared across tables)
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
