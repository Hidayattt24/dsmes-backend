-- ============================================================================
-- Migration: 000018_add_nutrition_basis_to_foods.down.sql
-- Module:    Food Master Data Refactor Rollback
-- ============================================================================

DROP INDEX IF EXISTS idx_foods_nutrition_basis;

ALTER TABLE foods DROP COLUMN IF EXISTS nutrition_basis;
