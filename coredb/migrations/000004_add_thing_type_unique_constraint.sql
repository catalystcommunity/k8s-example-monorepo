-- +goose Up
-- Add unique constraint to thing_types.name
ALTER TABLE thing_types ADD CONSTRAINT thing_types_name_unique UNIQUE (name);

-- +goose Down
-- Remove unique constraint from thing_types.name
ALTER TABLE thing_types DROP CONSTRAINT thing_types_name_unique;