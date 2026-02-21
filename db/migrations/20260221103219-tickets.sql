
-- +migrate Up
CREATE TYPE ticket_status AS ENUM ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED', 'ONHOLD');
CREATE TYPE ticket_priority AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'URGENT');

CREATE TABLE tickets (
    id SERIAL PRIMARY KEY,
    ticket_code VARCHAR(100) NOT NULL UNIQUE,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    location_id INTEGER NOT NULL REFERENCES locations(id),
    part_id INTEGER NOT NULL REFERENCES parts(id),
    asset_id INTEGER NOT NULL REFERENCES asset_ids(id),
    reporter_id INTEGER NOT NULL REFERENCES users(id),
    assigned_to_id INTEGER NOT NULL REFERENCES users(id),
    status ticket_status NOT NULL DEFAULT 'OPEN',
    priority ticket_priority NOT NULL DEFAULT 'HIGH',
    description TEXT NOT NULL,
    attachment TEXT NULL,
    due_at TIMESTAMP WITH TIME ZONE NOT NULL,
    roselved_at TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- +migrate Down
DROP TABLE tickets;
DROP TYPE ticket_status;
DROP TYPE ticket_priority;