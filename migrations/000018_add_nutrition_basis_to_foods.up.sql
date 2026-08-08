-- ============================================================================
-- Migration: 000018_add_nutrition_basis_to_foods.up.sql
-- Module:    Food Master Data Refactor
-- Created:   2026-08-04
-- Add nutrition_basis column to foods table (PER_100G, PER_SERVING, PER_PACKAGE)
-- ============================================================================

ALTER TABLE foods ADD COLUMN IF NOT EXISTS nutrition_basis VARCHAR(50) NOT NULL DEFAULT 'PER_100G';

CREATE INDEX IF NOT EXISTS idx_foods_nutrition_basis ON foods(nutrition_basis);
