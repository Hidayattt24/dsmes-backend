-- Migration: 000009_fix_quiz_fk_constraints.up.sql
-- Fixes FK constraints on quiz_attempts and quiz_answers that still
-- reference old tables (quizzes, quiz_questions) from migration 000003.
-- The new tables (questionnaires, questions) were created in migration 000004
-- but the FK constraints were never updated.

-- 1. quiz_attempts.quiz_id → questionnaires(id) instead of quizzes(id)
ALTER TABLE quiz_attempts DROP CONSTRAINT IF EXISTS quiz_attempts_quiz_id_fkey;
ALTER TABLE quiz_attempts
    ADD CONSTRAINT quiz_attempts_quiz_id_fkey
    FOREIGN KEY (quiz_id) REFERENCES questionnaires(id) ON DELETE CASCADE;

-- 2. quiz_answers.question_id → questions(id) instead of quiz_questions(id)
ALTER TABLE quiz_answers DROP CONSTRAINT IF EXISTS quiz_answers_question_id_fkey;
ALTER TABLE quiz_answers
    ADD CONSTRAINT quiz_answers_question_id_fkey
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE;

-- 3. Add option_id column to quiz_answers & widen selected_option column
ALTER TABLE quiz_answers ADD COLUMN IF NOT EXISTS option_id UUID REFERENCES question_options(id) ON DELETE SET NULL;
ALTER TABLE quiz_answers ALTER COLUMN selected_option TYPE VARCHAR(200);
