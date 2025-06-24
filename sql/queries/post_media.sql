-- name: CreatePostMedia :one
INSERT INTO post_media(media_url, order_index, created_at, updated_at, post_id)
VALUES(
    $1,
    $2,
    NOW(),
    NOW(),
    $3
)
RETURNING *;