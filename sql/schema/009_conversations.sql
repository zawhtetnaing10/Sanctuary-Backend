-- +goose Up
CREATE TABLE conversations(
    id BIGSERIAL PRIMARY KEY,
    conversation_type TEXT NOT NULL,
    conversation_name TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- +goose Down
DROP TABLE conversations;
