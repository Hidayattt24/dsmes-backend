ALTER TABLE reminders
    ADD COLUMN IF NOT EXISTS system_key VARCHAR(50),
    ADD COLUMN IF NOT EXISTS action_type VARCHAR(50) NOT NULL DEFAULT 'reminder_management';

CREATE UNIQUE INDEX IF NOT EXISTS idx_reminders_patient_system_key
    ON reminders(patient_id, system_key)
    WHERE system_key IS NOT NULL;

ALTER TABLE notification_logs
    ADD COLUMN IF NOT EXISTS action_type VARCHAR(50) NOT NULL DEFAULT 'reminder_management',
    ADD COLUMN IF NOT EXISTS occurrence_id VARCHAR(150),
    ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_notification_logs_occurrence
    ON notification_logs(occurrence_id)
    WHERE occurrence_id IS NOT NULL;

UPDATE reminders
SET action_type = CASE
    WHEN icon_name = 'blood' OR LOWER(activity_name) LIKE '%gula%' THEN 'blood_sugar_entry'
    WHEN icon_name = 'medicine' OR LOWER(activity_name) LIKE '%obat%' OR LOWER(activity_name) LIKE '%insulin%' THEN 'medication_entry'
    WHEN icon_name IN ('walk', 'run', 'exercise', 'bike', 'yoga') OR LOWER(activity_name) LIKE '%olahraga%' THEN 'activity_entry'
    WHEN icon_name IN ('breakfast', 'lunch', 'dinner', 'restaurant') OR LOWER(activity_name) LIKE '%makan%' THEN 'meal_entry'
    ELSE 'reminder_management'
END
WHERE action_type = 'reminder_management';

UPDATE notification_logs nl
SET action_type = COALESCE(NULLIF(r.action_type, ''), 'reminder_management')
FROM reminders r
WHERE nl.reminder_id = r.id AND nl.notif_type = 'reminder';

UPDATE notification_logs
SET action_type = 'education_detail'
WHERE notif_type = 'education';
