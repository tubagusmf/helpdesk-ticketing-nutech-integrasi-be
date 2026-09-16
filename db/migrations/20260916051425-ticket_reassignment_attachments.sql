
-- +migrate Up
CREATE TABLE ticket_reassignment_attachments (
    id SERIAL PRIMARY KEY,
    reassignment_id INT NOT NULL REFERENCES ticket_reassignments(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url TEXT NOT NULL,
    public_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE ticket_reassignment_attachments;