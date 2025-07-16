package handlers

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
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
	UserId         int
	Username       string
	ConversationId int
}

// Hub
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// New Hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Runs the hub.
func (hub *Hub) Run() {
	log.Println("Websocket started.")

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

		c.hub.broadcast <- message
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
