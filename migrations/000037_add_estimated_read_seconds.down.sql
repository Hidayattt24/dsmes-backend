-- ============================================================================
-- Migration: 000037_add_estimated_read_seconds.down.sql
-- ============================================================================

ALTER TABLE articles
    DROP COLUMN IF EXISTS estimated_read_seconds;
