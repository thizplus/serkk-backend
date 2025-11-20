package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

func SetupAutoPostRoutes(api fiber.Router, h *handlers.Handlers) {
	autoPost := api.Group("/auto-post")

	// Settings routes (require authentication)
	settings := autoPost.Group("/settings", middleware.Protected())
	settings.Post("/", h.AutoPostHandler.CreateSetting)
	settings.Get("/", h.AutoPostHandler.ListSettings)
	settings.Get("/:id", h.AutoPostHandler.GetSetting)
	settings.Put("/:id", h.AutoPostHandler.UpdateSetting)
	settings.Delete("/:id", h.AutoPostHandler.DeleteSetting)

	// Enable/Disable
	settings.Post("/:id/enable", h.AutoPostHandler.EnableSetting)
	settings.Post("/:id/disable", h.AutoPostHandler.DisableSetting)

	// Trigger manual post generation
	settings.Post("/:id/trigger", h.AutoPostHandler.TriggerAutoPost)

	// Logs routes (require authentication)
	logs := autoPost.Group("/logs", middleware.Protected())
	logs.Get("/", h.AutoPostHandler.ListLogs)
	logs.Get("/:id", h.AutoPostHandler.GetLog)
}
