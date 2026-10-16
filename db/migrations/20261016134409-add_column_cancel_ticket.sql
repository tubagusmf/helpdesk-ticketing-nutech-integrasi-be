
-- +migrate Up
ALTER TYPE ticket_status ADD VALUE IF NOT EXISTS 'CANCEL';

ALTER TABLE tickets
    ADD COLUMN IF NOT EXISTS cancellation_reason TEXT,
    ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS cancelled_by_id INT NULL;

-- +migrate Down
ALTER TABLE tickets
    DROP COLUMN IF EXISTS cancellation_reason,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS cancelled_by_id;