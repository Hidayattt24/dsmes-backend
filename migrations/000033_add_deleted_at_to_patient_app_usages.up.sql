ALTER TABLE patient_app_usages
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_patient_app_usages_deleted_at
    ON patient_app_usages(deleted_at);
