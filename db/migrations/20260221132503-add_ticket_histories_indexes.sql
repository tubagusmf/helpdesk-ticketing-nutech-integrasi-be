
-- +migrate Up
CREATE INDEX idx_ticket_histories_ticket_id
ON ticket_histories(ticket_id);

CREATE INDEX idx_ticket_histories_user_id
ON ticket_histories(user_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_ticket_histories_ticket_id;
DROP INDEX IF EXISTS idx_ticket_histories_user_id;