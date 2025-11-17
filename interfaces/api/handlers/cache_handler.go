package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/infrastructure/redis"
)

type CacheHandler struct {
	feedCache *redis.FeedCacheService
}

func NewCacheHandler(feedCache *redis.FeedCacheService) *CacheHandler {
	return &CacheHandler{
		feedCache: feedCache,
	}
}

// GetCacheStats returns cache statistics including hit rate
// GET /api/v1/cache/stats
func (h *CacheHandler) GetCacheStats(c *fiber.Ctx) error {
	ctx := c.Context()

	stats, err := h.feedCache.GetCacheStats(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get cache stats",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// ResetCacheStats resets the cache statistics counters
// POST /api/v1/cache/stats/reset
func (h *CacheHandler) ResetCacheStats(c *fiber.Ctx) error {
	ctx := c.Context()

	err := h.feedCache.ResetCacheStats(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to reset cache stats",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Cache stats reset successfully",
	})
}

// InvalidateAllCaches clears all feed caches
// POST /api/v1/cache/invalidate
func (h *CacheHandler) InvalidateAllCaches(c *fiber.Ctx) error {
	ctx := c.Context()

	err := h.feedCache.InvalidateAllFeeds(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to invalidate caches",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "All feed caches invalidated successfully",
	})
}
