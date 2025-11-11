package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Level compress.Level // Compression level
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Level: compress.LevelDefault, // Best speed/compression ratio
	}
}

// NewCompression creates a compression middleware with default config
func NewCompression() fiber.Handler {
	return NewCompressionWithConfig(DefaultCompressionConfig())
}

// NewCompressionWithConfig creates compression middleware with custom config
func NewCompressionWithConfig(config CompressionConfig) fiber.Handler {
	return compress.New(compress.Config{
		Level: config.Level,
		// Only compress responses >= 1KB
		Next: func(c *fiber.Ctx) bool {
			// Skip compression for WebSocket upgrades
			if c.Get("Upgrade") == "websocket" {
				return true
			}
			// Skip compression for Server-Sent Events
			if c.Get("Accept") == "text/event-stream" {
				return true
			}
			return false
		},
	})
}

// NewBestSpeedCompression creates compression with focus on speed
func NewBestSpeedCompression() fiber.Handler {
	return compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
		Next: func(c *fiber.Ctx) bool {
			return c.Get("Upgrade") == "websocket" || c.Get("Accept") == "text/event-stream"
		},
	})
}

// NewBestCompressionCompression creates compression with focus on size
func NewBestCompressionCompression() fiber.Handler {
	return compress.New(compress.Config{
		Level: compress.LevelBestCompression,
		Next: func(c *fiber.Ctx) bool {
			return c.Get("Upgrade") == "websocket" || c.Get("Accept") == "text/event-stream"
		},
	})
}
