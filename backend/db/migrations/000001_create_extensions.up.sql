-- =============================================================================
-- Migration: 000001_create_extensions
-- Description: Enable required PostgreSQL extensions.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;
