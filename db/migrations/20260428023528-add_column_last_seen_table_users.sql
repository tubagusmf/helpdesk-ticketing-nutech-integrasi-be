
-- +migrate Up
ALTER TABLE users ADD COLUMN last_seen TIMESTAMP;

-- +migrate Down
DROP COLUMN last_seen;