-- +goose Up
-- +goose StatementBegin
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    author TEXT NOT NULL,
    email TEXT,
    content TEXT NOT NULL
);

CREATE INDEX idx_messages_deleted_at ON messages(deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE messages;
-- +goose StatementEnd
