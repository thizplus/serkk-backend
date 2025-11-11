package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_WithinLimit(t *testing.T) {
	// Setup
	app := fiber.New()

	config := RateLimiterConfig{
		Max:        5,
		Expiration: 1 * time.Minute,
		Message:    "Rate limit exceeded",
	}

	app.Use(NewRateLimiter(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Act - Make requests within limit
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestRateLimiter_ExceedLimit(t *testing.T) {
	// Setup
	app := fiber.New()

	config := RateLimiterConfig{
		Max:        3,
		Expiration: 1 * time.Minute,
		Message:    "Too many requests",
	}

	app.Use(NewRateLimiter(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Act - Make requests up to limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, _ := app.Test(req)
		assert.Equal(t, 200, resp.StatusCode)
	}

	// Act - Exceed limit
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 429, resp.StatusCode) // Too Many Requests
}

func TestDefaultRateLimiterConfig(t *testing.T) {
	config := DefaultRateLimiterConfig()

	assert.Equal(t, 100, config.Max)
	assert.Equal(t, 1*time.Minute, config.Expiration)
	assert.NotEmpty(t, config.Message)
}

func TestStrictRateLimiterConfig(t *testing.T) {
	config := StrictRateLimiterConfig()

	assert.Equal(t, 5, config.Max)
	assert.Equal(t, 1*time.Minute, config.Expiration)
	assert.NotEmpty(t, config.Message)
}

func TestAuthRateLimiterConfig(t *testing.T) {
	config := AuthRateLimiterConfig()

	assert.Equal(t, 10, config.Max)
	assert.Equal(t, 5*time.Minute, config.Expiration)
	assert.NotEmpty(t, config.Message)
}
