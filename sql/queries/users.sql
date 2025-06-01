-- name: CreateUser :one
INSERT INTO users(email, user_name, full_name, profile_image_url, dob, hashed_password, created_at, updated_at)
VALUES(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    NOW(),
    NOW()
)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users
WHERE id = $1;

-- name: DeleteAllUsers: exec
DELETE FROM users;