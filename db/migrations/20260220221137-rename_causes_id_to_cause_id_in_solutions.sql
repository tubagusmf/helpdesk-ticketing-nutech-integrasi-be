
-- +migrate Up
ALTER TABLE solutions RENAME COLUMN causes_id TO cause_id;

-- +migrate Down
ALTER TABLE solutions RENAME COLUMN cause_id TO causes_id;
