package websocket

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	ws "gofiber-template/infrastructure/websocket"
)

type NotificationWebSocketHandler struct {
	notificationHub *ws.NotificationHub
}

func NewNotificationWebSocketHandler(notificationHub *ws.NotificationHub) *NotificationWebSocketHandler {
	return &NotificationWebSocketHandler{
		notificationHub: notificationHub,
	}
}

// WebSocketUpgrade checks if the request is a WebSocket upgrade
func (h *NotificationWebSocketHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// HandleNotificationWebSocket handles notification WebSocket connections
func (h *NotificationWebSocketHandler) HandleNotificationWebSocket(c *websocket.Conn) {
	var userID uuid.UUID

	// Get user from locals (set by Protected middleware)
	if userContext := c.Locals("userID"); userContext != nil {
		if id, ok := userContext.(uuid.UUID); ok {
			userID = id
		}
	}

	// If no user from middleware, deny connection
	if userID == uuid.Nil {
		log.Printf("❌ Notification WebSocket: Unauthorized connection attempt")
		c.WriteJSON(map[string]interface{}{
			"type":  "error",
			"error": map[string]string{
				"code":    "unauthorized",
				"message": "Authentication required",
			},
		})
		c.Close()
		return
	}

	log.Printf("✅ Notification WebSocket: User connected: %s", userID)

	// Create notification client
	client := &ws.NotificationClient{
		UserID: userID,
		Conn:   c,
		Send:   make(chan []byte, 256),
		Hub:    h.notificationHub,
	}

	// Register client
	h.notificationHub.RegisterClient(client)

	// Start read and write pumps
	go client.WritePump()
	client.ReadPump() // Blocking call

	// When ReadPump returns, the connection has been closed
	log.Printf("📴 Notification WebSocket: Connection closed for user: %s", userID)
}
