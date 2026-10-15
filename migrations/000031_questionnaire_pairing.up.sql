-- Repair schema drift on databases where migration 000030 was recorded before
-- these questionnaire columns were added to that migration file.
ALTER TABLE questionnaires
    ADD COLUMN IF NOT EXISTS scope VARCHAR(50);

ALTER TABLE questionnaires
    ADD COLUMN IF NOT EXISTS source_pre_test_id UUID
    REFERENCES questionnaires(id) ON DELETE SET NULL;

ALTER TABLE questionnaires
    ADD COLUMN IF NOT EXISTS is_linked_to_pre_test BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE question_categories
    ADD COLUMN IF NOT EXISTS instrument_category VARCHAR(50) DEFAULT 'PRE_TEST';

ALTER TABLE quiz_attempts
    ADD COLUMN IF NOT EXISTS score_dsmes INT;

ALTER TABLE quiz_attempts
    ADD COLUMN IF NOT EXISTS score_dsmqr NUMERIC(6, 2);

ALTER TABLE quiz_attempts
    ADD COLUMN IF NOT EXISTS dsmqr_category VARCHAR(100);

ALTER TABLE quiz_attempts
    ADD COLUMN IF NOT EXISTS score_pre_test INT;

-- Backfill scope from education linkage before enforcing questionnaire pairing.
UPDATE questionnaires
SET scope = CASE WHEN education_id IS NULL THEN 'APPLICATION' ELSE 'EDUCATION' END
WHERE scope IS NULL
   OR scope NOT IN ('APPLICATION', 'EDUCATION');

ALTER TABLE questionnaires
    ALTER COLUMN scope SET DEFAULT 'EDUCATION';

ALTER TABLE questionnaires
    ALTER COLUMN scope SET NOT NULL;

-- Preserve the oldest live pair when legacy data links multiple post-tests.
WITH duplicate_links AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY source_pre_test_id ORDER BY created_at, id) AS position
    FROM questionnaires
    WHERE source_pre_test_id IS NOT NULL
      AND type = 'POST_TEST'
      AND deleted_at IS NULL
)
UPDATE questionnaires AS q
SET source_pre_test_id = NULL,
    is_linked_to_pre_test = FALSE
FROM duplicate_links AS d
WHERE q.id = d.id
  AND d.position > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_questionnaires_one_post_per_pre
    ON questionnaires (source_pre_test_id)
    WHERE type = 'POST_TEST' AND source_pre_test_id IS NOT NULL AND deleted_at IS NULL;
