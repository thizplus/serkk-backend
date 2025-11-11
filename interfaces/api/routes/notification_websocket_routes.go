package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

// SetupNotificationWebSocketRoutes sets up notification WebSocket routes
func SetupNotificationWebSocketRoutes(app *fiber.App, h *handlers.Handlers) {
	// Notification WebSocket endpoint with JWT authentication from query parameter
	app.Use("/ws/notifications", middleware.WebSocketProtected())
	app.Use("/ws/notifications", h.NotificationWSHandler.WebSocketUpgrade)
	app.Get("/ws/notifications", websocket.New(h.NotificationWSHandler.HandleNotificationWebSocket))
}
