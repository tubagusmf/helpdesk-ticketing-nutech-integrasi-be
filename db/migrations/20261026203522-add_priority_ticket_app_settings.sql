
-- +migrate Up
ALTER TABLE app_settings
    ADD COLUMN IF NOT EXISTS priority_urgent_minutes
        INTEGER NOT NULL DEFAULT 15
        CHECK (
            priority_urgent_minutes BETWEEN 1 AND 525600
        ),
    ADD COLUMN IF NOT EXISTS priority_high_minutes
        INTEGER NOT NULL DEFAULT 60
        CHECK (
            priority_high_minutes BETWEEN 1 AND 525600
        ),
    ADD COLUMN IF NOT EXISTS priority_medium_minutes
        INTEGER NOT NULL DEFAULT 120
        CHECK (
            priority_medium_minutes BETWEEN 1 AND 525600
        ),
    ADD COLUMN IF NOT EXISTS priority_low_minutes
        INTEGER NOT NULL DEFAULT 240
        CHECK (
            priority_low_minutes BETWEEN 1 AND 525600
        );

-- +migrate Down
ALTER TABLE app_settings
    DROP COLUMN IF EXISTS priority_urgent_minutes,
    DROP COLUMN IF EXISTS priority_high_minutes,
    DROP COLUMN IF EXISTS priority_medium_minutes,
    DROP COLUMN IF EXISTS priority_low_minutes;
