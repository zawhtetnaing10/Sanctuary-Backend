-- name: CreateUsersHasInterests :copyfrom
INSERT INTO users_has_interests(user_id, interest_id, created_at, updated_at)
VALUES($1, $2, $3, $4);

-- name: GetDuplicateInterestIds :many
SELECT interest_id 
FROM users_has_interests
WHERE user_id = $1 AND interest_id = ANY($2::bigint[]);

-- name: DeleteAllUsersHasInterests :exec
DELETE FROM users_has_interests;