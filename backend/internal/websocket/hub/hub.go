package hub

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	clients      map[string]*Client
	broadcast    chan Message
	register     chan *Client
	unregister   chan *Client
	mu           sync.RWMutex
	logger       *logrus.Logger
}

// Client represents a WebSocket client connection
type Client struct {
	ID            string
	conn          *websocket.Conn
	send          chan Message
	subscriptions map[string]bool
	hub           *Hub
	logger        *logrus.Logger
}

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// SubscribeMessage represents a subscription request
type SubscribeMessage struct {
	Environments []string `json:"environments"`
}

// NewHub creates a new WebSocket hub
func NewHub(logger *logrus.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastToSubscribers(message)
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastEnvironmentUpdate broadcasts an environment status update
func (h *Hub) BroadcastEnvironmentUpdate(envID string, update map[string]interface{}) {
	message := Message{
		Type: "status_update",
		Payload: map[string]interface{}{
			"environmentId": envID,
			"status":        update,
		},
	}
	h.broadcast <- message
}

// BroadcastOperationUpdate broadcasts an operation update
func (h *Hub) BroadcastOperationUpdate(operationID string, update map[string]interface{}) {
	message := Message{
		Type: "operation_update",
		Payload: map[string]interface{}{
			"operationId": operationID,
			"update":      update,
		},
	}
	h.broadcast <- message
}

// registerClient handles client registration
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client.ID] = client
	h.logger.WithFields(logrus.Fields{
		"clientId": client.ID,
		"total":    len(h.clients),
	}).Info("Client registered")
}

// unregisterClient handles client unregistration
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.send)
		h.logger.WithFields(logrus.Fields{
			"clientId": client.ID,
			"total":    len(h.clients),
		}).Info("Client unregistered")
	}
}

// broadcastToSubscribers sends message to subscribed clients
func (h *Hub) broadcastToSubscribers(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Extract environment ID if present
	envID, ok := message.Payload["environmentId"].(string)
	if !ok && message.Type == "status_update" {
		return // Skip if no environment ID for status updates
	}

	for _, client := range h.clients {
		// For status updates, check if client is subscribed to the environment
		if message.Type == "status_update" && envID != "" {
			if !client.IsSubscribedTo(envID) {
				continue
			}
		}

		select {
		case client.send <- message:
		default:
			// Client's send channel is full, close it
			close(client.send)
			delete(h.clients, client.ID)
		}
	}
}

// NewClient creates a new client
func NewClient(id string, conn *websocket.Conn, hub *Hub, logger *logrus.Logger) *Client {
	return &Client{
		ID:            id,
		conn:          conn,
		send:          make(chan Message, 256),
		subscriptions: make(map[string]bool),
		hub:           hub,
		logger:        logger,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512 * 1024) // 512KB
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message Message
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.WithError(err).Error("WebSocket read error")
			}
			break
		}

		// Handle message based on type
		switch message.Type {
		case "subscribe":
			c.handleSubscribe(message.Payload)
		case "unsubscribe":
			c.handleUnsubscribe(message.Payload)
		case "ping":
			// Send pong response
			c.send <- Message{Type: "pong", Payload: map[string]interface{}{}}
		default:
			c.logger.WithField("type", message.Type).Warn("Unknown message type")
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				c.logger.WithError(err).Error("WebSocket write error")
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal subscribe payload")
		return
	}

	var sub SubscribeMessage
	if err := json.Unmarshal(data, &sub); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal subscribe message")
		return
	}

	for _, envID := range sub.Environments {
		c.subscriptions[envID] = true
	}

	c.logger.WithFields(logrus.Fields{
		"clientId":     c.ID,
		"environments": sub.Environments,
	}).Info("Client subscribed to environments")

	// Send confirmation
	c.send <- Message{
		Type: "subscribed",
		Payload: map[string]interface{}{
			"environments": sub.Environments,
		},
	}
}

// handleUnsubscribe handles unsubscription requests
func (c *Client) handleUnsubscribe(payload map[string]interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		c.logger.WithError(err).Error("Failed to marshal unsubscribe payload")
		return
	}

	var unsub SubscribeMessage
	if err := json.Unmarshal(data, &unsub); err != nil {
		c.logger.WithError(err).Error("Failed to unmarshal unsubscribe message")
		return
	}

	for _, envID := range unsub.Environments {
		delete(c.subscriptions, envID)
	}

	c.logger.WithFields(logrus.Fields{
		"clientId":     c.ID,
		"environments": unsub.Environments,
	}).Info("Client unsubscribed from environments")

	// Send confirmation
	c.send <- Message{
		Type: "unsubscribed",
		Payload: map[string]interface{}{
			"environments": unsub.Environments,
		},
	}
}

// IsSubscribedTo checks if the client is subscribed to an environment
func (c *Client) IsSubscribedTo(envID string) bool {
	return c.subscriptions[envID]
}
