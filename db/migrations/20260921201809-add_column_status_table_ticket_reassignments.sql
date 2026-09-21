-- +migrate Up
ALTER TABLE ticket_reassignments
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'PENDING';

ALTER TABLE ticket_reassignments
ADD CONSTRAINT ticket_reassignments_status_check
CHECK (status IN ('PENDING', 'DONE'));


-- +migrate Down
ALTER TABLE ticket_reassignments
DROP CONSTRAINT IF EXISTS ticket_reassignments_status_check;

ALTER TABLE ticket_reassignments
DROP COLUMN IF EXISTS status;