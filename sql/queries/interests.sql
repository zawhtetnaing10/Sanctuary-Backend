-- name: GetAllInterests :many
SELECT * FROM interests;

-- name: GetExistingInterestIds :many
SELECT id FROM interests
WHERE id = ANY($1::bigint[]);

-- name: GetInterestsForUser :many
SELECT interests.* FROM interests
INNER JOIN users_has_interests 
ON interests.id = users_has_interests.interest_id
WHERE users_has_interests.user_id = $1;