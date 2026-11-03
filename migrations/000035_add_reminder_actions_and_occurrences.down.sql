DROP INDEX IF EXISTS idx_notification_logs_occurrence;
ALTER TABLE notification_logs
    DROP COLUMN IF EXISTS scheduled_at,
    DROP COLUMN IF EXISTS occurrence_id,
    DROP COLUMN IF EXISTS action_type;

DROP INDEX IF EXISTS idx_reminders_patient_system_key;
ALTER TABLE reminders
    DROP COLUMN IF EXISTS action_type,
    DROP COLUMN IF EXISTS system_key;
