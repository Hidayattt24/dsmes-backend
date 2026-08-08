CREATE TABLE patient_activity_logs (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id        UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    activity_name     VARCHAR(255) NOT NULL DEFAULT '',
    duration_minutes  INT          NOT NULL DEFAULT 0,
    intensity         VARCHAR(50)  NOT NULL DEFAULT 'Ringan',
    notes             TEXT         NOT NULL DEFAULT '',
    logged_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_patient_activity_logs_patient ON patient_activity_logs(patient_id);
