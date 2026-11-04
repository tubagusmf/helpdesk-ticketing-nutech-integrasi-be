
-- +migrate Up
CREATE TABLE problems (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE UNIQUE INDEX idx_problems_name_active
ON problems (LOWER(name))
WHERE deleted_at IS NULL;

-- +migrate Down
DROP TABLE IF EXISTS problems;
