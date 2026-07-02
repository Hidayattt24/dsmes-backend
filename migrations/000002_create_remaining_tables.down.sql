-- ============================================================================
-- Migration: 000002_create_remaining_tables.down.sql
-- Rolls back all tables created in migration 000002.
-- ============================================================================

DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS faqs;
DROP TABLE IF EXISTS weekly_health_summaries;
DROP TABLE IF EXISTS user_saved_articles;
DROP TABLE IF EXISTS article_views;
DROP TABLE IF EXISTS user_article_completions;
DROP TABLE IF EXISTS article_section_steps;
DROP TABLE IF EXISTS article_sections;
DROP TABLE IF EXISTS articles;
DROP TABLE IF EXISTS article_categories;
DROP TABLE IF EXISTS notification_logs;
DROP TABLE IF EXISTS system_reminder_templates;
DROP TABLE IF EXISTS daily_reminder_logs;
DROP TABLE IF EXISTS reminder_active_days;
DROP TABLE IF EXISTS reminders;
DROP TABLE IF EXISTS recent_food_searches;
DROP TABLE IF EXISTS meal_logs;
DROP TABLE IF EXISTS foods;
DROP TABLE IF EXISTS daily_medical_checkins;
DROP TABLE IF EXISTS blood_sugar_logs;
DROP TABLE IF EXISTS routine_log_entries;
DROP TABLE IF EXISTS routine_times;
DROP TABLE IF EXISTS routines;
DROP TABLE IF EXISTS patients;

DROP TYPE IF EXISTS ticket_status_enum;
DROP TYPE IF EXISTS trend_status_enum;
DROP TYPE IF EXISTS section_type_enum;
DROP TYPE IF EXISTS article_status_enum;
DROP TYPE IF EXISTS reminder_log_status_enum;
DROP TYPE IF EXISTS reminder_category_enum;
DROP TYPE IF EXISTS reminder_type_enum;
DROP TYPE IF EXISTS meal_type_enum;
DROP TYPE IF EXISTS glucose_status_enum;
DROP TYPE IF EXISTS measurement_time_enum;
DROP TYPE IF EXISTS routine_log_status_enum;
DROP TYPE IF EXISTS waktu_status_enum;
DROP TYPE IF EXISTS waktu_type_enum;
DROP TYPE IF EXISTS routine_type_enum;
DROP TYPE IF EXISTS blood_type_enum;
DROP TYPE IF EXISTS gender_enum;
