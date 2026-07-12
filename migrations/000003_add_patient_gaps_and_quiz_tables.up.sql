-- ============================================================================
-- Migration: 000003_add_patient_gaps_and_quiz_tables.up.sql
-- Module:    Patients, Articles, and Quizzes Gaps
-- Created:   2026-07-12
-- ============================================================================

-- ── 1. Update Patients table columns ─────────────────────────────────────────
ALTER TABLE patients ADD COLUMN IF NOT EXISTS bpjs VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS nik VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS emergency_name VARCHAR(150);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS emergency_relation VARCHAR(100);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS emergency_phone VARCHAR(20);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS diabetes_type VARCHAR(50);
ALTER TABLE patients ADD COLUMN IF NOT EXISTS compliance INT DEFAULT 0;
ALTER TABLE patients ADD COLUMN IF NOT EXISTS intervention_type VARCHAR(50);

-- ── 2. Update Articles table columns ─────────────────────────────────────────
ALTER TABLE articles ADD COLUMN IF NOT EXISTS content TEXT;
ALTER TABLE articles ADD COLUMN IF NOT EXISTS youtube_link VARCHAR(255);

-- ── 3. Quizzes ───────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quizzes (
    id                     UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title                  VARCHAR(200) NOT NULL,
    linked_article_id      UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    difficulty             VARCHAR(50)  NOT NULL DEFAULT 'Sedang',
    passing_score          INT          NOT NULL DEFAULT 80,
    status                 VARCHAR(50)  NOT NULL DEFAULT 'draft',
    created_by             UUID         REFERENCES staff_accounts(id) ON DELETE SET NULL,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at             TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_quizzes_article ON quizzes(linked_article_id);

-- ── 4. Quiz Questions ────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz_questions (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id        UUID         NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text  TEXT         NOT NULL,
    option_a       TEXT         NOT NULL,
    option_b       TEXT         NOT NULL,
    option_c       TEXT         NOT NULL,
    option_d       TEXT         NOT NULL,
    correct_option VARCHAR(10)  NOT NULL,
    explanation    TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_qq_quiz ON quiz_questions(quiz_id);

-- ── 5. Quiz Attempts ─────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id           UUID         NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    patient_id        UUID         NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    score             INT          NOT NULL,
    passed            BOOLEAN      NOT NULL,
    duration_seconds  INT          NOT NULL,
    completed_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_qa_quiz_patient ON quiz_attempts(quiz_id, patient_id);

-- ── 6. Quiz Answers ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS quiz_answers (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id      UUID         NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id     UUID         NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    selected_option VARCHAR(10)  NOT NULL,
    is_correct      BOOLEAN      NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_qans_attempt ON quiz_answers(attempt_id);
