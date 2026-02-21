
-- +migrate Up
ALTER TABLE tickets
RENAME COLUMN roselved_at TO resolved_at;

-- +migrate Down
ALTER TABLE tickets
RENAME COLUMN resolved_at TO roselved_at;