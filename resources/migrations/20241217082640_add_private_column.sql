-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS geodata.layers
ADD COLUMN IF NOT EXISTS "private" boolean;

ALTER TABLE IF EXISTS geodata.layers
ALTER COLUMN "private"
SET DEFAULT false;

UPDATE geodata.layers
SET
    private = false;

ALTER TABLE IF EXISTS geodata.layers
ALTER COLUMN "private"
SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS geodata.layers
DROP COLUMN "private"
-- +goose StatementEnd
