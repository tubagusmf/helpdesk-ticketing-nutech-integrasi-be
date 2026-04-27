
-- +migrate Up
ALTER TABLE users ADD COLUMN is_online BOOLEAN DEFAULT false;

-- +migrate Down
DROP COLUMN is_online;