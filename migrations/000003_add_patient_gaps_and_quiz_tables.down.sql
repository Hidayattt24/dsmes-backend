-- ============================================================================
-- Migration: 000003_add_patient_gaps_and_quiz_tables.down.sql
-- Module:    Patients, Articles, and Quizzes Gaps Rollback
-- Created:   2026-07-12
-- ============================================================================

DROP TABLE IF EXISTS quiz_answers;
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quizzes;

ALTER TABLE articles DROP COLUMN IF EXISTS content;
ALTER TABLE articles DROP COLUMN IF EXISTS youtube_link;

ALTER TABLE patients DROP COLUMN IF EXISTS bpjs;
ALTER TABLE patients DROP COLUMN IF EXISTS nik;
ALTER TABLE patients DROP COLUMN IF EXISTS emergency_name;
ALTER TABLE patients DROP COLUMN IF EXISTS emergency_relation;
ALTER TABLE patients DROP COLUMN IF EXISTS emergency_phone;
ALTER TABLE patients DROP COLUMN IF EXISTS diabetes_type;
ALTER TABLE patients DROP COLUMN IF EXISTS compliance;
ALTER TABLE patients DROP COLUMN IF EXISTS intervention_type;
