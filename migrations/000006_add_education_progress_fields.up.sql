-- ============================================================================
-- Migration: 000006_add_education_progress_fields.up.sql
-- Module:    Education Progress Tracking
-- Created:   2026-07-25
-- ============================================================================

-- ── 1. Extend user_article_completions with separate media tracking ────────
-- This supports the research requirement to know HOW a patient consumed
-- educational material: article reading vs YouTube video watching.
ALTER TABLE user_article_completions
    ADD COLUMN IF NOT EXISTS article_read        BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS article_read_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS youtube_watched     BOOLEAN     NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS youtube_watched_at  TIMESTAMPTZ;

-- Make completed_at nullable since completion depends on OR logic
ALTER TABLE user_article_completions
    ALTER COLUMN completed_at DROP NOT NULL;

-- ── 2. Backfill existing completion records ─────────────────────────────────
-- All records created before this migration represent article completions.
UPDATE user_article_completions
SET
    article_read    = TRUE,
    article_read_at = completed_at
WHERE
    article_read = FALSE
    AND completed_at IS NOT NULL;
