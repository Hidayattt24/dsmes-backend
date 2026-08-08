-- ============================================================================
-- Migration: 000014_add_notif_type_to_notification_logs.up.sql
-- Module:    Notification Logs
-- Created:   2026-08-03
-- ============================================================================

ALTER TABLE notification_logs ADD COLUMN IF NOT EXISTS notif_type VARCHAR(50) DEFAULT 'reminder';
ALTER TABLE notification_logs ADD COLUMN IF NOT EXISTS article_id UUID REFERENCES articles(id) ON DELETE SET NULL;
