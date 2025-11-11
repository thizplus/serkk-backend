package handlers

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gofiber-template/pkg/health"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestHealthChecker(t *testing.T) *health.HealthChecker {
	// Create in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skip("CGO not enabled, skipping health handler tests")
		return nil
	}

	checker := health.NewHealthChecker("1.0.0-test")
	checker.AddChecker(health.NewDatabaseChecker(db))
	return checker
}

func TestHealthHandler_Check(t *testing.T) {
	// Setup
	checker := setupTestHealthChecker(t)
	if checker == nil {
		return
	}

	handler := NewHealthHandler(checker)
	app := fiber.New()
	app.Get("/health", handler.Check)

	// Test
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthHandler_Live(t *testing.T) {
	// Setup
	checker := setupTestHealthChecker(t)
	if checker == nil {
		return
	}

	handler := NewHealthHandler(checker)
	app := fiber.New()
	app.Get("/health/live", handler.Live)

	// Test
	req := httptest.NewRequest("GET", "/health/live", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthHandler_Ready(t *testing.T) {
	// Setup
	checker := setupTestHealthChecker(t)
	if checker == nil {
		return
	}

	handler := NewHealthHandler(checker)
	app := fiber.New()
	app.Get("/health/ready", handler.Ready)

	// Test
	req := httptest.NewRequest("GET", "/health/ready", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthHandler_Ready_Unhealthy(t *testing.T) {
	// Setup with unhealthy checker
	checker := health.NewHealthChecker("1.0.0-test")
	checker.AddChecker(&mockUnhealthyChecker{})

	handler := NewHealthHandler(checker)
	app := fiber.New()
	app.Get("/health/ready", handler.Ready)

	// Test
	req := httptest.NewRequest("GET", "/health/ready", nil)
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 503, resp.StatusCode)
}

// mockUnhealthyChecker always returns unhealthy status
type mockUnhealthyChecker struct{}

func (m *mockUnhealthyChecker) Name() string {
	return "mock"
}

func (m *mockUnhealthyChecker) Check(ctx context.Context) health.CheckResult {
	return health.CheckResult{
		Status:  health.StatusUnhealthy,
		Message: "Mock unhealthy",
	}
}
