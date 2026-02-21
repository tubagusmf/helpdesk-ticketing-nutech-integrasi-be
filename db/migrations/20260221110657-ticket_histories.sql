
-- +migrate Up
CREATE TABLE ticket_histories (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT NULL,
    new_value TEXT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- +migrate Down
DROP TABLE ticket_histories;