-- 000020: Reclassify blood sugar status columns to the unified diagnostic
-- categories (hypoglycemia / normal / target / prediabetes / elevated /
-- hyperglycemia). The previous categories (tinggi, sangat_tinggi, rendah,
-- waspada, berbahaya, severe_hypoglycemia, severe_hyperglycemia) are
-- superseded by the single ClassifyBloodGlucose() function in domain.
-- ============================================================================

-- 1. Add status column to patient_measurements so dashboard queries can read
--    the pre-computed category without a fragile CASE WHEN.
ALTER TABLE patient_measurements ADD COLUMN IF NOT EXISTS status VARCHAR(50);

-- 2. Backfill patient_measurements.status using the same logic as the Go
--    classifier (ClassifyBloodGlucose). All thresholds here MUST match what
--    domain/blood_sugar.go implements.

UPDATE patient_measurements
SET status = (
    CASE
        -- Universal: < 70 → hypoglycemia
        WHEN blood_sugar < 70 THEN 'hypoglycemia'

        -- Puasa (fasting / before_meal)
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('fasting','puasa','before_meal','sebelum_makan')
             AND blood_sugar >= 126 THEN 'hyperglycemia'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('fasting','puasa','before_meal','sebelum_makan')
             AND blood_sugar >= 100 THEN 'prediabetes'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('fasting','puasa','before_meal','sebelum_makan')
             AND blood_sugar < 100  THEN 'normal'

        -- 2 Jam Setelah Makan (after_meal / GD2PP)
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND blood_sugar >= 200 THEN 'hyperglycemia'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND blood_sugar >= 140 THEN 'prediabetes'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND blood_sugar < 140  THEN 'normal'

        -- Sebelum Tidur (before_bed)
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('before_bed','sebelum_tidur')
             AND blood_sugar >= 200 THEN 'hyperglycemia'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('before_bed','sebelum_tidur')
             AND blood_sugar >= 140 THEN 'elevated'
        WHEN COALESCE(blood_sugar_time_type, 'random') IN ('before_bed','sebelum_tidur')
             AND blood_sugar < 140  THEN 'target'

        -- Sewaktu / default (GDS / random)
        WHEN blood_sugar >= 200 THEN 'hyperglycemia'
        ELSE 'normal'
    END
)
WHERE blood_sugar IS NOT NULL AND blood_sugar > 0 AND deleted_at IS NULL;

-- 3. Backfill blood_sugar_logs.status to new categories (recomputed from
--    glucose_value + measurement_time_type so values are deterministic).
--    This is the SAME logic as the Go classifier — no divergence.

UPDATE blood_sugar_logs
SET status = (
    CASE
        WHEN glucose_value < 70 THEN 'hypoglycemia'

        WHEN measurement_time_type IN ('fasting','puasa','before_meal','sebelum_makan')
             AND glucose_value >= 126 THEN 'hyperglycemia'
        WHEN measurement_time_type IN ('fasting','puasa','before_meal','sebelum_makan')
             AND glucose_value >= 100 THEN 'prediabetes'
        WHEN measurement_time_type IN ('fasting','puasa','before_meal','sebelum_makan')
             AND glucose_value < 100  THEN 'normal'

        WHEN measurement_time_type IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND glucose_value >= 200 THEN 'hyperglycemia'
        WHEN measurement_time_type IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND glucose_value >= 140 THEN 'prediabetes'
        WHEN measurement_time_type IN ('after_meal','sesudah_makan','2_jam_sesudah_makan')
             AND glucose_value < 140  THEN 'normal'

        WHEN measurement_time_type IN ('before_bed','sebelum_tidur')
             AND glucose_value >= 200 THEN 'hyperglycemia'
        WHEN measurement_time_type IN ('before_bed','sebelum_tidur')
             AND glucose_value >= 140 THEN 'elevated'
        WHEN measurement_time_type IN ('before_bed','sebelum_tidur')
             AND glucose_value < 140  THEN 'target'

        WHEN glucose_value >= 200 THEN 'hyperglycemia'
        ELSE 'normal'
    END
)
WHERE deleted_at IS NULL;
