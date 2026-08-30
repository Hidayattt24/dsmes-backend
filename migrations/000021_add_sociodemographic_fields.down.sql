-- ============================================================================
-- Migration: 000021_add_sociodemographic_fields.down.sql
-- ============================================================================

ALTER TABLE patients DROP COLUMN IF EXISTS diabetes_duration;
ALTER TABLE patients DROP COLUMN IF EXISTS education_level;
ALTER TABLE patients DROP COLUMN IF EXISTS living_arrangement;
ALTER TABLE patients DROP COLUMN IF EXISTS health_facility;
ALTER TABLE patients DROP COLUMN IF EXISTS district;
ALTER TABLE patients DROP COLUMN IF EXISTS city;
