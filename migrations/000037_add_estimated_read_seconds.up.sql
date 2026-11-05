-- ============================================================================
-- Migration: 000037_add_estimated_read_seconds.up.sql
-- Module:    Education Content
-- Purpose:   Add estimated_read_seconds column and populate from estimated_read_minutes
-- ============================================================================

ALTER TABLE articles
    ADD COLUMN IF NOT EXISTS estimated_read_seconds INT NOT NULL DEFAULT 60;

UPDATE articles
SET estimated_read_seconds = COALESCE(estimated_read_minutes * 60, 60)
WHERE estimated_read_seconds = 60 AND estimated_read_minutes IS NOT NULL;
