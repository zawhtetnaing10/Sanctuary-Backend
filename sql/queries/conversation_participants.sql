-- name: CreateConversationParticipant :one
INSERT INTO conversation_participants(conversation_id, user_id)
VALUES(
    $1,
    $2
)
RETURNING *;
