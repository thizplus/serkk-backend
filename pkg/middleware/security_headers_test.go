package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeaders(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewSecurityHeaders())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Check security headers
	assert.Equal(t, "DENY", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", resp.Header.Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"))
	assert.NotEmpty(t, resp.Header.Get("Content-Security-Policy"))
	assert.NotEmpty(t, resp.Header.Get("Referrer-Policy"))
	assert.NotEmpty(t, resp.Header.Get("Permissions-Policy"))

	// X-Powered-By should be empty
	assert.Empty(t, resp.Header.Get("X-Powered-By"))
}

func TestSecurityHeadersWithConfig(t *testing.T) {
	// Setup
	config := SecurityHeadersConfig{
		XFrameOptions:       "SAMEORIGIN",
		XContentTypeOptions: "nosniff",
		XSSProtection:       "1; mode=block",
	}

	app := fiber.New()
	app.Use(NewSecurityHeadersWithConfig(config))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Act
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "SAMEORIGIN", resp.Header.Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", resp.Header.Get("X-Content-Type-Options"))
}

func TestDefaultSecurityHeadersConfig(t *testing.T) {
	config := DefaultSecurityHeadersConfig()

	assert.Equal(t, "DENY", config.XFrameOptions)
	assert.Equal(t, "nosniff", config.XContentTypeOptions)
	assert.Equal(t, "1; mode=block", config.XSSProtection)
	assert.NotEmpty(t, config.StrictTransportSecurity)
	assert.NotEmpty(t, config.ContentSecurityPolicy)
	assert.NotEmpty(t, config.ReferrerPolicy)
	assert.NotEmpty(t, config.PermissionsPolicy)
}
