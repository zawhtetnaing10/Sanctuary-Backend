-- name: CreatePost :one
INSERT INTO posts(content, created_at, updated_at, user_id)
VALUES(
    $1,
    NOW(),
    NOW(),
    $2
)
RETURNING *;

-- name: DeleteAllPosts :exec
DELETE FROM posts;

-- name: GetAllPosts :many
SELECT
    p.id,
    p.content,
    p.created_at,
    p.updated_at,
    p.user_id,
    u.id AS author_id, 
    u.email AS author_email, 
    u.user_name AS author_user_name,
    u.full_name AS author_full_name,
    u.profile_image_url AS author_profile_image_url,
    u.dob AS author_dob,
    u.created_at AS author_created_at, 
    u.updated_at AS author_updated_at, 
    (SELECT COUNT(*) FROM post_likes pl WHERE pl.post_id = p.id) AS like_count,
    (SELECT COUNT(*) FROM comments c WHERE c.post_id = p.id) AS comment_count,
    (
        SELECT EXISTS(
            SELECT 1 FROM post_likes upl WHERE upl.post_id = p.id AND upl.user_id = $1
        )
    ) AS liked_by_user,
    CAST(COALESCE(ARRAY_AGG(pm.media_url ORDER BY pm.id) FILTER (WHERE pm.media_url IS NOT NULL), '{}'::text[]) AS text[]) AS media_urls_array

FROM posts p
INNER JOIN users u ON p.user_id = u.id
LEFT JOIN post_media pm ON p.id = pm.post_id
WHERE p.deleted_at IS NULL 
GROUP BY
    p.id,               
    p.content,
    p.created_at,
    p.updated_at,
    p.user_id,
    u.id,               
    u.email,
    u.user_name,
    u.full_name,
    u.profile_image_url,
    u.dob,
    u.created_at,
    u.updated_at
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3; 


