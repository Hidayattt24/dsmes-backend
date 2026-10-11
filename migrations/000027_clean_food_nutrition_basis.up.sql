-- ============================================================================
-- Migration: 000027_clean_food_nutrition_basis.up.sql
-- Module:    Food Nutrition Basis & Serving Size Normalization
-- Created:   2026-10-10
-- ============================================================================

-- 1. Replace hardcoded BDD strings in serving_size with clean '100 g'
UPDATE foods
SET serving_size = '100 g',
    nutrition_basis = 'PER_100G'
WHERE serving_size = 'Per 100 g BDD (Berat Dapat Dimakan)'
   OR serving_size ILIKE '%BDD%';

-- 2. Mark foods with 0 or empty serving sizes as UNVERIFIED
UPDATE foods
SET nutrition_basis = 'UNVERIFIED'
WHERE serving_size IN ('0', '0 g', '0 ml', '0.00 g', '0.0 g', '0 porsi', '')
   OR serving_size IS NULL;

-- 3. Set default for nutrition_basis column to UNVERIFIED instead of PER_100G
ALTER TABLE foods ALTER COLUMN nutrition_basis SET DEFAULT 'UNVERIFIED';
