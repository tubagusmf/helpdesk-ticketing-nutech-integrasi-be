
-- +migrate Up
ALTER TABLE projects ADD COLUMN code_prefix VARCHAR(10);

-- +migrate Down
DROP COLUMN code_prefix;