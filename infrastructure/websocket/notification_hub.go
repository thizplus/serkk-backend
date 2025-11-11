package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// NotificationHub manages notification-specific WebSocket connections
type NotificationHub struct {
	// Client management
	clients      map[uuid.UUID]*NotificationClient
	clientsMutex sync.RWMutex

	// Channels
	register   chan *NotificationClient
	unregister chan *NotificationClient
	broadcast  chan *NotificationMessage

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NotificationClient represents a connected notification user
type NotificationClient struct {
	UserID uuid.UUID
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *NotificationHub
}

// NotificationMessage represents a notification WebSocket message
type NotificationMessage struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
	Error   *NotificationError     `json:"error,omitempty"`
}

// NotificationError represents an error message
type NotificationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewNotificationHub creates a new NotificationHub
func NewNotificationHub() *NotificationHub {
	ctx, cancel := context.WithCancel(context.Background())

	return &NotificationHub{
		clients:    make(map[uuid.UUID]*NotificationClient),
		register:   make(chan *NotificationClient, 10),
		unregister: make(chan *NotificationClient, 10),
		broadcast:  make(chan *NotificationMessage, 256),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Run starts the hub's main loop
func (h *NotificationHub) Run() {
	log.Println("🚀 NotificationHub started")

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-h.ctx.Done():
			log.Println("NotificationHub stopping...")
			return
		}
	}
}

// RegisterClient adds a new client to the hub
func (h *NotificationHub) RegisterClient(client *NotificationClient) {
	h.register <- client
}

// UnregisterClient removes a client from the hub
func (h *NotificationHub) UnregisterClient(client *NotificationClient) {
	h.unregister <- client
}

// SendToUser sends a message to a specific user
func (h *NotificationHub) SendToUser(userID uuid.UUID, message *NotificationMessage) {
	h.clientsMutex.RLock()
	client, exists := h.clients[userID]
	h.clientsMutex.RUnlock()

	if exists {
		h.sendToClient(client, message)
	} else {
		log.Printf("⚠️  Notification client not found for user: %s", userID)
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *NotificationHub) BroadcastToAll(message *NotificationMessage) {
	h.broadcast <- message
}

// registerClient handles client registration
func (h *NotificationHub) registerClient(client *NotificationClient) {
	h.clientsMutex.Lock()
	h.clients[client.UserID] = client
	h.clientsMutex.Unlock()

	log.Printf("✅ Notification client registered: UserID=%s, Total clients=%d", client.UserID, len(h.clients))

	// Send connection success message
	h.sendToClient(client, &NotificationMessage{
		Type: "connection.success",
		Payload: map[string]interface{}{
			"message":   "Connected to notification service",
			"userId":    client.UserID.String(),
			"timestamp": time.Now().Unix(),
		},
	})
}

// unregisterClient handles client disconnection
func (h *NotificationHub) unregisterClient(client *NotificationClient) {
	h.clientsMutex.Lock()
	if _, exists := h.clients[client.UserID]; exists {
		delete(h.clients, client.UserID)
		close(client.Send)
	}
	h.clientsMutex.Unlock()

	log.Printf("❌ Notification client unregistered: UserID=%s, Total clients=%d", client.UserID, len(h.clients))
}

// broadcastMessage sends a message to all connected clients
func (h *NotificationHub) broadcastMessage(message *NotificationMessage) {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()

	for _, client := range h.clients {
		h.sendToClient(client, message)
	}
}

// sendToClient sends a message to a specific client
func (h *NotificationHub) sendToClient(client *NotificationClient, message *NotificationMessage) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return
	}

	select {
	case client.Send <- messageJSON:
		// Message sent successfully
	default:
		// Client's send channel is full, skip
		log.Printf("⚠️  Client send buffer full, skipping notification for user: %s", client.UserID)
	}
}

// GetOnlineUsersCount returns the number of connected users
func (h *NotificationHub) GetOnlineUsersCount() int {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	return len(h.clients)
}

// IsUserOnline checks if a user is connected
func (h *NotificationHub) IsUserOnline(userID uuid.UUID) bool {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	_, exists := h.clients[userID]
	return exists
}

// Stop gracefully shuts down the hub
func (h *NotificationHub) Stop() {
	log.Println("Stopping NotificationHub...")
	h.cancel()

	// Close all client connections
	h.clientsMutex.Lock()
	for _, client := range h.clients {
		close(client.Send)
	}
	h.clientsMutex.Unlock()
}
