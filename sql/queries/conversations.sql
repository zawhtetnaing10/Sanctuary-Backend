-- name: CreateConversation :one
INSERT INTO conversations(conversation_type, conversation_name, created_at, updated_at)
VALUES(
    $1,
    $2,
    NOW() AT TIME ZONE 'UTC',
    NOW() AT TIME ZONE 'UTC'
)
RETURNING *;


-- name: GetConversationsForUser :many
WITH UserConversations AS (
    SELECT cp.conversation_id
    FROM conversation_participants cp
    WHERE cp.user_id = $1
),
SingleOtherParticipantConversationIds AS (
    SELECT
        cp_current_user.conversation_id,
        cp_other_user.user_id AS other_user_id
    FROM conversation_participants cp_current_user
    JOIN conversation_participants cp_other_user
        ON cp_current_user.conversation_id = cp_other_user.conversation_id
    WHERE
        cp_current_user.user_id = $1
        AND cp_other_user.user_id != $1 
    GROUP BY cp_current_user.conversation_id, cp_other_user.user_id
    HAVING COUNT(cp_current_user.user_id) = 1 
        AND (SELECT COUNT(DISTINCT user_id) FROM conversation_participants WHERE conversation_id = cp_current_user.conversation_id) = 2
),
OtherParticipantDetails AS (
    SELECT
        sop.conversation_id,
        u.id AS other_user_id,
        u.email AS other_user_email,
        u.user_name AS other_user_user_name,
        u.full_name AS other_user_full_name,
        u.profile_image_url AS other_user_profile_image_url,
        u.dob AS other_user_dob,
        u.created_at AS other_user_created_at,
        u.updated_at AS other_user_updated_at

    FROM SingleOtherParticipantConversationIds sop
    JOIN users u ON sop.other_user_id = u.id
),
RankedMessages AS (
    SELECT
        cm.id,
        cm.conversation_id,
        cm.sender_id,
        cm.content,
        cm.created_at,
        cm.updated_at,
        ROW_NUMBER() OVER (PARTITION BY cm.conversation_id ORDER BY cm.created_at DESC, cm.id DESC) as rn
    FROM chat_messages cm
    INNER JOIN UserConversations uc ON cm.conversation_id = uc.conversation_id
)
SELECT
    c.id AS conversation_id,
    c.conversation_type AS conversation_type,
    COALESCE(c.conversation_name, opd.other_user_full_name) AS display_name,
    c.created_at AS conversation_created_at,
    c.updated_at AS conversation_updated_at,
    rm.id AS last_message_id,
    rm.content AS last_message_content,
    rm.conversation_id AS last_message_conversation_id,
    rm.sender_id AS last_message_sender_id,
    rm.created_at AS last_message_created_at,
    rm.updated_at AS last_message_updated_at,
    opd.other_user_id,
    opd.other_user_email,
    opd.other_user_user_name,
    opd.other_user_full_name,
    opd.other_user_profile_image_url,
    opd.other_user_dob,
    opd.other_user_created_at,
    opd.other_user_updated_at
FROM conversations c
INNER JOIN UserConversations uc ON c.id = uc.conversation_id 
INNER JOIN RankedMessages rm ON c.id = rm.conversation_id
LEFT JOIN OtherParticipantDetails opd ON c.id = opd.conversation_id
WHERE rm.rn = 1;

-- name: GetConversationForParticipants :one
SELECT conversation_id
FROM conversation_participants 
WHERE user_id IN ($1,$2)
GROUP BY conversation_id
HAVING COUNT(DISTINCT user_id) = 2;
