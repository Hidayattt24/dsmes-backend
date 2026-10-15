-- Migration: 000030_pre_test_multicategory_and_post_test_sync.up.sql
-- Modules: Quiz / Questionnaire (Multi-Category Pre-Test: PRE_TEST, DSMQR, DSMES + Dynamic Choices + Post-Test sync)

-- 1. Tambah scope & source_pre_test_id ke questionnaires untuk relasi Pre-Test -> Post-Test
ALTER TABLE questionnaires ADD COLUMN IF NOT EXISTS scope VARCHAR(50) NOT NULL DEFAULT 'EDUCATION';
ALTER TABLE questionnaires ADD COLUMN IF NOT EXISTS source_pre_test_id UUID REFERENCES questionnaires(id) ON DELETE SET NULL;
ALTER TABLE questionnaires ADD COLUMN IF NOT EXISTS is_linked_to_pre_test BOOLEAN NOT NULL DEFAULT FALSE;

-- 2. Tambah kategori instrumen pada question_categories (PRE_TEST, DSMQR, DSMES)
ALTER TABLE question_categories ADD COLUMN IF NOT EXISTS instrument_category VARCHAR(50) DEFAULT 'PRE_TEST';

-- 3. Tambah kolom skor terpisah pada quiz_attempts (DSMQR & DSMES)
ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS score_dsmes INT;
ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS score_dsmqr NUMERIC(6, 2);
ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS dsmqr_category VARCHAR(100);
ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS score_pre_test INT;
