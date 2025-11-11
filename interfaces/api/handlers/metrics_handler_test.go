package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gofiber-template/pkg/metrics"
)

func TestMetricsHandler_GetMetrics(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	m.IncrementRequests()
	m.IncrementSuccessful()
	m.RecordStatusCode(200)
	m.RecordResponseTime(100)

	handler := NewMetricsHandler(m)
	app := fiber.New()
	app.Get("/metrics", handler.GetMetrics)

	// Test
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	body, _ := io.ReadAll(resp.Body)
	var snapshot metrics.MetricsSnapshot
	err = json.Unmarshal(body, &snapshot)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), snapshot.TotalRequests)
	assert.Equal(t, uint64(1), snapshot.SuccessfulReqs)
	assert.Equal(t, uint64(1), snapshot.Status2xx)
}

func TestMetricsHandler_GetMetrics_Empty(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	handler := NewMetricsHandler(m)
	app := fiber.New()
	app.Get("/metrics", handler.GetMetrics)

	// Test
	req := httptest.NewRequest("GET", "/metrics", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	body, _ := io.ReadAll(resp.Body)
	var snapshot metrics.MetricsSnapshot
	err = json.Unmarshal(body, &snapshot)

	assert.NoError(t, err)
	assert.Equal(t, uint64(0), snapshot.TotalRequests)
}

func TestMetricsHandler_ResetMetrics(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	m.IncrementRequests()
	m.IncrementSuccessful()

	handler := NewMetricsHandler(m)
	app := fiber.New()
	app.Post("/metrics/reset", handler.ResetMetrics)

	// Test reset
	req := httptest.NewRequest("POST", "/metrics/reset", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify metrics are reset
	snapshot := m.GetSnapshot()
	assert.Equal(t, uint64(0), snapshot.TotalRequests)
	assert.Equal(t, uint64(0), snapshot.SuccessfulReqs)
}
