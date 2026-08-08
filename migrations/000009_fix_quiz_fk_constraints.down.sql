-- Migration: 000009_fix_quiz_fk_constraints.down.sql
-- Reverts FK constraints to point back to old tables.

-- 1. quiz_attempts.quiz_id → quizzes(id)
ALTER TABLE quiz_attempts DROP CONSTRAINT IF EXISTS quiz_attempts_quiz_id_fkey;
ALTER TABLE quiz_attempts
    ADD CONSTRAINT quiz_attempts_quiz_id_fkey
    FOREIGN KEY (quiz_id) REFERENCES quizzes(id) ON DELETE CASCADE;

-- 2. quiz_answers.question_id → quiz_questions(id)
ALTER TABLE quiz_answers DROP CONSTRAINT IF EXISTS quiz_answers_question_id_fkey;
ALTER TABLE quiz_answers
    ADD CONSTRAINT quiz_answers_question_id_fkey
    FOREIGN KEY (question_id) REFERENCES quiz_questions(id) ON DELETE CASCADE;

-- 3. Drop option_id column
ALTER TABLE quiz_answers DROP COLUMN IF EXISTS option_id;
