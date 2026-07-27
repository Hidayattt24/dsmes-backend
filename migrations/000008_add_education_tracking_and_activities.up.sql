-- ============================================================================
-- Migration: 000008_add_education_tracking_and_activities.up.sql
-- Module:    Education Details and Activities Tracking
-- Created:   2026-07-27
-- ============================================================================

-- ── 1. Extend user_article_completions with detail metrics ──────────────────
ALTER TABLE user_article_completions
    ADD COLUMN IF NOT EXISTS article_started_at          TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS article_finished_at         TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS article_reading_duration    INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS article_last_scroll_position INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS video_started_at            TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS video_finished_at           TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS video_watch_duration        INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS video_last_timestamp        INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completion_source           VARCHAR(20);

-- ── 2. Create patient_education_activities audit log table ──────────────────
CREATE TABLE IF NOT EXISTS patient_education_activities (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id    UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    article_id    UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    activity_type VARCHAR(50)  NOT NULL,
    metadata      JSONB,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pea_patient  ON patient_education_activities(patient_id);
CREATE INDEX IF NOT EXISTS idx_pea_article  ON patient_education_activities(article_id);
