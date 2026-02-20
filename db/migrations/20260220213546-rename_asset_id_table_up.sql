
-- +migrate Up
ALTER TABLE asset_id RENAME TO asset_ids;

-- +migrate Down
ALTER TABLE asset_ids RENAME TO asset_id;