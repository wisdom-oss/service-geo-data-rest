-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS geodata.layers
ADD COLUMN IF NOT EXISTS "attribution_url" text;
-- +goose StatementEnd
