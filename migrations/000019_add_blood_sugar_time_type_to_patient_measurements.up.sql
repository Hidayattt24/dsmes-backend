-- Add blood_sugar_time_type column to patient_measurements.
-- The PatientMeasurement domain model has carried this field for a while but the
-- original 000007 migration never created the column. Without it, inserting a
-- blood sugar log (which also writes a measurement snapshot) fails with a 500.
ALTER TABLE patient_measurements ADD COLUMN IF NOT EXISTS blood_sugar_time_type VARCHAR(50);
