-- +goose Up
ALTER TABLE sessions
ADD COLUMN rotated boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE sessions
DROP COLUMN rotated;