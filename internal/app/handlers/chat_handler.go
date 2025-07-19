package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Web socket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(request *http.Request) bool {
		return true
	},
}

// Handler for web socket
func (cfg *ApiConfig) WsHandler(writer http.ResponseWriter, request *http.Request) {

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

	// Get the conversation id from request
	var conversationId int
	conversationIdStr := request.URL.Query().Get("conversation_id")
	if conversationIdStr == "" {
		// if no conversation id is provided, return an error
		RespondWithError(writer, http.StatusBadRequest, "conversation id be provided")
		return
	} else {
		conversationIdConverted, err := strconv.Atoi(conversationIdStr)
		if err != nil || conversationIdConverted < 1 {
			RespondWithError(writer, http.StatusBadRequest, "Error converting conversation id")
			return
		}
		conversationId = conversationIdConverted
	}

	// Upgrade the request to web socket.
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		RespondWithError(writer, http.StatusInternalServerError, "cannot establish connection.")
		return
	}

	// Create a client
	client := &Client{
		hub:            cfg.Hub,
		conn:           conn,
		send:           make(chan []byte, 256),
		UserId:         userId,
		ConversationId: int64(conversationId),
	}
	// Register the client to the hub
	client.hub.register <- client
	// Run the Read and Write go-routines
	go client.ReadPump()
	go client.WritePump()

	// Log the fact that a client has connected.
	cfg.Logger.Info("New client connected", zap.String("remote_addr", conn.RemoteAddr().String()))
}
