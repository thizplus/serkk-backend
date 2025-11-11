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

// Check godoc
// @Summary Health check
// @Description Check overall system health
// @Tags Health
// @Produce json
// @Success 200 {object} health.HealthCheck "System is healthy"
// @Failure 503 {object} health.HealthCheck "System is unhealthy"
// @Router /health [get]
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

// Live godoc
// @Summary Liveness probe
// @Description Check if application is alive
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "Application is alive"
// @Router /health/live [get]
func (h *HealthHandler) Live(c *fiber.Ctx) error {
	// Simple liveness check - just return OK if server is running
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

// Ready godoc
// @Summary Readiness probe
// @Description Check if application is ready to serve traffic
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "Application is ready"
// @Failure 503 {object} map[string]interface{} "Application is not ready"
// @Router /health/ready [get]
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
