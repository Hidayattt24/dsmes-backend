-- ============================================================================
-- Migration: 000006_add_education_progress_fields.down.sql
-- Module:    Education Progress Tracking Rollback
-- Created:   2026-07-25
-- ============================================================================

ALTER TABLE user_article_completions
    ALTER COLUMN completed_at SET NOT NULL;

ALTER TABLE user_article_completions
    DROP COLUMN IF EXISTS youtube_watched_at,
    DROP COLUMN IF EXISTS youtube_watched,
    DROP COLUMN IF EXISTS article_read_at,
    DROP COLUMN IF EXISTS article_read;
