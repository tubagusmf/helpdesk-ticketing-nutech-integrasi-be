
-- +migrate Up
ALTER TABLE ticket_engineer_resolutions
ADD COLUMN is_read BOOLEAN NOT NULL DEFAULT FALSE;

-- +migrate Down
ALTER TABLE ticket_engineer_resolutions
DROP COLUMN is_read;