
-- +migrate Up
CREATE TABLE asset_id (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    part_id INTEGER NOT NULL REFERENCES parts(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

-- +migrate Down
DROP TABLE asset_id;