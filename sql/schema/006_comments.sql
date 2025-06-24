-- +goose Up
CREATE TABLE comments(
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    user_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    
    CONSTRAINT fk_user_id
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE,

    CONSTRAINT fk_post_id
    FOREIGN KEY (post_id)
    REFERENCES posts(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE comments;