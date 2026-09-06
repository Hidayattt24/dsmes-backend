CREATE TABLE IF NOT EXISTS patient_device_tokens (
    token TEXT PRIMARY KEY,
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    platform VARCHAR(20) NOT NULL,
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patient_device_tokens_patient_id
    ON patient_device_tokens(patient_id);
