package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

func SetupChatWebSocketRoutes(app *fiber.App, h *handlers.Handlers) {
	// Chat WebSocket endpoint with JWT authentication from query parameter
	app.Use("/ws/chat", middleware.WebSocketProtected())
	app.Use("/ws/chat", h.ChatWSHandler.WebSocketUpgrade)
	app.Get("/ws/chat", websocket.New(h.ChatWSHandler.HandleChatWebSocket))
}
