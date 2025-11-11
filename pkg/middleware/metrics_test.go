package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gofiber-template/pkg/metrics"
)

func TestMetricsMiddleware_Success(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	app := fiber.New()
	app.Use(NewMetricsMiddleware(m))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	snapshot := m.GetSnapshot()
	assert.Equal(t, uint64(1), snapshot.TotalRequests)
	assert.Equal(t, uint64(1), snapshot.SuccessfulReqs)
	assert.Equal(t, uint64(1), snapshot.Status2xx)
	assert.GreaterOrEqual(t, snapshot.AvgResponseTime, float64(0))
}

func TestMetricsMiddleware_Error(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	app := fiber.New()
	app.Use(NewMetricsMiddleware(m))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.Status(500).SendString("Error")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	snapshot := m.GetSnapshot()
	assert.Equal(t, uint64(1), snapshot.TotalRequests)
	assert.Equal(t, uint64(1), snapshot.FailedReqs)
	assert.Equal(t, uint64(1), snapshot.Status5xx)
	assert.Equal(t, uint64(1), snapshot.TotalErrors)
}

func TestMetricsMiddleware_ClientError(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	app := fiber.New()
	app.Use(NewMetricsMiddleware(m))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.Status(404).SendString("Not Found")
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	snapshot := m.GetSnapshot()
	assert.Equal(t, uint64(1), snapshot.TotalRequests)
	assert.Equal(t, uint64(1), snapshot.FailedReqs)
	assert.Equal(t, uint64(1), snapshot.Status4xx)
	assert.Equal(t, uint64(0), snapshot.TotalErrors) // 4xx is not an error
}

func TestMetricsMiddleware_MultipleRequests(t *testing.T) {
	// Setup
	m := metrics.NewMetrics()
	app := fiber.New()
	app.Use(NewMetricsMiddleware(m))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test - make 5 requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		_, err := app.Test(req)
		assert.NoError(t, err)
	}

	// Assert
	snapshot := m.GetSnapshot()
	assert.Equal(t, uint64(5), snapshot.TotalRequests)
	assert.Equal(t, uint64(5), snapshot.SuccessfulReqs)
	assert.Equal(t, uint64(5), snapshot.Status2xx)
}
