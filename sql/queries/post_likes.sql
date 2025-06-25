-- name: CreatePostLike :one
INSERT INTO post_likes(user_id, post_id, created_at, updated_at)
VALUES(
    $1,
    $2,
    NOW(),
    NOW()
)
RETURNING *;

-- name: DeletePostLike :exec
DELETE FROM post_likes WHERE user_id=$1 AND post_id=$2;

-- name: GetPostLike :one
SELECT * FROM post_likes WHERE user_id=$1 AND post_id=$2 LIMIT 1;