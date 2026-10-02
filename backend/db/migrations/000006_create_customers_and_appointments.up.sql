-- =============================================================================
-- Migration: 000006_create_customers_and_appointments
-- Description: Core booking tables — customers, appointments, status history.
-- =============================================================================

-- ---------------------------------------------------------------------------
-- customers
-- Persons who book appointments. Scoped to organization.
-- Soft-deleted for audit trail preservation.
-- ---------------------------------------------------------------------------
CREATE TABLE customers (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id         UUID        REFERENCES users(id) ON DELETE SET NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    email           CITEXT,
    phone           VARCHAR(30),
    notes           TEXT,

    -- Aggregated stats (updated asynchronously on appointment events)
    total_appointments      INT     NOT NULL DEFAULT 0,
    completed_appointments  INT     NOT NULL DEFAULT 0,
    no_show_count           INT     NOT NULL DEFAULT 0,
    last_appointment_at     TIMESTAMPTZ,
    total_spent_cents       BIGINT  NOT NULL DEFAULT 0,

    deleted_at  TIMESTAMPTZ, -- soft delete
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_customers_org_id ON customers (organization_id);
CREATE INDEX idx_customers_email ON customers (organization_id, email) WHERE email IS NOT NULL;
CREATE INDEX idx_customers_phone ON customers (organization_id, phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_customers_user_id ON customers (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_customers_active ON customers (organization_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_customers_updated_at
    BEFORE UPDATE ON customers
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- appointments
-- The central booking entity. Everything connects here.
-- ---------------------------------------------------------------------------
CREATE TABLE appointments (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    location_id     UUID        NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    service_id      UUID        NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    staff_id        UUID        NOT NULL REFERENCES staff(id) ON DELETE RESTRICT,
    customer_id     UUID        REFERENCES customers(id) ON DELETE SET NULL,
    resource_id     UUID        REFERENCES resources(id) ON DELETE SET NULL,

    -- Timing (always UTC in DB)
    start_time      TIMESTAMPTZ NOT NULL,
    end_time        TIMESTAMPTZ NOT NULL,
    timezone        VARCHAR(100) NOT NULL DEFAULT 'UTC', -- timezone customer booked in

    -- Status
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',

    -- Pricing snapshot at booking time
    price_cents     BIGINT      NOT NULL DEFAULT 0,
    currency        VARCHAR(3)  NOT NULL DEFAULT 'USD',

    -- Guest booking fields (used when customer_id is NULL)
    guest_name      VARCHAR(200),
    guest_email     CITEXT,
    guest_phone     VARCHAR(30),

    -- Metadata
    notes           TEXT,
    internal_notes  TEXT,
    cancel_reason   TEXT,
    source          VARCHAR(30)  NOT NULL DEFAULT 'online', -- online | manual | api

    -- References
    payment_id      UUID,
    created_by      UUID        NOT NULL REFERENCES users(id),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT appointments_status_check CHECK (
        status IN ('pending', 'confirmed', 'rescheduled', 'completed', 'cancelled', 'no_show')
    ),
    CONSTRAINT appointments_time_order CHECK (end_time > start_time),
    CONSTRAINT appointments_source_check CHECK (
        source IN ('online', 'manual', 'api', 'import')
    ),
    -- Guest booking: must have either customer_id OR guest contact info
    CONSTRAINT appointments_customer_check CHECK (
        customer_id IS NOT NULL
        OR (guest_name IS NOT NULL AND (guest_email IS NOT NULL OR guest_phone IS NOT NULL))
    )
);

CREATE INDEX idx_appointments_org_id ON appointments (organization_id);
CREATE INDEX idx_appointments_staff_time ON appointments (staff_id, start_time, end_time);
CREATE INDEX idx_appointments_customer ON appointments (customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_appointments_location ON appointments (location_id);
CREATE INDEX idx_appointments_service ON appointments (service_id);
CREATE INDEX idx_appointments_status ON appointments (organization_id, status);
CREATE INDEX idx_appointments_date_range ON appointments (organization_id, start_time, end_time);
CREATE INDEX idx_appointments_resource ON appointments (resource_id) WHERE resource_id IS NOT NULL;

-- Critical index for availability engine: detect staff schedule conflicts.
-- Only active appointments (not cancelled/no_show) block slots.
CREATE INDEX idx_appointments_staff_active
    ON appointments (staff_id, start_time, end_time)
    WHERE status NOT IN ('cancelled', 'no_show');

-- Partial unique index to prevent double-booking the same staff slot.
-- Two confirmed/pending appointments cannot overlap for the same staff.
-- The actual enforcement is done in the application via SELECT FOR UPDATE,
-- but this index provides an additional database-level safety net.
CREATE UNIQUE INDEX idx_appointments_no_double_book
    ON appointments (staff_id, start_time)
    WHERE status NOT IN ('cancelled', 'no_show');

CREATE TRIGGER trigger_appointments_updated_at
    BEFORE UPDATE ON appointments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- appointment_status_history
-- Immutable audit trail for appointment status changes.
-- ---------------------------------------------------------------------------
CREATE TABLE appointment_status_history (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    appointment_id  UUID        NOT NULL REFERENCES appointments(id) ON DELETE CASCADE,
    from_status     VARCHAR(20) NOT NULL,
    to_status       VARCHAR(20) NOT NULL,
    reason          TEXT,
    changed_by      UUID        NOT NULL REFERENCES users(id),
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_appt_history_appointment ON appointment_status_history (appointment_id);
CREATE INDEX idx_appt_history_changed_at ON appointment_status_history (changed_at DESC);
