-- =============================================================================
-- Migration: 000001_create_extensions (DOWN)
-- =============================================================================

DROP EXTENSION IF EXISTS citext;
DROP EXTENSION IF EXISTS pgcrypto;
