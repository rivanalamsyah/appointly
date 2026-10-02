-- =============================================================================
-- Migration: 000007_create_payments_and_notifications
-- =============================================================================

-- ---------------------------------------------------------------------------
-- payments
-- Separate lifecycle from appointments.
-- ---------------------------------------------------------------------------
CREATE TABLE payments (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    appointment_id  UUID        NOT NULL REFERENCES appointments(id) ON DELETE RESTRICT,
    customer_id     UUID        REFERENCES customers(id) ON DELETE SET NULL,

    status          VARCHAR(30)  NOT NULL DEFAULT 'pending',
    provider        VARCHAR(20)  NOT NULL DEFAULT 'manual',

    amount_cents        BIGINT  NOT NULL,
    currency            VARCHAR(3) NOT NULL DEFAULT 'USD',
    refunded_amount_cents BIGINT NOT NULL DEFAULT 0,

    -- Provider references
    external_id         TEXT,   -- payment ID at the provider
    external_reference  TEXT,   -- order/invoice ID
    payment_url         TEXT,   -- checkout page URL
    receipt_url         TEXT,

    notes       TEXT,
    paid_at     TIMESTAMPTZ,
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT payments_status_check CHECK (
        status IN ('pending', 'paid', 'failed', 'refunded', 'partially_refunded', 'expired', 'cancelled')
    ),
    CONSTRAINT payments_provider_check CHECK (
        provider IN ('stripe', 'midtrans', 'xendit', 'manual')
    ),
    CONSTRAINT payments_amount_check CHECK (amount_cents > 0),
    CONSTRAINT payments_refund_check CHECK (refunded_amount_cents >= 0 AND refunded_amount_cents <= amount_cents)
);

CREATE INDEX idx_payments_org_id ON payments (organization_id);
CREATE INDEX idx_payments_appointment ON payments (appointment_id);
CREATE INDEX idx_payments_customer ON payments (customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_payments_external_id ON payments (external_id) WHERE external_id IS NOT NULL;
CREATE INDEX idx_payments_status ON payments (organization_id, status);
CREATE INDEX idx_payments_created_at ON payments (organization_id, created_at DESC);

CREATE TRIGGER trigger_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Now add the FK from appointments to payments (circular dependency resolved)
ALTER TABLE appointments
    ADD CONSTRAINT fk_appointments_payment
    FOREIGN KEY (payment_id) REFERENCES payments(id) ON DELETE SET NULL;

-- ---------------------------------------------------------------------------
-- refunds
-- ---------------------------------------------------------------------------
CREATE TABLE refunds (
    id          UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    payment_id  UUID        NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount_cents BIGINT     NOT NULL,
    reason      TEXT,
    external_id TEXT,
    status      VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_by  UUID        NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT refunds_amount_check CHECK (amount_cents > 0),
    CONSTRAINT refunds_status_check CHECK (status IN ('pending', 'completed', 'failed'))
);

CREATE INDEX idx_refunds_payment ON refunds (payment_id);

CREATE TRIGGER trigger_refunds_updated_at
    BEFORE UPDATE ON refunds
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- notifications
-- Event queue for all notifications.
-- ---------------------------------------------------------------------------
CREATE TABLE notifications (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    event_type      VARCHAR(50)  NOT NULL,
    recipient_type  VARCHAR(20)  NOT NULL, -- customer | staff | member
    recipient_id    UUID,
    recipient_email CITEXT,
    recipient_phone VARCHAR(30),
    recipient_name  VARCHAR(200),
    reference_type  VARCHAR(50)  NOT NULL, -- appointment | payment | invitation
    reference_id    UUID        NOT NULL,
    channels        TEXT[]      NOT NULL DEFAULT '{}', -- email, sms, whatsapp, in_app, webhook
    template_data   JSONB       NOT NULL DEFAULT '{}',
    scheduled_at    TIMESTAMPTZ, -- NULL = send immediately
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_org ON notifications (organization_id);
CREATE INDEX idx_notifications_scheduled ON notifications (scheduled_at)
    WHERE scheduled_at IS NOT NULL;
CREATE INDEX idx_notifications_reference ON notifications (reference_type, reference_id);

-- ---------------------------------------------------------------------------
-- notification_deliveries
-- Per-channel delivery tracking with retry support.
-- ---------------------------------------------------------------------------
CREATE TABLE notification_deliveries (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    notification_id UUID        NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    channel         VARCHAR(20)  NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    attempt_count   INT         NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    sent_at         TIMESTAMPTZ,
    error_message   TEXT,
    external_id     TEXT, -- message ID from the provider
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT notification_deliveries_channel_check CHECK (
        channel IN ('email', 'sms', 'whatsapp', 'in_app', 'webhook')
    ),
    CONSTRAINT notification_deliveries_status_check CHECK (
        status IN ('pending', 'sent', 'failed', 'skipped')
    )
);

CREATE INDEX idx_notif_deliveries_notification ON notification_deliveries (notification_id);
CREATE INDEX idx_notif_deliveries_pending ON notification_deliveries (status, created_at)
    WHERE status = 'pending';

CREATE TRIGGER trigger_notif_deliveries_updated_at
    BEFORE UPDATE ON notification_deliveries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
