
-- +migrate Up
ALTER TABLE app_settings
    ADD COLUMN IF NOT EXISTS login_session_minutes
        INTEGER NOT NULL DEFAULT 600
        CHECK (login_session_minutes BETWEEN 5 AND 43200),

    ADD COLUMN IF NOT EXISTS staff_response_sla_minutes
        INTEGER NOT NULL DEFAULT 3
        CHECK (staff_response_sla_minutes BETWEEN 1 AND 525600);

-- +migrate Down
ALTER TABLE app_settings
    DROP COLUMN IF EXISTS login_session_minutes,
    DROP COLUMN IF EXISTS staff_response_sla_minutes;