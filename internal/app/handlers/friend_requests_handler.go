package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/app"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

type FRRequest struct {
	SenderId   int `json:"sender_id"`
	ReceiverId int `json:"receiver_id"`
}

type RequestWithFRId struct {
	FriendRequestId int `json:"friend_request_id"`
}

type FriendRequestResponseWithoutUser struct {
	Id            int64     `json:"id"`
	SenderId      int64     `json:"sender_id"`
	ReceiverId    int64     `json:"receiver_id"`
	RequestStatus string    `json:"request_status"`
	RequestedAt   time.Time `json:"requested_at"`
	AcceptedAt    time.Time `json:"accepted_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type FriendRequestResponseWithUser struct {
	Id            int64                    `json:"id"`
	SenderId      int64                    `json:"sender_id"`
	ReceiverId    int64                    `json:"receiver_id"`
	RequestStatus string                   `json:"request_status"`
	RequestedAt   time.Time                `json:"requested_at"`
	AcceptedAt    time.Time                `json:"accepted_at"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	User          userWithoutTokenResponse `json:"user"`
}

// Create Friend Request
func (cfg *ApiConfig) CreateFriendRequestHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_FRIEND_REQUEST)
		return
	}

	// Verify the bearer token and get the id
	_, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_FRIEND_REQUEST)
		return
	}

	// Decode response
	decoder := json.NewDecoder(request.Body)
	requestParams := FRRequest{}
	if err := decoder.Decode(&requestParams); err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusBadRequest, CLIENT_CANNOT_CREATE_FRIEND_REQUEST)
		return
	}

	senderId := requestParams.SenderId
	receiverId := requestParams.ReceiverId

	if senderId == 0 {
		cfg.LogError(SERVER_EMPTY_SENDER_ID, errors.New(SERVER_EMPTY_SENDER_ID))
		RespondWithError(writer, http.StatusBadRequest, CLIENT_SENDER_ID_CANNOT_BE_EMPTY)
		return
	}

	if receiverId == 0 {
		cfg.LogError(SERVER_EMPTY_RECEIVER_ID, errors.New(SERVER_EMPTY_RECEIVER_ID))
		RespondWithError(writer, http.StatusBadRequest, CLIENT_RECEIVER_ID_CANNOT_BE_EMPTY)
		return
	}

	params := database.CreateFriendRequestParams{
		SenderID:      int64(senderId),
		ReceiverID:    int64(receiverId),
		RequestStatus: app.PENDING,
	}

	createdFriendRequest, err := cfg.Db.CreateFriendRequest(request.Context(), params)
	if err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_CANNOT_CREATE_FRIEND_REQUEST)
		return
	}

	response := FriendRequestResponseWithoutUser{
		Id:            createdFriendRequest.ID,
		SenderId:      createdFriendRequest.SenderID,
		ReceiverId:    createdFriendRequest.ReceiverID,
		RequestStatus: createdFriendRequest.RequestStatus,
		RequestedAt:   createdFriendRequest.RequestedAt.Time,
		AcceptedAt:    createdFriendRequest.AcceptedAt.Time,
		CreatedAt:     createdFriendRequest.CreatedAt.Time,
		UpdatedAt:     createdFriendRequest.UpdatedAt.Time,
	}

	RespondWithJson(writer, http.StatusCreated, response)
}

// Get All Friend Requests For User
func (cfg *ApiConfig) GetAllFriendRequestsForUserHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_GET_FRIEND_REQUEST)
		return
	}

	// Verify the bearer token and get the id
	receiverId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_GET_FRIEND_REQUEST)
		return
	}

	friendRequests, err := cfg.Db.GetFriendRequestsWithSenderDetails(request.Context(), receiverId)
	if err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_CANNOT_GET_FRIEND_REQUEST)
		return
	}

	response := []FriendRequestResponseWithUser{}
	for _, fr := range friendRequests {
		frResponse := FriendRequestResponseWithUser{
			Id:            fr.RequestID,
			SenderId:      fr.SenderID,
			ReceiverId:    fr.ReceiverID,
			RequestStatus: fr.RequestStatus,
			RequestedAt:   fr.RequestedAt.Time,
			AcceptedAt:    fr.AcceptedAt.Time,
			CreatedAt:     fr.RequestCreatedAt.Time,
			UpdatedAt:     fr.RequestUpdatedAt.Time,
			User: userWithoutTokenResponse{
				ID:              fr.SenderUserID,
				Email:           fr.SenderUserEmail,
				UserName:        fr.SenderUserName,
				FullName:        fr.SenderUserFullName,
				ProfileImageUrl: fr.SenderUserProfileImageUrl.String,
				Dob:             FormatNullDobString(fr.SenderUserDob.Time),
				CreatedAt:       fr.SenderUserCreatedAt.Time,
				UpdatedAt:       fr.SenderUserUpdatedAt.Time,
			},
		}
		response = append(response, frResponse)
	}

	RespondWithJson(writer, http.StatusOK, response)
}

// Accept Friend Request
func (cfg *ApiConfig) AcceptFriendRequestHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_ACCEPT_FRIEND_REQUESTS)
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_ACCEPT_FRIEND_REQUESTS)
		return
	}

	// Decode response
	decoder := json.NewDecoder(request.Body)
	requestParams := RequestWithFRId{}
	if err := decoder.Decode(&requestParams); err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusBadRequest, CLIENT_CANNOT_CREATE_FRIEND_REQUEST)
		return
	}

	frId := requestParams.FriendRequestId

	params := database.AcceptFriendRequestParams{
		ID:         int64(frId),
		ReceiverID: userId,
	}

	_, err := cfg.Db.AcceptFriendRequest(request.Context(), params)
	if err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_CANNOT_ACCEPT_FRIEND_REQUEST)
		return
	}

	writer.WriteHeader(http.StatusOK)
}

// Get All Friends
func (cfg *ApiConfig) GetAllFriendsHandler(writer http.ResponseWriter, request *http.Request) {
	// Get the bearer token from the request
	token, tokenErr := GetBearerToken(request.Header)
	if tokenErr != nil {
		cfg.LogError(SERVER_MSG_ERROR_GET_BEARER_TOKEN, tokenErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_ACCEPT_FRIEND_REQUESTS)
		return
	}

	// Verify the bearer token and get the id
	userId, jwtErr := ValidateJWT(token, cfg.TokenSecret)
	if jwtErr != nil {
		cfg.LogError(SERVER_MSG_JWT_VALIDATION_FAILED, jwtErr)
		RespondWithError(writer, http.StatusUnauthorized, CLIENT_NOT_AUTHORIZED_ACCEPT_FRIEND_REQUESTS)
		return
	}

	friends, err := cfg.Db.GetFriends(request.Context(), userId)
	if err != nil {
		cfg.LogError(err.Error(), err)
		RespondWithError(writer, http.StatusInternalServerError, CLIENT_CANNOT_GET_FRIENDS)
		return
	}

	response := []FriendRequestResponseWithUser{}

	for _, fr := range friends {
		frResponse := FriendRequestResponseWithUser{
			Id:            fr.RequestID,
			SenderId:      fr.SenderID,
			ReceiverId:    fr.ReceiverID,
			RequestStatus: fr.RequestStatus,
			RequestedAt:   fr.RequestedAt.Time,
			AcceptedAt:    fr.AcceptedAt.Time,
			CreatedAt:     fr.RequestCreatedAt.Time,
			UpdatedAt:     fr.RequestUpdatedAt.Time,
			User: userWithoutTokenResponse{
				ID:              fr.FriendUserID,
				Email:           fr.FriendUserEmail,
				UserName:        fr.FriendUserName,
				FullName:        fr.FriendUserFullName,
				ProfileImageUrl: fr.FriendUserProfileImageUrl.String,
				Dob:             FormatNullDobString(fr.FriendUserDob.Time),
				CreatedAt:       fr.FriendUserCreatedAt.Time,
				UpdatedAt:       fr.FriendUserUpdatedAt.Time,
			},
		}

		response = append(response, frResponse)
	}

	RespondWithJson(writer, http.StatusOK, response)
}
