-- name: CreateComment :one
INSERT INTO comments(content, user_id, post_id, created_at, updated_at)
VALUES(
    $1,
    $2,
    $3,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetCommentsForPost :many
SELECT
    c.id,
    c.content,
    c.created_at,
    c.updated_at,
    c.user_id,
    c.post_id,
    u.id AS author_id, 
    u.email AS author_email, 
    u.user_name AS author_user_name,
    u.full_name AS author_full_name,
    u.profile_image_url AS author_profile_image_url,
    u.dob AS author_dob,
    u.created_at AS author_created_at, 
    u.updated_at AS author_updated_at
FROM comments c 
INNER JOIN users u ON c.user_id = u.id
WHERE c.post_id = $1
ORDER BY c.created_at ASC;
    


