-- ============================================================================
-- Migration: 000022_fix_sociodemographic_columns.down.sql
-- ============================================================================

ALTER TABLE patients DROP COLUMN IF EXISTS health_facility;
ALTER TABLE patients DROP COLUMN IF EXISTS district;
