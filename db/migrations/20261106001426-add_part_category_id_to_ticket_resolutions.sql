
-- +migrate Up
ALTER TABLE ticket_resolutions
ADD COLUMN part_category_id INTEGER NULL
REFERENCES part_categories(id);

CREATE INDEX idx_ticket_resolutions_part_category_id
ON ticket_resolutions(part_category_id);

-- +migrate Down
DROP INDEX IF EXISTS idx_ticket_resolutions_part_category_id;

ALTER TABLE ticket_resolutions
DROP COLUMN IF EXISTS part_category_id;