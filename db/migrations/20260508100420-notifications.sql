
-- +migrate Up
CREATE TYPE notification_type AS ENUM (
    'TICKET_CREATED',
    'TICKET_ASSIGNED',
    'TICKET_UPDATED',
    'TICKET_COMMENT',
    'TICKET_RESOLVED',
    'TICKET_CLOSED'
);

CREATE TYPE notification_reference_type AS ENUM (
    'TICKET',
    'COMMENT',
    'RESOLUTION'
);

CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    ticket_id INTEGER NOT NULL REFERENCES tickets(id),
    actor_id INTEGER NOT NULL REFERENCES users(id),
    type notification_type NOT NULL,
    reference_type notification_reference_type NOT NULL,
    reference_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_user_id
ON notifications(user_id);

CREATE INDEX idx_notifications_ticket_id
ON notifications(ticket_id);

CREATE INDEX idx_notifications_is_read
ON notifications(is_read);

CREATE INDEX idx_notifications_created_at
ON notifications(created_at DESC);

-- +migrate Down
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_notifications_ticket_id;
DROP INDEX IF EXISTS idx_notifications_is_read;
DROP INDEX IF EXISTS idx_notifications_created_at;

DROP TABLE notifications;

DROP TYPE notification_type;
DROP TYPE notification_reference_type;