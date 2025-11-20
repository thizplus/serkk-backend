package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

func SetupSimpleAutoPostRoutes(api fiber.Router, h *handlers.Handlers) {
	simpleAutoPost := api.Group("/simple-auto-post")

	// Setup (require authentication)
	simpleAutoPost.Post("/setup", middleware.Protected(), h.SimpleAutoPostHandler.SetupTable)

	// CSV Upload (require authentication)
	simpleAutoPost.Post("/upload", middleware.Protected(), h.SimpleAutoPostHandler.UploadCSV)

	// Queue Management (require authentication)
	simpleAutoPost.Get("/queue/status", middleware.Protected(), h.SimpleAutoPostHandler.GetQueueStatus)
	simpleAutoPost.Get("/queue", middleware.Protected(), h.SimpleAutoPostHandler.ListQueue)
}
