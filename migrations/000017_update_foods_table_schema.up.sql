-- ============================================================================
-- Migration: 000017_update_foods_table_schema.up.sql
-- Module:    Food Master Data Refactor
-- Created:   2026-08-04
-- Synchronize foods table with full dataset schema and NUMERIC(10,2) types
-- ============================================================================

-- Fix old migration 000002 calories NOT NULL constraint
ALTER TABLE foods ALTER COLUMN calories DROP NOT NULL;
ALTER TABLE foods ALTER COLUMN calories SET DEFAULT 0.00;

ALTER TABLE foods ADD COLUMN IF NOT EXISTS manufacturer VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE foods ADD COLUMN IF NOT EXISTS serving_size VARCHAR(255) NOT NULL DEFAULT '1 porsi';

-- Nutrition Values (NUMERIC(10,2))
ALTER TABLE foods ADD COLUMN IF NOT EXISTS energy_kcal NUMERIC(10,2) NOT NULL DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS carbohydrate_g NUMERIC(10,2) NOT NULL DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS sugar_g NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS sodium_mg NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS fiber_g NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS saturated_fat_g NUMERIC(10,2) DEFAULT 0.00;

-- Ensure existing numeric columns use NUMERIC(10,2)
ALTER TABLE foods ALTER COLUMN calories TYPE NUMERIC(10,2);
ALTER TABLE foods ALTER COLUMN protein_g TYPE NUMERIC(10,2);
ALTER TABLE foods ALTER COLUMN fat_g TYPE NUMERIC(10,2);

-- %DV Columns (NUMERIC(10,2))
ALTER TABLE foods ADD COLUMN IF NOT EXISTS energy_percentage_dv NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS protein_percentage_dv NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS carbohydrate_percentage_dv NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS fat_percentage_dv NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS sodium_percentage_dv NUMERIC(10,2) DEFAULT 0.00;

-- Additional Label Columns (NUMERIC(10,2))
ALTER TABLE foods ADD COLUMN IF NOT EXISTS total_fat NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS saturated_fat NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS sodium NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS protein NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS total_carbohydrate NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS dietary_fiber NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS energy NUMERIC(10,2) DEFAULT 0.00;

-- Metadata Columns
ALTER TABLE foods ADD COLUMN IF NOT EXISTS source VARCHAR(100) DEFAULT 'manual';
ALTER TABLE foods ADD COLUMN IF NOT EXISTS barcode VARCHAR(100);
ALTER TABLE foods ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE foods ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'active';

-- Indexes
CREATE INDEX IF NOT EXISTS idx_foods_manufacturer ON foods(manufacturer);
CREATE INDEX IF NOT EXISTS idx_foods_barcode ON foods(barcode);
CREATE INDEX IF NOT EXISTS idx_foods_status ON foods(status);

-- Composite Unique Constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_food_unique_entry ON foods(name, manufacturer, serving_size) WHERE deleted_at IS NULL;

-- ── Meal Logs Nutrition Snapshot Columns ──────────────────────────────────────
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS food_name VARCHAR(255);
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS serving_size VARCHAR(100);
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS calories NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS carbs_g NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS protein_g NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS fat_g NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS sugar_g NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS sodium_mg NUMERIC(8,2) DEFAULT 0.00;
ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS fiber_g NUMERIC(8,2) DEFAULT 0.00;

