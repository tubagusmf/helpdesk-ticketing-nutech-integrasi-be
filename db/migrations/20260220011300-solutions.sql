
-- +migrate Up
CREATE TABLE solutions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    causes_id INTEGER NOT NULL REFERENCES causes(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- +migrate Down
DROP TABLE solutions;