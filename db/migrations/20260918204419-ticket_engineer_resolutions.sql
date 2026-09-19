
-- +migrate Up
CREATE TABLE ticket_engineer_resolutions (
    id SERIAL PRIMARY KEY,
    ticket_id INT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    engineer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    solution TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_engineer_resolutions_ticket_id
    ON ticket_engineer_resolutions(ticket_id);

CREATE INDEX IF NOT EXISTS idx_ticket_engineer_resolutions_engineer_id
    ON ticket_engineer_resolutions(engineer_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_ticket_engineer_resolutions_ticket_id;
DROP INDEX IF EXISTS idx_ticket_engineer_resolutions_engineer_id;
DROP TABLE IF EXISTS ticket_engineer_resolutions;