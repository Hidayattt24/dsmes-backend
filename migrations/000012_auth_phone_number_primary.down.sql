-- Migration: 000012_auth_phone_number_primary.down.sql

DROP INDEX IF EXISTS idx_patients_email;
CREATE UNIQUE INDEX idx_patients_email ON patients(email);

DROP INDEX IF EXISTS idx_patients_phone_number;
ALTER TABLE patients ALTER COLUMN email SET NOT NULL;
ALTER TABLE patients DROP COLUMN IF EXISTS phone_number;
