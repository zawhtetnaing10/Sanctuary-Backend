package handlers

import (
	"net/http"

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
	// Upgrade the request to web socket.
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		RespondWithError(writer, http.StatusInternalServerError, "cannot establish connection.")
		return
	}

	// Create a client
	client := &Client{
		hub:  cfg.Hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	// Register the client to the hub
	client.hub.register <- client
	// Run the Read and Write go-routines
	go client.ReadPump()
	go client.WritePump()

	// Log the fact that a client has connected.
	cfg.Logger.Info("New client connected", zap.String("remote_addr", conn.RemoteAddr().String()))

}
