-- Migration: 000013_pre_test_dmses_refactor.down.sql

ALTER TABLE quiz_answers DROP COLUMN IF EXISTS selected_value;
ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS self_efficacy_category;
ALTER TABLE questions DROP COLUMN IF EXISTS question_image_url;
