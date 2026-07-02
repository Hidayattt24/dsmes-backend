-- ============================================================================
-- Migration: 000001_create_auth_and_staff_tables.down.sql
-- Rolls back the auth + staff tables migration.
-- ============================================================================

DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS staff_accounts;

DROP TYPE IF EXISTS account_status_enum;
DROP TYPE IF EXISTS owner_type_enum;
DROP TYPE IF EXISTS staff_role_enum;
