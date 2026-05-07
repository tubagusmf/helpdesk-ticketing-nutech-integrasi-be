
-- +migrate Up
ALTER TABLE projects 
ALTER COLUMN code_prefix SET NOT NULL;

ALTER TABLE projects 
ADD CONSTRAINT unique_code_prefix UNIQUE (code_prefix);

-- +migrate Down
DROP CONSTRAINT unique_code_prefix;