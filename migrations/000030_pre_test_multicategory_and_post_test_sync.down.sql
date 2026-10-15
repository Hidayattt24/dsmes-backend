-- Migration: 000030_pre_test_multicategory_and_post_test_sync.down.sql

ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS score_pre_test;
ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS dsmqr_category;
ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS score_dsmqr;
ALTER TABLE quiz_attempts DROP COLUMN IF EXISTS score_dsmes;
ALTER TABLE question_categories DROP COLUMN IF EXISTS instrument_category;
ALTER TABLE questionnaires DROP COLUMN IF EXISTS is_linked_to_pre_test;
ALTER TABLE questionnaires DROP COLUMN IF EXISTS source_pre_test_id;
