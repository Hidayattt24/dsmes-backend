-- ============================================================================
-- Migration: 000036_add_audio_support_to_education.down.sql
-- ============================================================================

ALTER TABLE user_article_completions
    DROP COLUMN IF EXISTS audio_last_timestamp,
    DROP COLUMN IF EXISTS audio_listen_duration,
    DROP COLUMN IF EXISTS audio_finished_at,
    DROP COLUMN IF EXISTS audio_started_at,
    DROP COLUMN IF EXISTS audio_listened_at,
    DROP COLUMN IF EXISTS audio_listened;

ALTER TABLE articles
    DROP COLUMN IF EXISTS audio_url;
