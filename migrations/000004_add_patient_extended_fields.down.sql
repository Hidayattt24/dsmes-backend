ALTER TABLE patients DROP COLUMN IF EXISTS patient_code;
ALTER TABLE patients DROP COLUMN IF EXISTS address;
ALTER TABLE patients DROP COLUMN IF EXISTS diagnosis_date;
ALTER TABLE patients DROP COLUMN IF EXISTS current_medication;
ALTER TABLE patients DROP COLUMN IF EXISTS allergies;
ALTER TABLE patients DROP COLUMN IF EXISTS smoking_status;
ALTER TABLE patients DROP COLUMN IF EXISTS physical_activity_level;
