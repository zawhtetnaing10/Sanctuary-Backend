-- name: CreateUsersHasInterests :one
INSERT INTO users_has_interests(user_id, interest_id, created_at, updated_at)
VALUES(
    $1,
    $2,
    NOW(),
    NOW()
)
RETURNING *;

-- name: DeleteAllUsersHasInterests :exec
DELETE FROM users_has_interests;