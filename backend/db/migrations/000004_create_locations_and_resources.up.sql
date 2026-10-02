-- =============================================================================
-- Migration: 000004_create_locations_and_resources
-- =============================================================================

-- ---------------------------------------------------------------------------
-- locations
-- Physical branches of an organization.
-- ---------------------------------------------------------------------------
CREATE TABLE locations (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100),
    description     TEXT,
    phone           VARCHAR(30),
    email           CITEXT,
    address_line1   VARCHAR(255),
    address_line2   VARCHAR(255),
    city            VARCHAR(100),
    state           VARCHAR(100),
    postal_code     VARCHAR(20),
    country         VARCHAR(2)   NOT NULL DEFAULT 'US',
    latitude        DECIMAL(10, 8),
    longitude       DECIMAL(11, 8),
    timezone        VARCHAR(100), -- overrides org timezone if set
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    is_default      BOOLEAN     NOT NULL DEFAULT false,
    display_order   INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT locations_status_check CHECK (status IN ('active', 'inactive')),
    -- Ensure at most one default location per organization
    CONSTRAINT locations_unique_slug UNIQUE (organization_id, slug)
);

CREATE INDEX idx_locations_org_id ON locations (organization_id);
CREATE INDEX idx_locations_status ON locations (organization_id, status);
CREATE UNIQUE INDEX idx_locations_default
    ON locations (organization_id)
    WHERE is_default = true;

CREATE TRIGGER trigger_locations_updated_at
    BEFORE UPDATE ON locations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- business_hours
-- Operating hours per location, per day of week.
-- ---------------------------------------------------------------------------
CREATE TABLE business_hours (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    location_id     UUID        NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    day_of_week     VARCHAR(10) NOT NULL,
    is_open         BOOLEAN     NOT NULL DEFAULT true,
    open_time       VARCHAR(5)  NOT NULL DEFAULT '09:00', -- HH:MM
    close_time      VARCHAR(5)  NOT NULL DEFAULT '17:00', -- HH:MM
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT business_hours_unique UNIQUE (location_id, day_of_week),
    CONSTRAINT business_hours_day_check CHECK (
        day_of_week IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')
    ),
    CONSTRAINT business_hours_time_check CHECK (
        open_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'
        AND close_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'
    )
);

CREATE INDEX idx_business_hours_location ON business_hours (location_id);

CREATE TRIGGER trigger_business_hours_updated_at
    BEFORE UPDATE ON business_hours
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- resources
-- Bookable resources (rooms, chairs, equipment). Optional — services may or
-- may not require a resource.
-- ---------------------------------------------------------------------------
CREATE TABLE resources (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    location_id     UUID        NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    description     TEXT,
    capacity        INT         NOT NULL DEFAULT 1, -- concurrent bookings (usually 1)
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT resources_capacity_check CHECK (capacity >= 1)
);

CREATE INDEX idx_resources_org_id ON resources (organization_id);
CREATE INDEX idx_resources_location_id ON resources (location_id);

CREATE TRIGGER trigger_resources_updated_at
    BEFORE UPDATE ON resources
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
