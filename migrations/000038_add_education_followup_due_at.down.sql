DROP INDEX IF EXISTS idx_notification_logs_education_followup_due;
DROP INDEX IF EXISTS idx_article_views_article_patient_active;

ALTER TABLE notification_logs
    DROP COLUMN IF EXISTS education_followup_due_at;
