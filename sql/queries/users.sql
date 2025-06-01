-- name: CreateUser :one
INSERT INTO users(email, hashed_password, user_name, full_name, created_at, updated_at)
VALUES(
    $1,
    $2,
    $3,
    $4,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;