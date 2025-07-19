-- name: GetChatMessagesForConversation :many
SELECT id, content, conversation_id, sender_id, created_at, updated_at
FROM chat_messages
WHERE conversation_id = $1;

-- name: CreateChatMessage :one
INSERT INTO chat_messages (content, conversation_id, sender_id, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    NOW() AT TIME ZONE 'UTC',
    NOW() AT TIME ZONE 'UTC'
)
RETURNING *;