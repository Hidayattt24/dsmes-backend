-- Rollback: drop the status column from patient_measurements.
-- Note: blood_sugar_logs.status categories are NOT reversible to the old
-- labels — keep them as the new diagnostic categories (they are a superset
-- and semantically equivalent for clinical purposes).
ALTER TABLE patient_measurements DROP COLUMN IF EXISTS status;
