-- name: GetChatMessagesForConversation :many
SELECT id, content, conversation_id, sender_id, created_at, updated_at
FROM chat_messages
WHERE conversation_id = $1;