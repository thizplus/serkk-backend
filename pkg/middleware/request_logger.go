package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gofiber-template/pkg/logger"
)

// RequestLoggerConfig holds request logger configuration
type RequestLoggerConfig struct {
	SkipPaths      []string
	LogHeaders     bool
	LogBody        bool
	LogResponse    bool
	MaxBodyLogSize int
}

// DefaultRequestLoggerConfig returns default request logger configuration
func DefaultRequestLoggerConfig() RequestLoggerConfig {
	return RequestLoggerConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
		},
		LogHeaders:     false,
		LogBody:        false,
		LogResponse:    false,
		MaxBodyLogSize: 1024, // 1KB
	}
}

// NewRequestLogger creates a request logging middleware
func NewRequestLogger(log *logger.Logger) fiber.Handler {
	config := DefaultRequestLoggerConfig()
	return NewRequestLoggerWithConfig(log, config)
}

// NewRequestLoggerWithConfig creates request logger with custom config
func NewRequestLoggerWithConfig(log *logger.Logger, config RequestLoggerConfig) fiber.Handler {
	// Create skip path map for fast lookup
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *fiber.Ctx) error {
		// Skip if path is in skip list
		if skipPaths[c.Path()] {
			return c.Next()
		}

		// Start timer
		start := time.Now()

		// Get request ID (if exists from header or generate new)
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = c.Get("X-Correlation-ID")
		}

		// Create logger with request context
		reqLogger := log.WithFields(map[string]interface{}{
			"request_id": requestID,
			"method":     c.Method(),
			"path":       c.Path(),
			"ip":         c.IP(),
			"user_agent": c.Get("User-Agent"),
		})

		// Log headers if enabled
		if config.LogHeaders {
			headers := make(map[string]string)
			c.Request().Header.VisitAll(func(key, value []byte) {
				headers[string(key)] = string(value)
			})
			reqLogger = reqLogger.WithField("headers", headers)
		}

		// Log request body if enabled (for POST, PUT, PATCH)
		if config.LogBody && (c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH") {
			body := c.Body()
			if len(body) > 0 && len(body) <= config.MaxBodyLogSize {
				reqLogger = reqLogger.WithField("body", string(body))
			}
		}

		// Log incoming request
		reqLogger.Info("Incoming request")

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Create response logger
		resLogger := reqLogger.WithFields(map[string]interface{}{
			"status":   c.Response().StatusCode(),
			"duration": duration.Milliseconds(),
		})

		// Log response body if enabled
		if config.LogResponse {
			body := c.Response().Body()
			if len(body) > 0 && len(body) <= config.MaxBodyLogSize {
				resLogger = resLogger.WithField("response_body", string(body))
			}
		}

		// Log based on status code
		statusCode := c.Response().StatusCode()
		if statusCode >= 500 {
			resLogger.Error("Request completed with server error")
		} else if statusCode >= 400 {
			resLogger.Warn("Request completed with client error")
		} else {
			resLogger.Info("Request completed successfully")
		}

		return err
	}
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = generateRequestID()
		}

		// Set request ID in context and response header
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
