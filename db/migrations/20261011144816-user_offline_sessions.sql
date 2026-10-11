
-- +migrate Up
CREATE TABLE IF NOT EXISTS user_offline_sessions (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    justification TEXT NOT NULL,
    offline_at TIMESTAMPTZ NOT NULL,
    online_at TIMESTAMPTZ NULL,
    duration_seconds INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_offline_sessions_user_id
    ON user_offline_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_user_offline_sessions_offline_at
    ON user_offline_sessions(offline_at);

-- +migrate Down
DROP TABLE IF EXISTS user_offline_sessions;