package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gofiber-template/pkg/metrics"
)

// NewMetricsMiddleware creates a middleware that tracks request metrics
func NewMetricsMiddleware(m *metrics.Metrics) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Increment total requests
		m.IncrementRequests()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		durationMs := uint64(duration.Milliseconds())

		// Record response time
		m.RecordResponseTime(durationMs)

		// Get status code
		statusCode := c.Response().StatusCode()

		// Record status code
		m.RecordStatusCode(statusCode)

		// Increment counters based on result
		if statusCode >= 500 {
			m.IncrementFailed()
			m.IncrementErrors()
		} else if statusCode >= 400 {
			m.IncrementFailed()
		} else {
			m.IncrementSuccessful()
		}

		// Record error if present
		if err != nil {
			m.IncrementErrors()
		}

		return err
	}
}
