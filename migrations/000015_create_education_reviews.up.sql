-- ============================================================================
-- Migration: 000015_create_education_reviews.up.sql
-- Module:    Education Reviews & Ratings
-- Created:   2026-08-03
-- ============================================================================

CREATE TABLE IF NOT EXISTS education_reviews (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    education_id UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    patient_id   UUID        NOT NULL REFERENCES patients(id) ON DELETE CASCADE,
    rating       INT         NOT NULL CHECK (rating >= 1 AND rating <= 5),
    note         TEXT        DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ,

    CONSTRAINT uq_education_patient_review UNIQUE (education_id, patient_id)
);

CREATE INDEX IF NOT EXISTS idx_education_reviews_education ON education_reviews(education_id);
CREATE INDEX IF NOT EXISTS idx_education_reviews_patient ON education_reviews(patient_id);
