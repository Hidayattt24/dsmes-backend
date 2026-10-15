-- Migration: 000029_create_patient_logins_and_app_usages.up.sql
-- Modules: Auth/Patient (Login Tracking) & App Usage (Duration Tracking)

CREATE TABLE IF NOT EXISTS patient_logins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    login_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    device_info VARCHAR(255),
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_patient_logins_patient_id ON patient_logins(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_logins_login_at ON patient_logins(login_at DESC);

CREATE TABLE IF NOT EXISTS patient_app_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    client_id UUID,
    session_id VARCHAR(100) NOT NULL,
    usage_date DATE NOT NULL,
    duration_seconds INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_app_usage_client_id 
    ON patient_app_usages(patient_id, client_id) 
    WHERE client_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_patient_app_usage_session_id 
    ON patient_app_usages(patient_id, session_id);

CREATE INDEX IF NOT EXISTS idx_patient_app_usages_patient_date 
    ON patient_app_usages(patient_id, usage_date DESC);
