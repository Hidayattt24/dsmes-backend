-- Migration: 000012_auth_phone_number_primary.up.sql
-- Module: Auth + Patient
-- Primary Identity: phone_number (NOT NULL, UNIQUE, INDEXED)
-- Optional Identity: email (NULL, UNIQUE WHERE email IS NOT NULL)

-- 1. Add phone_number column if not exists
ALTER TABLE patients ADD COLUMN IF NOT EXISTS phone_number VARCHAR(20);

-- 2. Populate phone_number from whatsapp_number for existing records if null
UPDATE patients 
SET phone_number = whatsapp_number 
WHERE phone_number IS NULL OR phone_number = '';

-- 3. Normalize existing phone_number values to standard format (628...)
UPDATE patients
SET phone_number = CASE
    WHEN phone_number LIKE '0%' THEN '62' || SUBSTRING(phone_number FROM 2)
    WHEN phone_number LIKE '8%' THEN '62' || phone_number
    WHEN phone_number LIKE '+62%' THEN SUBSTRING(phone_number FROM 2)
    ELSE phone_number
END
WHERE phone_number IS NOT NULL AND phone_number <> '';

-- 4. Clean up duplicate active phone_number entries (soft-delete older duplicates)
UPDATE patients
SET deleted_at = NOW()
WHERE id IN (
    SELECT id FROM (
        SELECT id, ROW_NUMBER() OVER (PARTITION BY phone_number ORDER BY created_at DESC) as rnum
        FROM patients
        WHERE deleted_at IS NULL AND phone_number IS NOT NULL AND phone_number <> ''
    ) t
    WHERE t.rnum > 1
);

-- 5. Enforce NOT NULL on phone_number
ALTER TABLE patients ALTER COLUMN phone_number SET NOT NULL;

-- 6. Create unique index on phone_number
DROP INDEX IF EXISTS idx_patients_phone_number;
CREATE UNIQUE INDEX idx_patients_phone_number ON patients(phone_number) WHERE deleted_at IS NULL;

-- 7. Make email nullable
ALTER TABLE patients ALTER COLUMN email DROP NOT NULL;

-- 8. Drop existing email unique constraint / index if exists
ALTER TABLE patients DROP CONSTRAINT IF EXISTS patients_email_key;
DROP INDEX IF EXISTS idx_patients_email;

-- 9. Create partial unique index on email
CREATE UNIQUE INDEX idx_patients_email ON patients(email) WHERE email IS NOT NULL AND email <> '' AND deleted_at IS NULL;
