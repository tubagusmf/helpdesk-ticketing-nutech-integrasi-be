
-- +migrate Up
ALTER TABLE users
ADD COLUMN IF NOT EXISTS last_heartbeat_at TIMESTAMPTZ NULL;

UPDATE users
SET last_heartbeat_at = COALESCE(last_seen, NOW())
WHERE is_online = TRUE
  AND last_heartbeat_at IS NULL;

-- +migrate Down
DROP COLUMN IF EXISTS last_heartbeat_at;