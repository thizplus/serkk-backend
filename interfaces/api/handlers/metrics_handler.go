package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/pkg/metrics"
)

// MetricsHandler handles metrics requests
type MetricsHandler struct {
	metrics *metrics.Metrics
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(m *metrics.Metrics) *MetricsHandler {
	return &MetricsHandler{
		metrics: m,
	}
}

// GetMetrics godoc
// @Summary Get application metrics
// @Description Get current application metrics including requests, response times, and status codes
// @Tags Metrics
// @Produce json
// @Success 200 {object} metrics.MetricsSnapshot "Current metrics"
// @Router /metrics [get]
func (h *MetricsHandler) GetMetrics(c *fiber.Ctx) error {
	snapshot := h.metrics.GetSnapshot()
	return c.JSON(snapshot)
}

// ResetMetrics handles POST /metrics/reset (should be protected)
func (h *MetricsHandler) ResetMetrics(c *fiber.Ctx) error {
	h.metrics.Reset()
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Metrics reset successfully",
	})
}
