-- ============================================================================
-- Migration: 000036_add_audio_support_to_education.up.sql
-- Module:    Education Audio & Progress Tracking (3 Aspek: Artikel, Video, Audio)
-- Created:   2026-11-05
-- ============================================================================

-- 1. Add audio_url to articles
ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS audio_url TEXT;

-- 2. Add audio tracking fields to user_article_completions
ALTER TABLE user_article_completions
    ADD COLUMN IF NOT EXISTS audio_listened        BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS audio_listened_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS audio_started_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS audio_finished_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS audio_listen_duration INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS audio_last_timestamp   INT NOT NULL DEFAULT 0;
