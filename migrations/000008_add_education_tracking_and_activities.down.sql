-- ============================================================================
-- Migration: 000008_add_education_tracking_and_activities.down.sql
-- Module:    Education Details and Activities Tracking
-- Created:   2026-07-27
-- ============================================================================

DROP TABLE IF EXISTS patient_education_activities;

ALTER TABLE user_article_completions
    DROP COLUMN IF EXISTS article_started_at,
    DROP COLUMN IF EXISTS article_finished_at,
    DROP COLUMN IF EXISTS article_reading_duration,
    DROP COLUMN IF EXISTS article_last_scroll_position,
    DROP COLUMN IF EXISTS video_started_at,
    DROP COLUMN IF EXISTS video_finished_at,
    DROP COLUMN IF EXISTS video_watch_duration,
    DROP COLUMN IF EXISTS video_last_timestamp,
    DROP COLUMN IF EXISTS completion_source;
