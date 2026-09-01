ALTER TABLE blood_sugar_logs
    DROP COLUMN IF EXISTS severity,
    DROP COLUMN IF EXISTS reference_min,
    DROP COLUMN IF EXISTS reference_max,
    DROP COLUMN IF EXISTS reference_range,
    DROP COLUMN IF EXISTS recommendation,
    DROP COLUMN IF EXISTS color;
