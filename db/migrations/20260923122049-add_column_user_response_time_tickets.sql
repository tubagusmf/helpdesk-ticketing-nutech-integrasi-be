-- +migrate Up
ALTER TABLE tickets
ADD COLUMN staff_assigned_to_id INTEGER NULL
    REFERENCES users(id);

ALTER TABLE tickets
ADD COLUMN staff_assigned_at TIMESTAMP WITH TIME ZONE NULL;

ALTER TABLE tickets
ADD COLUMN staff_first_response_at TIMESTAMP WITH TIME ZONE NULL;

ALTER TABLE tickets
ADD COLUMN staff_response_time_seconds BIGINT NULL;


-- +migrate Down
ALTER TABLE tickets
DROP COLUMN staff_response_time_seconds;

ALTER TABLE tickets
DROP COLUMN staff_first_response_at;

ALTER TABLE tickets
DROP COLUMN staff_assigned_at;

ALTER TABLE tickets
DROP COLUMN staff_assigned_to_id;