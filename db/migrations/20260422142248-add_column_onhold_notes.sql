
-- +migrate Up
ALTER TABLE tickets
ADD COLUMN onhold_notes TEXT NULL;

-- +migrate Down
DROP COLUMN onhold_notes;