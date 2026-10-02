-- =============================================================================
-- Migration: 000008_create_audit_subscriptions_webhooks
-- =============================================================================

-- ---------------------------------------------------------------------------
-- audit_logs
-- Immutable append-only audit trail. No UPDATE/DELETE triggers.
-- ---------------------------------------------------------------------------
CREATE TABLE audit_logs (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        REFERENCES organizations(id) ON DELETE SET NULL,
    actor_id        UUID        REFERENCES users(id) ON DELETE SET NULL,
    actor_email     CITEXT,
    actor_type      VARCHAR(20)  NOT NULL DEFAULT 'user', -- user | system | api_key
    action          VARCHAR(100) NOT NULL,
    resource_type   VARCHAR(50),
    resource_id     UUID,
    changes         JSONB,    -- before/after for updates
    metadata        JSONB,
    ip_address      INET,
    user_agent      TEXT,
    request_id      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()

    -- Intentionally NO updated_at — audit logs are immutable
);

CREATE INDEX idx_audit_logs_org ON audit_logs (organization_id, created_at DESC);
CREATE INDEX idx_audit_logs_actor ON audit_logs (actor_id, created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs (action, created_at DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs (resource_type, resource_id)
    WHERE resource_id IS NOT NULL;
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);

-- Prevent updates and deletes on audit_logs (enforce immutability at DB level)
CREATE OR REPLACE RULE audit_logs_no_update AS
    ON UPDATE TO audit_logs DO INSTEAD NOTHING;

CREATE OR REPLACE RULE audit_logs_no_delete AS
    ON DELETE TO audit_logs DO INSTEAD NOTHING;

-- ---------------------------------------------------------------------------
-- plans (subscription tiers)
-- ---------------------------------------------------------------------------
CREATE TABLE plans (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name                VARCHAR(100) NOT NULL,
    slug                VARCHAR(50)  NOT NULL UNIQUE,
    description         TEXT,
    is_active           BOOLEAN     NOT NULL DEFAULT true,

    -- Feature limits (NULL = unlimited)
    max_staff           INT,
    max_locations       INT,
    max_monthly_appointments INT,
    max_customers       INT,
    max_services        INT,

    -- Billing
    monthly_price_cents BIGINT      NOT NULL DEFAULT 0,
    annual_price_cents  BIGINT      NOT NULL DEFAULT 0,
    currency            VARCHAR(3)  NOT NULL DEFAULT 'USD',

    -- Features included
    features            JSONB       NOT NULL DEFAULT '{}',

    display_order       INT         NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_plans_updated_at
    BEFORE UPDATE ON plans
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Seed default plans
INSERT INTO plans (id, name, slug, description, max_staff, max_locations, max_monthly_appointments, monthly_price_cents, annual_price_cents, currency, features, display_order)
VALUES
    (gen_random_uuid(), 'Starter',    'starter',    'For solo practitioners',          2,   1,   100,  0,       0,       'USD', '{"online_booking":true,"basic_reports":true}', 1),
    (gen_random_uuid(), 'Growth',     'growth',     'For growing businesses',          10,  3,   500,  4900,    49000,   'USD', '{"online_booking":true,"basic_reports":true,"advanced_reports":true,"custom_branding":true}', 2),
    (gen_random_uuid(), 'Professional','professional','For established businesses',    50,  10,  2000, 9900,    99000,   'USD', '{"online_booking":true,"basic_reports":true,"advanced_reports":true,"custom_branding":true,"api_access":true,"webhooks":true}', 3),
    (gen_random_uuid(), 'Enterprise', 'enterprise', 'Unlimited scale',                 NULL,NULL, NULL, 29900,   299000,  'USD', '{"online_booking":true,"basic_reports":true,"advanced_reports":true,"custom_branding":true,"api_access":true,"webhooks":true,"sla":true,"dedicated_support":true}', 4);

-- ---------------------------------------------------------------------------
-- subscriptions
-- Links organizations to plans.
-- ---------------------------------------------------------------------------
CREATE TABLE subscriptions (
    id                  UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id     UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE UNIQUE,
    plan_id             UUID        NOT NULL REFERENCES plans(id),
    status              VARCHAR(20)  NOT NULL DEFAULT 'active',
    billing_period      VARCHAR(10)  NOT NULL DEFAULT 'monthly', -- monthly | annual
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end   TIMESTAMPTZ NOT NULL,
    cancelled_at        TIMESTAMPTZ,
    trial_ends_at       TIMESTAMPTZ,
    external_id         TEXT,  -- Stripe subscription ID etc.
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT subscriptions_status_check CHECK (
        status IN ('trialing', 'active', 'past_due', 'cancelled', 'unpaid')
    ),
    CONSTRAINT subscriptions_billing_check CHECK (
        billing_period IN ('monthly', 'annual')
    )
);

CREATE INDEX idx_subscriptions_plan ON subscriptions (plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions (status);

CREATE TRIGGER trigger_subscriptions_updated_at
    BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- webhooks
-- Organization-level outbound webhooks for event notifications.
-- ---------------------------------------------------------------------------
CREATE TABLE webhooks (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    organization_id UUID        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    url             TEXT        NOT NULL,
    secret          TEXT        NOT NULL, -- HMAC signing secret (hashed)
    events          TEXT[]      NOT NULL DEFAULT '{}', -- event types to subscribe to
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    description     TEXT,
    last_triggered_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhooks_org ON webhooks (organization_id);
CREATE INDEX idx_webhooks_active ON webhooks (organization_id, is_active);

CREATE TRIGGER trigger_webhooks_updated_at
    BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- webhook_deliveries
-- Log of webhook delivery attempts.
-- ---------------------------------------------------------------------------
CREATE TABLE webhook_deliveries (
    id              UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    webhook_id      UUID        NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSONB       NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    http_status     INT,
    response_body   TEXT,
    attempt_count   INT         NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMPTZ,
    next_attempt_at TIMESTAMPTZ,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT webhook_deliveries_status_check CHECK (
        status IN ('pending', 'delivered', 'failed', 'cancelled')
    )
);

CREATE INDEX idx_webhook_deliveries_webhook ON webhook_deliveries (webhook_id);
CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries (status, next_attempt_at)
    WHERE status = 'pending';
