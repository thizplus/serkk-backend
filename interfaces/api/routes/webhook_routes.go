package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
)

func SetupWebhookRoutes(api fiber.Router, h *handlers.Handlers) {
	webhooks := api.Group("/webhooks")

	// Bunny Stream webhook (no authentication - Bunny doesn't support auth headers well)
	// Security: Validate source IP or use signed webhooks in production
	webhooks.Post("/bunny/video-status", h.WebhookHandler.BunnyStreamWebhook)
}
