
-- +migrate Up
ALTER TABLE problems
DROP CONSTRAINT IF EXISTS idx_problems_name_active;

DROP INDEX IF EXISTS idx_problems_name_active;

CREATE UNIQUE INDEX idx_problems_name_part_active
ON problems (LOWER(name), part_id)
WHERE deleted_at IS NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_problems_name_part_active;
