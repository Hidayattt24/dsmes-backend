-- ============================================================================
-- Migration: 000021_add_sociodemographic_fields.up.sql
-- Module:    Patient Sociodemographic (manual text: district & health facility)
-- Created:   2026-08-28
-- ============================================================================

-- ── Patient sociodemographic columns ──────────────────────────────────────────
ALTER TABLE patients ADD COLUMN IF NOT EXISTS city            VARCHAR(100) NOT NULL DEFAULT 'Banda Aceh';
ALTER TABLE patients ADD COLUMN IF NOT EXISTS district        VARCHAR(100);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS health_facility VARCHAR(150);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS living_arrangement VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS education_level    VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS diabetes_duration  VARCHAR(50);
