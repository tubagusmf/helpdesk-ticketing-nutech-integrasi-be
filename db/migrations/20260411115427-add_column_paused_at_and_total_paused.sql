
-- +migrate Up
ALTER TABLE tickets
ADD COLUMN paused_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
ADD COLUMN total_paused INTEGER DEFAULT 0;

-- +migrate Down
ALTER TABLE tickets
DROP COLUMN paused_at,
DROP COLUMN total_paused;
