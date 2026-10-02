-- =============================================================================
-- Migration: 000002_create_users_and_auth (DOWN)
-- =============================================================================

DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
