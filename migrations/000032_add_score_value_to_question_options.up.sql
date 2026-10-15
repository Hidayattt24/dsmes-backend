ALTER TABLE question_options
    ADD COLUMN IF NOT EXISTS score_value NUMERIC(8, 2);
