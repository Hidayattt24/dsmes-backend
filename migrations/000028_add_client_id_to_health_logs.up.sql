-- 000028_add_client_id_to_health_logs.up.sql
-- Adds client_id and partial unique index per patient for permanent idempotency

ALTER TABLE blood_sugar_logs ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_blood_sugar_patient_client 
    ON blood_sugar_logs(patient_id, client_id) 
    WHERE client_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE patient_activity_logs ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_activity_patient_client 
    ON patient_activity_logs(patient_id, client_id) 
    WHERE client_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE meal_logs ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_meal_patient_client 
    ON meal_logs(patient_id, client_id) 
    WHERE client_id IS NOT NULL AND deleted_at IS NULL;

ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_quiz_attempt_patient_client 
    ON quiz_attempts(patient_id, client_id) 
    WHERE client_id IS NOT NULL AND deleted_at IS NULL;
