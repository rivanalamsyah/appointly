-- =============================================================================
-- Migration: 000003_create_organizations (DOWN)
-- =============================================================================

DROP TABLE IF EXISTS member_invitations;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
