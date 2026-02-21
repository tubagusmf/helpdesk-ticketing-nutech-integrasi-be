
-- +migrate Up
CREATE TABLE ticket_resolutions (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id),
    cause_id INTEGER NOT NULL REFERENCES causes(id),
    solution_id INTEGER NOT NULL REFERENCES solutions(id),
    resolution_notes TEXT NULL,
    completion_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    attachment_url TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE ticket_resolutions;