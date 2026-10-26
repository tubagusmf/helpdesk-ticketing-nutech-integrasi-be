
-- +migrate Up
ALTER TABLE users
ALTER COLUMN last_seen TYPE TIMESTAMPTZ
USING last_seen AT TIME ZONE 'Asia/Jakarta';

-- +migrate Down
ALTER TABLE users
ALTER COLUMN last_seen TYPE TIMESTAMP WITHOUT TIME ZONE
USING last_seen AT TIME ZONE 'Asia/Jakarta';