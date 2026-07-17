-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages ALTER COLUMN author DROP NOT NULL;

ALTER TABLE messages ADD COLUMN ip_address TEXT;
ALTER TABLE messages ADD COLUMN user_agent TEXT;
ALTER TABLE messages ADD COLUMN referer TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN IF EXISTS referer;
ALTER TABLE messages DROP COLUMN IF EXISTS user_agent;
ALTER TABLE messages DROP COLUMN IF EXISTS ip_address;

UPDATE messages SET author = 'Anonymous' WHERE author IS NULL OR author = '';
ALTER TABLE messages ALTER COLUMN author SET NOT NULL;
-- +goose StatementEnd
