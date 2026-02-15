-- +goose Up
-- +goose StatementBegin
ALTER TABLE collection_items ADD COLUMN extraction_rules JSONB DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE collection_items DROP COLUMN IF EXISTS extraction_rules;
-- +goose StatementEnd
