
-- +migrate Up
CREATE TABLE ticket_engineer_resolution_attachments (
    id SERIAL PRIMARY KEY,
    resolution_id INT NOT NULL REFERENCES ticket_engineer_resolutions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    public_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_engineer_resolution_attachments_resolution_id
    ON ticket_engineer_resolution_attachments(resolution_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_ticket_engineer_resolution_attachments_resolution_id;
DROP TABLE IF EXISTS ticket_engineer_resolution_attachments;
