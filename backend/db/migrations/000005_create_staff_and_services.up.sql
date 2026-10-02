-- =============================================================================
-- Migration: 000005_create_staff_and_services
-- =============================================================================

-- ---------------------------------------------------------------------------
-- staff
-- Service providers (stylists, doctors, consultants, etc.).
-- Vertical-agnostic: no domain-specific columns at core level.
-- ---------------------------------------------------------------------------
CREATE TABLE staff (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID        REFERENCES users(id) ON DELETE SET NULL, -- optional account link
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           CITEXT,
    phone           VARCHAR(30),
    avatar_url      TEXT,
    title           VARCHAR(100), -- e.g., "Senior Stylist", "Dr.", "Senior Associate"
    bio             TEXT,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    accepts_online  BOOLEAN     NOT NULL DEFAULT true, -- visible for online booking
    display_order   INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staff_org_id ON staff (organization_id);
CREATE INDEX idx_staff_user_id ON staff (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_staff_active ON staff (organization_id, is_active);

CREATE TRIGGER trigger_staff_updated_at
    BEFORE UPDATE ON staff
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- staff_locations
-- Many-to-many: a staff member can work at multiple locations.
-- ---------------------------------------------------------------------------
CREATE TABLE staff_locations (
    staff_id    UUID NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (staff_id, location_id)
);

CREATE INDEX idx_staff_locations_location ON staff_locations (location_id);

-- ---------------------------------------------------------------------------
-- staff_schedules
-- Weekly recurring schedule per staff member.
-- ---------------------------------------------------------------------------
CREATE TABLE staff_schedules (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    staff_id        UUID        NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    day_of_week     VARCHAR(10) NOT NULL,
    is_working      BOOLEAN     NOT NULL DEFAULT true,
    start_time      VARCHAR(5)  NOT NULL DEFAULT '09:00',
    end_time        VARCHAR(5)  NOT NULL DEFAULT '17:00',
    break_start     VARCHAR(5),
    break_end       VARCHAR(5),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT staff_schedules_unique UNIQUE (staff_id, day_of_week),
    CONSTRAINT staff_schedules_day_check CHECK (
        day_of_week IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')
    ),
    CONSTRAINT staff_schedules_time_format CHECK (
        start_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'
        AND end_time ~ '^([01][0-9]|2[0-3]):[0-5][0-9]$'
    )
);

CREATE INDEX idx_staff_schedules_staff ON staff_schedules (staff_id);

CREATE TRIGGER trigger_staff_schedules_updated_at
    BEFORE UPDATE ON staff_schedules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- staff_time_offs
-- Date range exceptions to the recurring schedule.
-- ---------------------------------------------------------------------------
CREATE TABLE staff_time_offs (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    staff_id        UUID        NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    start_date      DATE        NOT NULL,
    end_date        DATE        NOT NULL,
    reason          TEXT,
    is_all_day      BOOLEAN     NOT NULL DEFAULT true,
    start_time      VARCHAR(5),  -- for partial-day time off
    end_time        VARCHAR(5),
    approved_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT staff_time_off_date_order CHECK (end_date >= start_date),
    CONSTRAINT staff_time_off_partial_day CHECK (
        is_all_day = true
        OR (start_time IS NOT NULL AND end_time IS NOT NULL)
    )
);

CREATE INDEX idx_staff_time_off_staff ON staff_time_offs (staff_id);
CREATE INDEX idx_staff_time_off_dates ON staff_time_offs (staff_id, start_date, end_date);

CREATE TRIGGER trigger_staff_time_offs_updated_at
    BEFORE UPDATE ON staff_time_offs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- service_categories
-- ---------------------------------------------------------------------------
CREATE TABLE service_categories (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    color           VARCHAR(7),  -- hex color code #RRGGBB
    display_order   INT         NOT NULL DEFAULT 0,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT service_categories_unique_name UNIQUE (organization_id, name)
);

CREATE INDEX idx_service_categories_org ON service_categories (organization_id);

CREATE TRIGGER trigger_service_categories_updated_at
    BEFORE UPDATE ON service_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- services
-- Bookable offerings. Prices in smallest currency unit (cents).
-- ---------------------------------------------------------------------------
CREATE TABLE services (
    id                UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id   UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    category_id       UUID        REFERENCES service_categories(id) ON DELETE SET NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT,
    status            VARCHAR(20)  NOT NULL DEFAULT 'active',
    duration_minutes  INT         NOT NULL,
    buffer_before     INT         NOT NULL DEFAULT 0,  -- prep time in minutes
    buffer_after      INT         NOT NULL DEFAULT 0,  -- cleanup time in minutes
    price_cents       BIGINT      NOT NULL DEFAULT 0,  -- 0 = free/price on request
    currency          VARCHAR(3)  NOT NULL DEFAULT 'USD',
    max_capacity      INT         NOT NULL DEFAULT 1,
    requires_resource BOOLEAN     NOT NULL DEFAULT false,
    color             VARCHAR(7),
    image_url         TEXT,
    display_order     INT         NOT NULL DEFAULT 0,
    is_public         BOOLEAN     NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT services_status_check CHECK (status IN ('active', 'inactive', 'archived')),
    CONSTRAINT services_duration_check CHECK (duration_minutes > 0),
    CONSTRAINT services_buffer_check CHECK (buffer_before >= 0 AND buffer_after >= 0),
    CONSTRAINT services_price_check CHECK (price_cents >= 0),
    CONSTRAINT services_capacity_check CHECK (max_capacity >= 1)
);

CREATE INDEX idx_services_org_id ON services (organization_id);
CREATE INDEX idx_services_category ON services (category_id);
CREATE INDEX idx_services_status ON services (organization_id, status);
CREATE INDEX idx_services_public ON services (organization_id, is_public, status);

CREATE TRIGGER trigger_services_updated_at
    BEFORE UPDATE ON services
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- staff_services
-- Many-to-many: which staff provides which service.
-- Allows per-staff price/duration overrides.
-- ---------------------------------------------------------------------------
CREATE TABLE staff_services (
    staff_id              UUID    NOT NULL REFERENCES staff(id) ON DELETE CASCADE,
    service_id            UUID    NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    organization_id       UUID    NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    custom_price_cents    BIGINT, -- NULL = use service default
    custom_duration_minutes INT,  -- NULL = use service default
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (staff_id, service_id)
);

CREATE INDEX idx_staff_services_service ON staff_services (service_id);
CREATE INDEX idx_staff_services_org ON staff_services (organization_id);

-- ---------------------------------------------------------------------------
-- service_resources
-- Links services to required resources (optional).
-- ---------------------------------------------------------------------------
CREATE TABLE service_resources (
    service_id  UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (service_id, resource_id)
);
