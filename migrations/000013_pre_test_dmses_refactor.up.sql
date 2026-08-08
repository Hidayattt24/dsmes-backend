-- Migration: 000013_pre_test_dmses_refactor.up.sql
-- Module: Quiz / Questionnaire (DMSES Pre-Test Refactor)

-- 1. Add question_image_url to questions table
ALTER TABLE questions ADD COLUMN IF NOT EXISTS question_image_url TEXT;

-- 2. Make category_id optional in questions table for PRE_TEST questions
ALTER TABLE questions ALTER COLUMN category_id DROP NOT NULL;

-- 3. Add self_efficacy_category to quiz_attempts table
ALTER TABLE quiz_attempts ADD COLUMN IF NOT EXISTS self_efficacy_category VARCHAR(50);

-- 4. Add selected_value to quiz_answers table (stores 1-5 integer scale for PRE_TEST)
ALTER TABLE quiz_answers ADD COLUMN IF NOT EXISTS selected_value INT;
ALTER TABLE quiz_answers ALTER COLUMN selected_option DROP NOT NULL;
