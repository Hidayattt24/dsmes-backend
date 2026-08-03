-- ============================================================================
-- Migration: 000014_add_notif_type_to_notification_logs.down.sql
-- Module:    Notification Logs
-- Created:   2026-08-03
-- ============================================================================

ALTER TABLE notification_logs DROP COLUMN IF EXISTS article_id;
ALTER TABLE notification_logs DROP COLUMN IF EXISTS notif_type;
