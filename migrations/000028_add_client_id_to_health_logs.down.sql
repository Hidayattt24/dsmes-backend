-- 000028_add_client_id_to_health_logs.down.sql

DROP INDEX IF EXISTS idx_quiz_attempt_patient_client;
ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS client_id;

DROP INDEX IF EXISTS idx_meal_patient_client;
ALTER TABLE meal_logs DROP COLUMN IF EXISTS client_id;

DROP INDEX IF EXISTS idx_activity_patient_client;
ALTER TABLE patient_activity_logs DROP COLUMN IF EXISTS client_id;

DROP INDEX IF EXISTS idx_blood_sugar_patient_client;
ALTER TABLE blood_sugar_logs DROP COLUMN IF EXISTS client_id;
