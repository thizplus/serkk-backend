package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/pkg/health"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	checker *health.HealthChecker
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(checker *health.HealthChecker) *HealthHandler {
	return &HealthHandler{
		checker: checker,
	}
}

// Check handles GET /health
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx := c.Context()

	// Perform health check
	result := h.checker.Check(ctx)

	// Return appropriate status code based on health status
	statusCode := fiber.StatusOK
	if result.Status == health.StatusUnhealthy {
		statusCode = fiber.StatusServiceUnavailable
	} else if result.Status == health.StatusDegraded {
		statusCode = fiber.StatusOK // Still return 200 for degraded
	}

	return c.Status(statusCode).JSON(result)
}

// Live handles GET /health/live (liveness probe)
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	// Simple liveness check - just return OK if server is running
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

// Ready handles GET /health/ready (readiness probe)
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx := c.Context()

	// Perform health check to determine readiness
	result := h.checker.Check(ctx)

	if result.Status == health.StatusUnhealthy {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status": "not_ready",
			"checks": result.Checks,
		})
	}

	return c.JSON(fiber.Map{
		"status": "ready",
		"checks": result.Checks,
	})
}
