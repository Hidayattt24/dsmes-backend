-- ============================================================================
-- Migration: 000027_clean_food_nutrition_basis.down.sql
-- Module:    Food Nutrition Basis & Serving Size Normalization
-- ============================================================================

ALTER TABLE foods ALTER COLUMN nutrition_basis SET DEFAULT 'PER_100G';
