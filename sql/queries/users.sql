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

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: UpdateUserProfile :one
UPDATE users
SET
    full_name = $2,
    user_name = $3,
    profile_image_url = $4,
    dob = $5
WHERE
    id = $1
RETURNING *;