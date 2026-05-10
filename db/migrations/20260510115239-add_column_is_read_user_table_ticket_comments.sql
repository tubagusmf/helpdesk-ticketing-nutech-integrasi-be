
-- +migrate Up
ALTER TABLE ticket_comments
ADD COLUMN is_read_by_user BOOLEAN DEFAULT FALSE,
ADD COLUMN is_read_by_staff BOOLEAN DEFAULT FALSE,
ADD COLUMN is_read_by_administrator BOOLEAN DEFAULT FALSE;

-- +migrate Down
DROP COLUMN is_read_by_user,
DROP COLUMN is_read_by_staff,
DROP COLUMN is_read_by_administrator;