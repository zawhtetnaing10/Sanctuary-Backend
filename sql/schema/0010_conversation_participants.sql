-- +goose Up
CREATE TABLE conversation_participants(
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(conversation_id,user_id)
);

-- +goose Down
DROP TABLE conversation_participants;