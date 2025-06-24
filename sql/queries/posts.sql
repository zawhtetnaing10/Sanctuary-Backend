-- name: CreatePost :one
INSERT INTO posts(content, created_at, updated_at, user_id)
VALUES(
    $1,
    NOW(),
    NOW(),
    $2
)
RETURNING *;


