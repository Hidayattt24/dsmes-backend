-- ============================================================================
-- Migration: 000023_create_health_facilities.down.sql
-- ============================================================================

ALTER TABLE staff_accounts DROP COLUMN IF EXISTS health_facility_id;
DROP TABLE IF EXISTS health_facilities;
