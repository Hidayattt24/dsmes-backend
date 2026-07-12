-- ============================================================================
-- Migration: 000002_create_remaining_tables.up.sql
-- Module:    Remaining DSMES Tables
-- Created:   2026-07-02
-- ============================================================================

-- ── Remaining Enum types ─────────────────────────────────────────────────────
CREATE TYPE gender_enum              AS ENUM ('laki_laki', 'perempuan');
CREATE TYPE blood_type_enum          AS ENUM ('A', 'B', 'AB', 'O', 'tidak_tahu');
CREATE TYPE routine_type_enum        AS ENUM ('Jalan_Pagi', 'Minum_Air', 'Cek_Gula');
CREATE TYPE waktu_type_enum          AS ENUM ('Default', 'Kustom');
CREATE TYPE waktu_status_enum        AS ENUM ('Set', 'Unset');
CREATE TYPE routine_log_status_enum  AS ENUM ('Completed', 'Skipped', 'Pending');
CREATE TYPE measurement_time_enum    AS ENUM ('sebelum_makan', 'sesudah_makan', 'sewaktu');
CREATE TYPE glucose_status_enum      AS ENUM ('rendah', 'normal', 'tinggi', 'sangat_tinggi');
CREATE TYPE meal_type_enum           AS ENUM ('sarapan', 'makan_siang', 'makan_malam', 'camilan');
CREATE TYPE reminder_type_enum       AS ENUM ('sistem', 'personal');
CREATE TYPE reminder_category_enum   AS ENUM ('medis_obat', 'nutrisi_air', 'aktivitas_fisik', 'lainnya');
CREATE TYPE reminder_log_status_enum AS ENUM ('selesai', 'terlewat', 'pending');
CREATE TYPE article_status_enum      AS ENUM ('draft', 'publikasi');
CREATE TYPE section_type_enum        AS ENUM ('paragraf', 'langkah', 'info_penting');
CREATE TYPE trend_status_enum        AS ENUM ('stabil', 'meningkat', 'menurun');
CREATE TYPE ticket_status_enum       AS ENUM ('open', 'in_progress', 'closed');

-- ── Patients ─────────────────────────────────────────────────────────────────
CREATE TABLE patients (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    email                  VARCHAR(150) UNIQUE NOT NULL,
    password_hash          VARCHAR(255) NOT NULL,
    full_name              VARCHAR(150) NOT NULL,
    nickname               VARCHAR(50),
    whatsapp_number        VARCHAR(20)  NOT NULL,
    gender                 gender_enum  NOT NULL,
    date_of_birth          DATE         NOT NULL,
    height_cm              NUMERIC(5,2),
    weight_kg              NUMERIC(5,2),
    blood_type             blood_type_enum DEFAULT 'tidak_tahu',
    daily_calorie_target   INT          NOT NULL DEFAULT 2000,
    medical_status         VARCHAR(100),
    profile_photo_url      TEXT,
    status                 account_status_enum NOT NULL DEFAULT 'aktif',
    assigned_staff_id      UUID         REFERENCES staff_accounts(id) ON DELETE SET NULL,
    last_active_at         TIMESTAMPTZ,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX idx_patients_staff      ON patients(assigned_staff_id);
CREATE INDEX idx_patients_deleted   ON patients(deleted_at) WHERE deleted_at IS NOT NULL;

-- ── Routines ─────────────────────────────────────────────────────────────────
CREATE TABLE routines (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id         UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    routine_type       routine_type_enum NOT NULL,
    descriptive_name   VARCHAR(150),
    base_frequency     VARCHAR(50)  NOT NULL,
    is_active          BOOLEAN      NOT NULL DEFAULT true,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX idx_routines_patient ON routines(patient_id);

-- ── Routine Times ────────────────────────────────────────────────────────────
CREATE TABLE routine_times (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    routine_id       UUID         NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    time_type        waktu_type_enum NOT NULL,
    scheduled_time   TIME,
    status           waktu_status_enum NOT NULL DEFAULT 'Unset',
    reminder_active  BOOLEAN      NOT NULL DEFAULT false,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX idx_routine_times_routine ON routine_times(routine_id);

-- ── Routine Log Entries ──────────────────────────────────────────────────────
CREATE TABLE routine_log_entries (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id        UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    routine_time_id   UUID         NOT NULL REFERENCES routine_times(id) ON DELETE CASCADE,
    logged_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    status            routine_log_status_enum NOT NULL DEFAULT 'Pending',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX idx_routine_logs_patient ON routine_log_entries(patient_id);
CREATE INDEX idx_routine_logs_time    ON routine_log_entries(routine_time_id);

-- ── Blood Sugar Logs ─────────────────────────────────────────────────────────
CREATE TABLE blood_sugar_logs (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id              UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    glucose_value           INT          NOT NULL,
    measurement_time_type   measurement_time_enum NOT NULL,
    measured_at             TIMESTAMPTZ  NOT NULL,
    status                  glucose_status_enum NOT NULL,
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_bsl_patient_time ON blood_sugar_logs(patient_id, measured_at DESC);

-- ── Daily Medical Checkins ───────────────────────────────────────────────────
CREATE TABLE daily_medical_checkins (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id    UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    checkin_date  DATE         NOT NULL,
    is_completed  BOOLEAN      NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    UNIQUE(patient_id, checkin_date)
);

-- ── Foods ────────────────────────────────────────────────────────────────────
CREATE TABLE foods (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name                   VARCHAR(150) NOT NULL,
    default_serving_unit   VARCHAR(50),
    default_serving_grams  NUMERIC(6,2),
    calories               NUMERIC(6,2) NOT NULL,
    carbs_g                NUMERIC(6,2),
    protein_g              NUMERIC(6,2),
    fat_g                  NUMERIC(6,2),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX idx_foods_name ON foods(name);

-- ── Meal Logs ────────────────────────────────────────────────────────────────
CREATE TABLE meal_logs (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id          UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    food_id             UUID         NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    meal_type           meal_type_enum NOT NULL,
    portion_multiplier  NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    logged_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_meal_logs_patient_date ON meal_logs(patient_id, logged_at);

-- ── Recent Food Searches ─────────────────────────────────────────────────────
CREATE TABLE recent_food_searches (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id     UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    food_id        UUID         NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    usage_count    INT          NOT NULL DEFAULT 1,
    last_used_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE(patient_id, food_id)
);

-- ── Reminders ────────────────────────────────────────────────────────────────
CREATE TABLE reminders (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id             UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    activity_name          VARCHAR(150) NOT NULL,
    reminder_type          reminder_type_enum NOT NULL,
    category               reminder_category_enum NOT NULL,
    scheduled_time         TIME         NOT NULL,
    is_active              BOOLEAN      NOT NULL DEFAULT true,
    notes                  TEXT,
    repeat_interval_days   INT          DEFAULT 1,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX idx_reminders_patient ON reminders(patient_id);

-- ── Reminder Active Days ─────────────────────────────────────────────────────
CREATE TABLE reminder_active_days (
    reminder_id  UUID         NOT NULL REFERENCES reminders(id) ON DELETE CASCADE,
    day_of_week  SMALLINT     NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    PRIMARY KEY (reminder_id, day_of_week)
);

-- ── Daily Reminder Logs ──────────────────────────────────────────────────────
CREATE TABLE daily_reminder_logs (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    reminder_id   UUID         NOT NULL REFERENCES reminders(id) ON DELETE CASCADE,
    log_date      DATE         NOT NULL,
    status        reminder_log_status_enum NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    UNIQUE(reminder_id, log_date)
);

-- ── System Reminder Templates ────────────────────────────────────────────────
CREATE TABLE system_reminder_templates (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_name      VARCHAR(150) NOT NULL,
    category           reminder_category_enum NOT NULL,
    default_time       TIME,
    default_frequency  VARCHAR(50),
    description        TEXT,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ
);

-- ── Notification Logs ────────────────────────────────────────────────────────
CREATE TABLE notification_logs (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    reminder_id    UUID         REFERENCES reminders(id) ON DELETE SET NULL,
    patient_id     UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    message_text   TEXT         NOT NULL,
    notified_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    is_read        BOOLEAN      NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_notification_patient_unread ON notification_logs(patient_id) WHERE is_read = false;

-- ── Article Categories ───────────────────────────────────────────────────────
CREATE TABLE article_categories (
    id    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name  VARCHAR(50)  UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

-- ── Articles ─────────────────────────────────────────────────────────────────
CREATE TABLE articles (
    id                       UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title                    VARCHAR(200) NOT NULL,
    category_id              UUID         NOT NULL REFERENCES article_categories(id) ON DELETE RESTRICT,
    estimated_read_minutes   INT,
    author_name              VARCHAR(150),
    banner_image_url         TEXT,
    summary                  TEXT,
    status                   article_status_enum NOT NULL DEFAULT 'draft',
    created_by               UUID         REFERENCES staff_accounts(id) ON DELETE SET NULL,
    created_at               TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at               TIMESTAMPTZ
);

CREATE INDEX idx_articles_category ON articles(category_id);

-- ── Article Sections ─────────────────────────────────────────────────────────
CREATE TABLE article_sections (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id     UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    section_order  INT          NOT NULL,
    section_title  VARCHAR(200),
    section_type   section_type_enum NOT NULL,
    content_text   TEXT,
    image_url      TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX idx_article_sections_article ON article_sections(article_id);

-- ── Article Section Steps ────────────────────────────────────────────────────
CREATE TABLE article_section_steps (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id   UUID         NOT NULL REFERENCES article_sections(id) ON DELETE CASCADE,
    step_order   INT          NOT NULL,
    step_text    TEXT         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_ass_section ON article_section_steps(section_id);

-- ── User Article Completions ─────────────────────────────────────────────────
CREATE TABLE user_article_completions (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id     UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    article_id     UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    completed_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE(patient_id, article_id)
);

-- ── Article Views ────────────────────────────────────────────────────────────
CREATE TABLE article_views (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id    UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    patient_id    UUID         REFERENCES patients(id) ON DELETE SET NULL,
    viewed_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ
);

CREATE INDEX idx_article_views_article ON article_views(article_id);

-- ── User Saved Articles ──────────────────────────────────────────────────────
CREATE TABLE user_saved_articles (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id    UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    article_id    UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    saved_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    UNIQUE(patient_id, article_id)
);

-- ── Weekly Health Summaries ──────────────────────────────────────────────────
CREATE TABLE weekly_health_summaries (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id             UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    week_start_date        DATE         NOT NULL,
    week_end_date          DATE         NOT NULL,
    blood_sugar_trend      trend_status_enum,
    avg_blood_sugar        NUMERIC(6,2),
    articles_read_count    INT          DEFAULT 0,
    articles_target_count  INT          DEFAULT 7,
    generated_at           TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ,
    UNIQUE(patient_id, week_start_date)
);

-- ── FAQs ─────────────────────────────────────────────────────────────────────
CREATE TABLE faqs (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    question       TEXT         NOT NULL,
    answer         TEXT         NOT NULL,
    display_order  INT          DEFAULT 0,
    is_active      BOOLEAN      DEFAULT true,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

-- ── Support Tickets ──────────────────────────────────────────────────────────
CREATE TABLE support_tickets (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id    UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    subject       VARCHAR(200) NOT NULL,
    message       TEXT         NOT NULL,
    status        ticket_status_enum NOT NULL DEFAULT 'open',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    resolved_at   TIMESTAMPTZ
);

-- Seed initial User (Patient) account (password is 'password123')
INSERT INTO patients (id, email, password_hash, full_name, nickname, whatsapp_number, gender, date_of_birth, height_cm, weight_kg, blood_type, daily_calorie_target, status)
VALUES 
    ('f9a74a10-2fbe-443b-8cf7-0d5db6509f6e', 'patient@dsmes.com', '$2a$10$w850aK9B74hC4s/5Z1yOSeKpeKkI2wL3121X77.i.J7cO7N0l/Jb.', 'Jane Doe', 'Jane', '081234567890', 'perempuan', '1990-01-01', 165.0, 55.0, 'O', 2000, 'aktif')
ON CONFLICT DO NOTHING;

