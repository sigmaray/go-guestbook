-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages ADD COLUMN accept_language TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN IF EXISTS accept_language;
-- +goose StatementEnd
