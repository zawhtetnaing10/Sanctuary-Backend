-- name: CreateFriendRequest :one
INSERT INTO friend_requests(sender_id, receiver_id, request_status, requested_at, created_at, updated_at)
VALUES(
    $1,
    $2,
    $3,
    NOW() AT TIME ZONE 'UTC',
    NOW() AT TIME ZONE 'UTC',
    NOW() AT TIME ZONE 'UTC'
)
RETURNING *;

-- name: GetPendingFriendRequestsBetweenTwoUsers :many
SELECT *
FROM friend_requests
WHERE sender_id = $1 AND receiver_id = $2 AND request_status = 'pending';

-- name: AcceptFriendRequest :one
UPDATE friend_requests
SET
    request_status = 'accepted',
    accepted_at = NOW() AT TIME ZONE 'UTC',
    updated_at = NOW() AT TIME ZONE 'UTC'
WHERE
    id = $1 AND receiver_id = $2 AND request_status = 'pending'
RETURNING *;

-- name: GetFriendRequestsWithSenderDetails :many
SELECT
    fr.id as request_id,
    fr.sender_id, 
    fr.receiver_id, 
    fr.request_status,
    fr.requested_at,
    fr.accepted_at,
    fr.created_at as request_created_at,
    fr.updated_at as request_updated_at,
    u.id AS sender_user_id, 
    u.email AS sender_user_email, 
    u.user_name AS sender_user_name,
    u.full_name AS sender_user_full_name,
    u.profile_image_url AS sender_user_profile_image_url,
    u.dob AS sender_user_dob,
    u.created_at AS sender_user_created_at, 
    u.updated_at AS sender_user_updated_at
FROM friend_requests fr
INNER JOIN users u
ON fr.sender_id = u.id
WHERE fr.receiver_id = $1 AND fr.request_status = 'pending';

-- name: GetFriends :many
SELECT
    fr.id as request_id,
    fr.sender_id, 
    fr.receiver_id, 
    fr.request_status,
    fr.requested_at,
    fr.accepted_at,
    fr.created_at as request_created_at,
    fr.updated_at as request_updated_at,
    u.id AS friend_user_id, 
    u.email AS friend_user_email, 
    u.user_name AS friend_user_name,
    u.full_name AS friend_user_full_name,
    u.profile_image_url AS friend_user_profile_image_url,
    u.dob AS friend_user_dob,
    u.created_at AS friend_user_created_at, 
    u.updated_at AS friend_user_updated_at
FROM friend_requests fr
JOIN users u
ON (fr.receiver_id = $1 AND u.id = fr.sender_id) OR (fr.sender_id = $1 AND u.id = fr.receiver_id)
WHERE
    fr.request_status = 'accepted'
    AND (fr.sender_id = $1 OR fr.receiver_id = $1)
ORDER BY
    fr.accepted_at DESC;

