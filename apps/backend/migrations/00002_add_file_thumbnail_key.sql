-- +goose Up
ALTER TABLE files ADD COLUMN thumbnail_key TEXT;

-- +goose Down
ALTER TABLE files DROP COLUMN thumbnail_key;
