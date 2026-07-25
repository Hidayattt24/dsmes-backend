-- Migration: 000004_redesign_questionnaires.up.sql
-- Module: Questionnaire Redesign (Pre-Test & Post-Test, Category Grouping)

DO $$ BEGIN
    CREATE TYPE questionnaire_type_enum AS ENUM ('PRE_TEST', 'POST_TEST');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS questionnaires (
    id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title             VARCHAR(200) NOT NULL,
    type              questionnaire_type_enum NOT NULL DEFAULT 'POST_TEST',
    description       TEXT,
    education_id      UUID         REFERENCES articles(id) ON DELETE SET NULL,
    passing_score     INT,
    difficulty        VARCHAR(50),
    status            VARCHAR(50)  NOT NULL DEFAULT 'draft',
    created_by        UUID         REFERENCES staff_accounts(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_questionnaires_education ON questionnaires(education_id);
CREATE INDEX IF NOT EXISTS idx_questionnaires_type_status ON questionnaires(type, status);

CREATE TABLE IF NOT EXISTS question_categories (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    questionnaire_id UUID         NOT NULL REFERENCES questionnaires(id) ON DELETE CASCADE,
    title            VARCHAR(200) NOT NULL,
    description      TEXT,
    display_order    INT          NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_qc_questionnaire ON question_categories(questionnaire_id);

CREATE TABLE IF NOT EXISTS questions (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id      UUID         NOT NULL REFERENCES question_categories(id) ON DELETE CASCADE,
    question_text    TEXT         NOT NULL,
    explanation      TEXT,
    display_order    INT          NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_q_category ON questions(category_id);

CREATE TABLE IF NOT EXISTS question_options (
    id             UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id    UUID         NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_text    TEXT         NOT NULL,
    is_correct     BOOLEAN      NOT NULL DEFAULT false,
    display_order  INT          NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
    deleted_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_qo_question ON question_options(question_id);

-- Migration of legacy quizzes table if present
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_tables WHERE tablename = 'quizzes') THEN
        INSERT INTO questionnaires (id, title, type, description, education_id, passing_score, difficulty, status, created_by, created_at, updated_at, deleted_at)
        SELECT id, title, 'POST_TEST'::questionnaire_type_enum, NULL, linked_article_id, passing_score, difficulty, status, created_by, created_at, updated_at, deleted_at
        FROM quizzes
        ON CONFLICT (id) DO NOTHING;
    END IF;
END $$;
