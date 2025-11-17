package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
)

func SetupCacheRoutes(api fiber.Router, h *handlers.Handlers) {
	// Skip if CacheHandler is not initialized
	if h.CacheHandler == nil {
		return
	}

	cache := api.Group("/cache")

	// Public endpoint - anyone can view cache stats
	cache.Get("/stats", h.CacheHandler.GetCacheStats)

	// Admin-only endpoints (add auth middleware if needed)
	// TODO: Add admin authentication middleware
	cache.Post("/stats/reset", h.CacheHandler.ResetCacheStats)
	cache.Post("/invalidate", h.CacheHandler.InvalidateAllCaches)
}
