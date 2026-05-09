
-- +migrate Up
ALTER TABLE users ADD COLUMN last_ticket_assigned_at TIMESTAMPTZ NULL;

-- +migrate Down
ALTER TABLE users DROP COLUMN last_ticket_assigned_at;
