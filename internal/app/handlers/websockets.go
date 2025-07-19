package handlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zawhtetnaing10/Sanctuary-Backend/internal/database"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Client
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// App specific
	UserId         int64
	ConversationId int64
}

// Hub
type Hub struct {
	Db         *database.Queries
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// New Hub
func NewHub(Db *database.Queries) *Hub {
	return &Hub{
		Db:         Db,
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Runs the hub.
func (hub *Hub) Run() {
	log.Println("Websocket started. Hub is running.")

	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
			log.Printf("Client registered %v. Total clients %d", client, len(hub.clients))
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				close(client.send)
				log.Printf("Client unregistered. Total clients: %d", len(hub.clients))
			}
		case message := <-hub.broadcast:
			for client := range hub.clients {
				// Unmarshal the sent chat message
				var chatMessage ChatMessageResponse
				if unmarshalErr := json.Unmarshal(message, &chatMessage); unmarshalErr != nil {
					log.Printf("Error unmarshalling message")
					continue
				}

				// Only send the message when the conversation ids are the same.
				if chatMessage.ConversationId == client.ConversationId {

					select {
					case client.send <- message:
						// Send message successful
					default:
						close(client.send)
						delete(hub.clients, client)
						log.Printf("Client %s send channel blocked/closed. Unregistering.\n", client.conn.RemoteAddr())
					}
				}
			}
		}
	}
}

// Message object which will be sent from the client side.
type IncomingMessage struct {
	Content        string `json:"content"`
	SenderId       int    `json:"sender_id"`
	ConversationId int    `json:"conversation_id"`
}

// Client ReadPump. Reads messages from client and forwards them to the hub
func (c *Client) ReadPump() {
	// Clean up
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
		log.Printf("Client %s readPump exited\n", c.conn.RemoteAddr())
	}()

	// Set necessary limits
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Read message from connection
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for %s: %v\n", c.conn.RemoteAddr(), err)
			} else {
				log.Printf("Client %s disconnected or read error: %v\n", c.conn.RemoteAddr(), err)
			}
			break
		}

		// TODO: - Decode the message to create the request and save the message
		var incomingMessage IncomingMessage
		if err := json.Unmarshal(message, &incomingMessage); err != nil {
			log.Printf("Error unmarshaling incoming message for %s: %v, raw: %s\n", c.conn.RemoteAddr(), err, string(message))
			continue
		}

		// Validation for security risks
		if int64(incomingMessage.SenderId) != c.UserId {
			log.Printf("Security risk: Client %d tried to send message in conversation %d as user %d", c.UserId, c.ConversationId, incomingMessage.SenderId)
			continue
		}
		if int64(incomingMessage.ConversationId) != c.ConversationId {
			log.Printf("Security alert: Client %d in conv %d tried to send message to conv %d\n", c.UserId, c.ConversationId, incomingMessage.ConversationId)
			continue
		}

		// Save chat message in db.
		chatMsgParams := database.CreateChatMessageParams{
			Content:        incomingMessage.Content,
			ConversationID: int64(incomingMessage.ConversationId),
			SenderID:       int64(incomingMessage.SenderId),
		}
		createdChatMessage, chatMsgErr := c.hub.Db.CreateChatMessage(context.Background(), chatMsgParams)
		if chatMsgErr != nil {
			log.Printf("Error saving message to DB for %s: %v\n", c.conn.RemoteAddr(), chatMsgErr)
			continue
		}

		// Create the ChatMessageResponse and marshal it here.
		chatMessageResponse := ChatMessageResponse{
			Id:             createdChatMessage.ID,
			Content:        createdChatMessage.Content,
			ConversationId: createdChatMessage.ConversationID,
			SenderId:       createdChatMessage.SenderID,
			CreatedAt:      createdChatMessage.CreatedAt.Time,
			UpdatedAt:      createdChatMessage.UpdatedAt.Time,
		}

		messageBytes, err := json.Marshal(chatMessageResponse)
		if err != nil {
			log.Printf("Error marshalling message for %s: %v\n", c.conn.RemoteAddr(), createdChatMessage)
			continue
		}

		// Send it over to the hub
		c.hub.broadcast <- messageBytes
		log.Printf("Client %s received message: %s\n", c.conn.RemoteAddr(), string(message))
	}
}

// Client WritePump. Write messages back to client. Will come from send channel.
// Will send periodic ping signals to the client as well.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	// Clean Up
	defer func() {
		ticker.Stop()
		c.conn.Close()
		log.Printf("Client %s writePump exited.\n", c.conn.RemoteAddr())
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// Client is already closed. Write closed message
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Get writer
			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Client %s NextWriter error: %v\n", c.conn.RemoteAddr(), err)
				return
			}
			writer.Write(message)

			// If multiple messages are queued. Send as one.
			n := len(c.send)
			for range n {
				writer.Write([]byte("\n"))
				writer.Write(message)
			}

			// Clean up the writer
			if err := writer.Close(); err != nil {
				log.Printf("Client %s writer close error: %v\n", c.conn.RemoteAddr(), err)
				return
			}
		case <-ticker.C:
			// Write Ping message
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Client %s ping write error: %v\n", c.conn.RemoteAddr(), err)
				return
			}
		}
	}
}
