ALTER TABLE notification_logs
    ADD COLUMN IF NOT EXISTS education_followup_due_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_notification_logs_education_followup_due
    ON notification_logs (education_followup_due_at)
    WHERE education_followup_due_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_article_views_article_patient_active
    ON article_views (article_id, patient_id)
    WHERE deleted_at IS NULL;
