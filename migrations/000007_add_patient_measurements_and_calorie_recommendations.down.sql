DROP TABLE IF EXISTS patient_measurements CASCADE;

ALTER TABLE patients 
    DROP COLUMN IF EXISTS maintenance_calories,
    DROP COLUMN IF EXISTS mild_weight_loss_calories,
    DROP COLUMN IF EXISTS weight_loss_calories,
    DROP COLUMN IF EXISTS extreme_weight_loss_calories,
    DROP COLUMN IF EXISTS maintenance_percentage,
    DROP COLUMN IF EXISTS mild_percentage,
    DROP COLUMN IF EXISTS weight_loss_percentage,
    DROP COLUMN IF EXISTS extreme_percentage;
