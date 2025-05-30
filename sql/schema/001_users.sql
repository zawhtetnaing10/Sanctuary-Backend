-- +goose Up
CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    user_name TEXT NOT NULL,
    full_name TEXT NOT NULL,
    profile_image_url TEXT,
    dob DATE,
    hashed_password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP
);

-- +goose Down
DROP TABLE users;