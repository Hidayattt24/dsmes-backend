ALTER TABLE patients
    ADD COLUMN IF NOT EXISTS treatment_facility VARCHAR(150);
