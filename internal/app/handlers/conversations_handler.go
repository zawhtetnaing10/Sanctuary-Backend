package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

type ChatMessageHistoryResponse struct {
	ConversationId int64                    `json:"conversation_id"`
	User           userWithoutTokenResponse `json:"user"`
	ChatMessages   []ChatMessageResponse    `json:"chat_messages"`
}

type ConversationResponse struct {
	Id               int64                    `json:"id"`
	ConversationName string                   `json:"conversation_name"`
	ConversationType string                   `json:"conversation_type"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
	LastMessage      ChatMessageResponse      `json:"last_message"`
	User             userWithoutTokenResponse `json:"user"`
}

type ChatMessageResponse struct {
	Id             int64     `json:"id"`
	Content        string    `json:"content"`
	ConversationId int64     `json:"conversation_id"`
	SenderId       int64     `json:"sender_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Get all conversations for user
func (cfg *ApiConfig) GetAllConversationsHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized get comments.")
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get comments.")
		return
	}

	conversationsFromDb, err := cfg.Db.GetConversationsForUser(request.Context(), userId)
	if err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusInternalServerError, "Unable to get conversations. Please try again.")
		return
	}

	response := []ConversationResponse{}

	for _, conversationFromDb := range conversationsFromDb {
		conversationResponse := ConversationResponse{
			Id:               conversationFromDb.ConversationID,
			ConversationName: conversationFromDb.DisplayName,
			ConversationType: conversationFromDb.ConversationType,
			CreatedAt:        conversationFromDb.ConversationCreatedAt.Time,
			UpdatedAt:        conversationFromDb.ConversationUpdatedAt.Time,
			LastMessage: ChatMessageResponse{
				Id:             conversationFromDb.LastMessageID,
				Content:        conversationFromDb.LastMessageContent,
				ConversationId: conversationFromDb.LastMessageConversationID,
				SenderId:       conversationFromDb.LastMessageSenderID,
				CreatedAt:      conversationFromDb.LastMessageCreatedAt.Time,
				UpdatedAt:      conversationFromDb.LastMessageUpdatedAt.Time,
			},
			User: userWithoutTokenResponse{
				ID:              conversationFromDb.OtherUserID.Int64,
				Email:           conversationFromDb.OtherUserEmail.String,
				UserName:        conversationFromDb.OtherUserUserName.String,
				FullName:        conversationFromDb.OtherUserFullName.String,
				ProfileImageUrl: conversationFromDb.OtherUserProfileImageUrl.String,
				Dob:             FormatNullDobString(conversationFromDb.OtherUserDob.Time),
				CreatedAt:       conversationFromDb.OtherUserCreatedAt.Time,
				UpdatedAt:       conversationFromDb.OtherUserUpdatedAt.Time,
			},
		}

		response = append(response, conversationResponse)
	}

	RespondWithJson(writer, http.StatusOK, response)
}

// Get message history.
func (cfg *ApiConfig) GetChatMessageHistoryHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized get comments.")
		return
	}

	// Verify the bearer token and get the id
	loggedInUserId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, "You are not authorized to get comments.")
		return
	}

	var otherUserId int
	otherUserIdStr := request.URL.Query().Get("other_user_id")
	if otherUserIdStr == "" {
		RespondWithError(writer, http.StatusBadRequest, "Other user's id must be provided.")
		return
	} else {
		postIdConverted, err := strconv.Atoi(otherUserIdStr)
		if err != nil || postIdConverted < 1 {
			RespondWithError(writer, http.StatusBadRequest, "Error converting user id")
			return
		}
		otherUserId = postIdConverted
	}

	// Get logged in user
	loggedInUserFromDb, loggedInUserErr := cfg.Db.GetUserById(request.Context(), int64(loggedInUserId))
	if loggedInUserErr != nil {
		cfg.LogError(loggedInUserErr.Error(), loggedInUserErr)
		RespondWithError(writer, http.StatusInternalServerError, "Failed to fetch logged in user. Please try again.")
		return
	}

	// Get other user
	otherUserFromDb, otherUserErr := cfg.Db.GetUserById(request.Context(), int64(otherUserId))
	if otherUserErr != nil {
		cfg.LogError(otherUserErr.Error(), otherUserErr)
		RespondWithError(writer, http.StatusInternalServerError, "Failed to fetch user. Please try again.")
		return
	}
	otherUser := userWithoutTokenResponse{
		ID:              otherUserFromDb.ID,
		Email:           otherUserFromDb.Email,
		FullName:        otherUserFromDb.FullName,
		UserName:        otherUserFromDb.UserName,
		ProfileImageUrl: otherUserFromDb.ProfileImageUrl.String,
		Dob:             FormatNullDobString(otherUserFromDb.Dob.Time),
		CreatedAt:       otherUserFromDb.CreatedAt.Time,
		UpdatedAt:       otherUserFromDb.UpdatedAt.Time,
	}

	// Get Conversation id
	params := database.GetConversationForParticipantsParams{
		UserID:   loggedInUserId,
		UserID_2: int64(otherUserId),
	}

	conversationId, conversationIdErr := cfg.Db.GetConversationForParticipants(request.Context(), params)
	if conversationIdErr != nil {
		if errors.Is(conversationIdErr, sql.ErrNoRows) {
			err := cfg.WithTransaction(request.Context(), func(q *database.Queries) error {
				// Create new conversation
				newConversationParams := database.CreateConversationParams{
					ConversationName: pgtype.Text{
						String: fmt.Sprintf("%s - %s", loggedInUserFromDb.UserName, otherUser.UserName),
						Valid:  true,
					},
					ConversationType: "peer-to-peer",
				}

				createdConversation, createConversationErr := q.CreateConversation(request.Context(), newConversationParams)
				if createConversationErr != nil {
					return createConversationErr
				}
				// Reset the conversationId with the currently created one.
				conversationId = createdConversation.ID

				// Create conversation participant for logged in user.
				conversationParticipantParamsOne := database.CreateConversationParticipantParams{
					ConversationID: createdConversation.ID,
					UserID:         loggedInUserId,
				}
				_, conversationParamErrOne := q.CreateConversationParticipant(request.Context(), conversationParticipantParamsOne)
				if conversationParamErrOne != nil {
					return conversationParamErrOne
				}

				// Create conversation participant for other user
				conversationParticipantParamTwo := database.CreateConversationParticipantParams{
					ConversationID: createdConversation.ID,
					UserID:         otherUser.ID,
				}
				_, conversationParticipantErrTwo := q.CreateConversationParticipant(request.Context(), conversationParticipantParamTwo)
				if conversationParticipantErrTwo != nil {
					return conversationParticipantErrTwo
				}

				return nil
			})

			if err != nil {
				cfg.LogError(err.Error(), err)
				RespondWithError(writer, http.StatusInternalServerError, "Failed to create conversation. Please try agian.")
				return
			}
		} else {
			cfg.LogError(conversationIdErr.Error(), conversationIdErr)
			RespondWithError(writer, http.StatusInternalServerError, "Failed to find conversation.")
			return
		}
	}

	// Get Messages
	chatMessagesFromDb, chatMessagesErr := cfg.Db.GetChatMessagesForConversation(request.Context(), conversationId)
	if chatMessagesErr != nil {
		cfg.LogError(chatMessagesErr.Error(), chatMessagesErr)
		RespondWithError(writer, http.StatusInternalServerError, "Failed to fetch chat messages. Please try agin.")
		return
	}

	chatMessages := []ChatMessageResponse{}
	for _, chatMessageFromDb := range chatMessagesFromDb {
		chatMessageResponse := ChatMessageResponse{
			Id:             chatMessageFromDb.ID,
			Content:        chatMessageFromDb.Content,
			ConversationId: chatMessageFromDb.ConversationID,
			SenderId:       chatMessageFromDb.SenderID,
			CreatedAt:      chatMessageFromDb.CreatedAt.Time,
			UpdatedAt:      chatMessageFromDb.UpdatedAt.Time,
		}
		chatMessages = append(chatMessages, chatMessageResponse)
	}

	response := ChatMessageHistoryResponse{
		ConversationId: conversationId,
		User:           otherUser,
		ChatMessages:   chatMessages,
	}

	RespondWithJson(writer, http.StatusOK, response)
}

func (cfg *ApiConfig) WithTransaction(ctx context.Context, fn func(q *database.Queries) error) error {
	tx, err := cfg.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	qtx := cfg.Db.WithTx(tx)

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
			if err != nil {
				cfg.LogError("Failed to commit transaction", err)
			}
		}
	}()

	err = fn(qtx)

	return err
}
