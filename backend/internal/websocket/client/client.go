package client

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512 * 1024
)

// Hub interface to avoid circular dependencies
type Hub interface {
	Unregister(client *Client)
	Broadcast(message []byte)
}

// Client represents a WebSocket client connection
type Client struct {
	ID             string
	hub            Hub
	conn           *websocket.Conn
	send           chan []byte
	subscriptions  map[string]bool
	authenticated  bool
}

// NewClient creates a new WebSocket client
func NewClient(id string, hub Hub, conn *websocket.Conn) *Client {
	return &Client{
		ID:            id,
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
		authenticated: false,
	}
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket error: %v", err)
			}
			break
		}

		// Handle incoming messages
		c.handleMessage(message)
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("error unmarshaling message: %v", err)
		return
	}

	switch msg.Type {
	case "ping":
		c.sendMessage("pong", nil)

	case "subscribe":
		var payload struct {
			Environments []string `json:"environments"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			for _, envID := range payload.Environments {
				c.subscriptions[envID] = true
			}
		}

	case "unsubscribe":
		var payload struct {
			Environments []string `json:"environments"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			for _, envID := range payload.Environments {
				delete(c.subscriptions, envID)
			}
		}

	case "authenticate":
		// Handle authentication if needed
		c.authenticated = true

	default:
		log.Printf("unknown message type: %s", msg.Type)
	}
}

// sendMessage sends a message to the client
func (c *Client) sendMessage(msgType string, payload interface{}) {
	msg := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error marshaling message: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		// Client's send channel is full, close it
		close(c.send)
	}
}

// SendStatusUpdate sends environment status update to the client
func (c *Client) SendStatusUpdate(environmentID string, status interface{}) {
	// Check if client is subscribed to this environment
	if !c.subscriptions[environmentID] {
		return
	}

	c.sendMessage("status_update", map[string]interface{}{
		"environmentId": environmentID,
		"status":        status,
	})
}

// SendOperationUpdate sends operation update to the client
func (c *Client) SendOperationUpdate(operationID string, update interface{}) {
	c.sendMessage("operation_update", map[string]interface{}{
		"operationId": operationID,
		"update":      update,
	})
}

// Close closes the client connection
func (c *Client) Close() {
	close(c.send)
}
