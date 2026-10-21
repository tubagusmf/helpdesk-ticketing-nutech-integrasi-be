
-- +migrate Up
ALTER TABLE ticket_resolutions
    ALTER COLUMN solution_id DROP NOT NULL;

ALTER TABLE ticket_resolutions
    ADD COLUMN IF NOT EXISTS custom_solution TEXT NOT NULL DEFAULT '';

-- +migrate Down
ALTER TABLE ticket_resolutions
    ALTER COLUMN solution_id SET NOT NULL;

ALTER TABLE ticket_resolutions
    DROP COLUMN IF EXISTS custom_solution;
