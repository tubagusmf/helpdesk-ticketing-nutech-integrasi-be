
-- +migrate Up
ALTER TABLE problems
ADD COLUMN part_id INTEGER NULL REFERENCES parts(id);

CREATE INDEX idx_problems_part_id
ON problems(part_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_problems_part_id;

ALTER TABLE problems
DROP COLUMN IF EXISTS part_id;