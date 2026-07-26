-- Create patient_measurements table
CREATE TABLE IF NOT EXISTS patient_measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id UUID NOT NULL,
    weight_kg NUMERIC(5,2),
    height_cm NUMERIC(5,2),
    bmi NUMERIC(4,1),
    blood_pressure_systolic INT,
    blood_pressure_diastolic INT,
    blood_sugar INT,
    waist_circumference_cm NUMERIC(5,2),
    daily_calorie_target INT,
    notes TEXT,
    recorded_by_id UUID,
    recorded_by_name VARCHAR(150) NOT NULL DEFAULT 'Admin',
    recorded_by_role VARCHAR(50) NOT NULL DEFAULT 'admin',
    measured_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create indexes for patient_measurements
CREATE INDEX IF NOT EXISTS idx_patient_measurements_patient_id ON patient_measurements(patient_id);
CREATE INDEX IF NOT EXISTS idx_patient_measurements_measured_at ON patient_measurements(measured_at);

-- Alter table patients to add calorie recommendations
ALTER TABLE patients 
    ADD COLUMN IF NOT EXISTS maintenance_calories INT,
    ADD COLUMN IF NOT EXISTS mild_weight_loss_calories INT,
    ADD COLUMN IF NOT EXISTS weight_loss_calories INT,
    ADD COLUMN IF NOT EXISTS extreme_weight_loss_calories INT,
    ADD COLUMN IF NOT EXISTS maintenance_percentage INT,
    ADD COLUMN IF NOT EXISTS mild_percentage INT,
    ADD COLUMN IF NOT EXISTS weight_loss_percentage INT,
    ADD COLUMN IF NOT EXISTS extreme_percentage INT;
