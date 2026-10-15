DROP INDEX IF EXISTS idx_patient_app_usages_deleted_at;

ALTER TABLE patient_app_usages
    DROP COLUMN IF EXISTS deleted_at;
