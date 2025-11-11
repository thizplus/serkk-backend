package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	Max        int           // Maximum number of requests
	Expiration time.Duration // Time window for rate limit
	Message    string        // Custom error message
}

// DefaultRateLimiterConfig returns default rate limiter settings
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Max:        100,             // 100 requests
		Expiration: 1 * time.Minute, // per minute
		Message:    "Too many requests, please try again later",
	}
}

// StrictRateLimiterConfig returns strict rate limiter for sensitive endpoints
func StrictRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Max:        5,               // 5 requests
		Expiration: 1 * time.Minute, // per minute
		Message:    "Too many attempts, please try again later",
	}
}

// AuthRateLimiterConfig returns rate limiter for auth endpoints
func AuthRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Max:        10,              // 10 requests
		Expiration: 5 * time.Minute, // per 5 minutes
		Message:    "Too many authentication attempts, please try again later",
	}
}

// NewRateLimiter creates a new rate limiter middleware
func NewRateLimiter(config RateLimiterConfig) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        config.Max,
		Expiration: config.Expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as key
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"error": fiber.Map{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": config.Message,
				},
			})
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Use in-memory storage (can be replaced with Redis)
	})
}

// NewGlobalRateLimiter creates global rate limiter for all routes
func NewGlobalRateLimiter() fiber.Handler {
	return NewRateLimiter(DefaultRateLimiterConfig())
}

// NewAuthRateLimiter creates rate limiter for authentication routes
func NewAuthRateLimiter() fiber.Handler {
	return NewRateLimiter(AuthRateLimiterConfig())
}

// NewStrictRateLimiter creates strict rate limiter for sensitive operations
func NewStrictRateLimiter() fiber.Handler {
	return NewRateLimiter(StrictRateLimiterConfig())
}
