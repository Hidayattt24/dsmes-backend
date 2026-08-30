-- ============================================================================
-- Migration: 000022_fix_sociodemographic_columns.up.sql
-- Module:    Patient Sociodemographic
--
-- Reconciles the patients schema after earlier iterations of migration 000021
-- used FK-based columns (district_id, health_facility_id) and master tables
-- (districts, health_facilities). The final design stores these as plain text
-- columns: `district` and `health_facility`. This migration is fully idempotent
-- and safe to run on both fresh and already-migrated databases.
-- ============================================================================

-- 1. Ensure the final text columns exist (idempotent)
ALTER TABLE patients ADD COLUMN IF NOT EXISTS city               VARCHAR(100) NOT NULL DEFAULT 'Banda Aceh';
ALTER TABLE patients ADD COLUMN IF NOT EXISTS district           VARCHAR(100);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS health_facility    VARCHAR(150);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS living_arrangement VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS education_level    VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS diabetes_duration  VARCHAR(50);

-- 2. Drop obsolete FK-based columns from the earlier 000021 design
ALTER TABLE patients DROP COLUMN IF EXISTS district_id;
ALTER TABLE patients DROP COLUMN IF EXISTS health_facility_id;

-- 3. Drop obsolete master data tables from the earlier 000021 design
DROP TABLE IF EXISTS health_facilities;
DROP TABLE IF EXISTS districts;
