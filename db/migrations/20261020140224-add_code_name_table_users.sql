
-- +migrate Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS code_name VARCHAR(20);

-- +migrate Down
ALTER TABLE users ALTER COLUMN code_name DROP NOT NULL;