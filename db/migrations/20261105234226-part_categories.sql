
-- +migrate Up
CREATE TABLE part_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    part_id INTEGER NOT NULL REFERENCES parts(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_part_categories_part_id
ON part_categories(part_id);

-- +migrate Down
DROP TABLE part_categories;