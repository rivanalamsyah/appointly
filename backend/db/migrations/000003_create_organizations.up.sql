-- =============================================================================
-- Migration: 000003_create_organizations
-- Description: Organization (tenant) tables — the root of the multi-tenant model.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- organizations
-- Every business data record must reference this table via organization_id.
-- ---------------------------------------------------------------------------
CREATE TABLE organizations (
    id            UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    slug          CITEXT      NOT NULL, -- URL-safe unique identifier
    business_type VARCHAR(50)  NOT NULL DEFAULT 'other',
    email         CITEXT      NOT NULL,
    phone         VARCHAR(30),
    website       TEXT,
    logo_url      TEXT,
    description   TEXT,
    timezone      VARCHAR(100) NOT NULL DEFAULT 'UTC',
    currency      VARCHAR(3)   NOT NULL DEFAULT 'USD', -- ISO 4217
    country       VARCHAR(2)   NOT NULL DEFAULT 'US',  -- ISO 3166-1 alpha-2
    status        VARCHAR(20)  NOT NULL DEFAULT 'active',
    owner_id      UUID        NOT NULL REFERENCES users(id),

    -- Settings (stored as JSONB — these are configuration, not queryable business data)
    settings      JSONB       NOT NULL DEFAULT '{}',

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT organizations_slug_unique UNIQUE (slug),
    CONSTRAINT organizations_slug_format CHECK (slug ~ '^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$'),
    CONSTRAINT organizations_status_check CHECK (
        status IN ('active', 'suspended', 'inactive')
    ),
    CONSTRAINT organizations_business_type_check CHECK (
        business_type IN ('salon', 'clinic', 'consultant', 'law_firm', 'tutor', 'gym', 'other')
    ),
    CONSTRAINT organizations_currency_check CHECK (LENGTH(currency) = 3),
    CONSTRAINT organizations_country_check CHECK (LENGTH(country) = 2)
);

CREATE INDEX idx_organizations_slug ON organizations (slug);
CREATE INDEX idx_organizations_owner_id ON organizations (owner_id);
CREATE INDEX idx_organizations_status ON organizations (status);
CREATE INDEX idx_organizations_created_at ON organizations (created_at DESC);

CREATE TRIGGER trigger_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- organization_members
-- Links users to organizations with a specific role.
-- A user can be a member of multiple organizations.
-- ---------------------------------------------------------------------------
CREATE TABLE organization_members (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role            VARCHAR(20)  NOT NULL DEFAULT 'STAFF',
    invited_by      UUID        REFERENCES users(id),
    invited_at      TIMESTAMPTZ,
    joined_at       TIMESTAMPTZ,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT organization_members_unique UNIQUE (organization_id, user_id),
    CONSTRAINT organization_members_role_check CHECK (
        role IN ('OWNER', 'ADMIN', 'MANAGER', 'RECEPTIONIST', 'STAFF')
    )
);

CREATE INDEX idx_org_members_org_id ON organization_members (organization_id);
CREATE INDEX idx_org_members_user_id ON organization_members (user_id);
CREATE INDEX idx_org_members_active ON organization_members (organization_id, is_active);

CREATE TRIGGER trigger_org_members_updated_at
    BEFORE UPDATE ON organization_members
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- member_invitations
-- Pending invitations to join an organization.
-- ---------------------------------------------------------------------------
CREATE TABLE member_invitations (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email           CITEXT      NOT NULL,
    role            VARCHAR(20)  NOT NULL,
    invited_by      UUID        NOT NULL REFERENCES users(id),
    token_hash      TEXT        NOT NULL UNIQUE,
    expires_at      TIMESTAMPTZ NOT NULL,
    accepted_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT member_invitations_role_check CHECK (
        role IN ('OWNER', 'ADMIN', 'MANAGER', 'RECEPTIONIST', 'STAFF')
    )
);

CREATE INDEX idx_member_invitations_org_id ON member_invitations (organization_id);
CREATE INDEX idx_member_invitations_email ON member_invitations (email);
CREATE UNIQUE INDEX idx_member_invitations_pending
    ON member_invitations (organization_id, email)
    WHERE accepted_at IS NULL;
