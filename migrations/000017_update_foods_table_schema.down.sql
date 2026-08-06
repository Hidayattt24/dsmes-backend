-- ============================================================================
-- Migration: 000017_update_foods_table_schema.down.sql
-- Module:    Food Master Data Refactor Rollback
-- ============================================================================

ALTER TABLE meal_logs DROP COLUMN IF EXISTS fiber_g;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS sodium_mg;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS sugar_g;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS fat_g;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS protein_g;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS carbs_g;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS calories;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS serving_size;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS food_name;

DROP INDEX IF EXISTS idx_food_unique_entry;
DROP INDEX IF EXISTS idx_foods_status;
DROP INDEX IF EXISTS idx_foods_barcode;
DROP INDEX IF EXISTS idx_foods_manufacturer;

ALTER TABLE foods DROP COLUMN IF EXISTS status;
ALTER TABLE foods DROP COLUMN IF EXISTS image_url;
ALTER TABLE foods DROP COLUMN IF EXISTS barcode;
ALTER TABLE foods DROP COLUMN IF EXISTS source;
ALTER TABLE foods DROP COLUMN IF EXISTS energy;
ALTER TABLE foods DROP COLUMN IF EXISTS dietary_fiber;
ALTER TABLE foods DROP COLUMN IF EXISTS total_carbohydrate;
ALTER TABLE foods DROP COLUMN IF EXISTS protein;
ALTER TABLE foods DROP COLUMN IF EXISTS sodium;
ALTER TABLE foods DROP COLUMN IF EXISTS saturated_fat;
ALTER TABLE foods DROP COLUMN IF EXISTS total_fat;
ALTER TABLE foods DROP COLUMN IF EXISTS sodium_percentage_dv;
ALTER TABLE foods DROP COLUMN IF EXISTS fat_percentage_dv;
ALTER TABLE foods DROP COLUMN IF EXISTS carbohydrate_percentage_dv;
ALTER TABLE foods DROP COLUMN IF EXISTS protein_percentage_dv;
ALTER TABLE foods DROP COLUMN IF EXISTS energy_percentage_dv;
ALTER TABLE foods DROP COLUMN IF EXISTS saturated_fat_g;
ALTER TABLE foods DROP COLUMN IF EXISTS fiber_g;
ALTER TABLE foods DROP COLUMN IF EXISTS sodium_mg;
ALTER TABLE foods DROP COLUMN IF EXISTS sugar_g;
ALTER TABLE foods DROP COLUMN IF EXISTS carbohydrate_g;
ALTER TABLE foods DROP COLUMN IF EXISTS energy_kcal;
ALTER TABLE foods DROP COLUMN IF EXISTS serving_size;
ALTER TABLE foods DROP COLUMN IF EXISTS manufacturer;
