-- Migration: 000016_create_survey_module_tables.up.sql

-- 1. Create surveys table
CREATE TABLE IF NOT EXISTS surveys (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title              VARCHAR(200) NOT NULL,
    description        TEXT,
    type               VARCHAR(50) NOT NULL DEFAULT 'USER_SATISFACTION',
    status             VARCHAR(50) NOT NULL DEFAULT 'draft',
    is_active          BOOLEAN NOT NULL DEFAULT false,
    start_date         TIMESTAMPTZ,
    end_date           TIMESTAMPTZ,
    created_by         UUID,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_surveys_status ON surveys(status);
CREATE INDEX IF NOT EXISTS idx_surveys_type ON surveys(type);
CREATE INDEX IF NOT EXISTS idx_surveys_deleted_at ON surveys(deleted_at);

-- 2. Create survey_questions table
CREATE TABLE IF NOT EXISTS survey_questions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id          UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    question_text      TEXT NOT NULL,
    description        TEXT,
    image_url          TEXT,
    svg_illustration   TEXT,
    likert_labels      JSONB DEFAULT '["Sangat Tidak Setuju", "Tidak Setuju", "Netral", "Setuju", "Sangat Setuju"]'::jsonb,
    is_required        BOOLEAN NOT NULL DEFAULT true,
    display_order      INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_survey_questions_survey_id ON survey_questions(survey_id);

-- 3. Create survey_responses table
CREATE TABLE IF NOT EXISTS survey_responses (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id          UUID NOT NULL REFERENCES surveys(id) ON DELETE CASCADE,
    patient_id         UUID NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    started_at         TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration_seconds   INT NOT NULL DEFAULT 0,
    total_score        DOUBLE PRECISION,
    average_score      DOUBLE PRECISION,
    percentage_score   DOUBLE PRECISION,
    raw_score          DOUBLE PRECISION,
    sus_score          DOUBLE PRECISION,
    interpretation     VARCHAR(50),
    passed             BOOLEAN,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_survey_responses_survey_id ON survey_responses(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_patient_id ON survey_responses(patient_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_survey_responses_unique ON survey_responses(survey_id, patient_id) WHERE deleted_at IS NULL;

-- 4. Create survey_answers table
CREATE TABLE IF NOT EXISTS survey_answers (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    response_id        UUID NOT NULL REFERENCES survey_responses(id) ON DELETE CASCADE,
    question_id        UUID NOT NULL REFERENCES survey_questions(id) ON DELETE CASCADE,
    rating_value       INT NOT NULL,
    adjusted_score     DOUBLE PRECISION,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_survey_answers_response_id ON survey_answers(response_id);
CREATE INDEX IF NOT EXISTS idx_survey_answers_question_id ON survey_answers(question_id);
