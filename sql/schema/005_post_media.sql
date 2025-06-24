-- +goose Up
CREATE TABLE post_media(
    id BIGSERIAL PRIMARY KEY,
    media_url TEXT NOT NULL,
    order_index INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP,
    post_id BIGINT NOT NULL,

    CONSTRAINT fk_post_id
    FOREIGN KEY (post_id)
    REFERENCES posts(id)
    ON DELETE CASCADE
);

-- +goose Down
DROP TABLE post_media;