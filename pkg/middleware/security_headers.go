package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeadersConfig holds security headers configuration
type SecurityHeadersConfig struct {
	XFrameOptions             string
	XContentTypeOptions       string
	XSSProtection             string
	StrictTransportSecurity   string
	ContentSecurityPolicy     string
	ReferrerPolicy            string
	PermissionsPolicy         string
}

// DefaultSecurityHeadersConfig returns default security headers
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		XFrameOptions:           "DENY",
		XContentTypeOptions:     "nosniff",
		XSSProtection:           "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains",
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
	}
}

// NewSecurityHeaders creates security headers middleware
func NewSecurityHeaders() fiber.Handler {
	config := DefaultSecurityHeadersConfig()

	return func(c *fiber.Ctx) error {
		// X-Frame-Options: Prevent clickjacking
		c.Set("X-Frame-Options", config.XFrameOptions)

		// X-Content-Type-Options: Prevent MIME type sniffing
		c.Set("X-Content-Type-Options", config.XContentTypeOptions)

		// X-XSS-Protection: Enable XSS filter
		c.Set("X-XSS-Protection", config.XSSProtection)

		// Strict-Transport-Security: Force HTTPS
		c.Set("Strict-Transport-Security", config.StrictTransportSecurity)

		// Content-Security-Policy: Control resources
		c.Set("Content-Security-Policy", config.ContentSecurityPolicy)

		// Referrer-Policy: Control referrer information
		c.Set("Referrer-Policy", config.ReferrerPolicy)

		// Permissions-Policy: Control browser features
		c.Set("Permissions-Policy", config.PermissionsPolicy)

		// Remove X-Powered-By header
		c.Set("X-Powered-By", "")

		return c.Next()
	}
}

// NewSecurityHeadersWithConfig creates security headers middleware with custom config
func NewSecurityHeadersWithConfig(config SecurityHeadersConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if config.XFrameOptions != "" {
			c.Set("X-Frame-Options", config.XFrameOptions)
		}
		if config.XContentTypeOptions != "" {
			c.Set("X-Content-Type-Options", config.XContentTypeOptions)
		}
		if config.XSSProtection != "" {
			c.Set("X-XSS-Protection", config.XSSProtection)
		}
		if config.StrictTransportSecurity != "" {
			c.Set("Strict-Transport-Security", config.StrictTransportSecurity)
		}
		if config.ContentSecurityPolicy != "" {
			c.Set("Content-Security-Policy", config.ContentSecurityPolicy)
		}
		if config.ReferrerPolicy != "" {
			c.Set("Referrer-Policy", config.ReferrerPolicy)
		}
		if config.PermissionsPolicy != "" {
			c.Set("Permissions-Policy", config.PermissionsPolicy)
		}

		c.Set("X-Powered-By", "")

		return c.Next()
	}
}
