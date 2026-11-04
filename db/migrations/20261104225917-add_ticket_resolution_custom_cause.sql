
-- +migrate Up
ALTER TABLE ticket_resolutions
ADD COLUMN custom_cause TEXT NOT NULL DEFAULT '';

ALTER TABLE ticket_resolutions
ALTER COLUMN cause_id DROP NOT NULL;

-- +migrate Down
ALTER TABLE ticket_resolutions
ALTER COLUMN cause_id SET NOT NULL;

ALTER TABLE ticket_resolutions
DROP COLUMN custom_cause;