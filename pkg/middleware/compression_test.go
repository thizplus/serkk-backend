package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestCompression_Success(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(strings.Repeat("Hello World! ", 100)) // >1KB response
	})

	// Test with gzip accept encoding
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	// Verify response can be decompressed
	reader, err := gzip.NewReader(resp.Body)
	assert.NoError(t, err)
	defer reader.Close()

	body, err := io.ReadAll(reader)
	assert.NoError(t, err)
	assert.True(t, len(body) > 0)
}

func TestCompression_SkipWebSocket(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/ws", func(c *fiber.Ctx) error {
		return c.SendString("WebSocket upgrade")
	})

	// Test with WebSocket upgrade header
	req := httptest.NewRequest("GET", "/ws", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Upgrade", "websocket")
	resp, err := app.Test(req)

	// Assert - should NOT be compressed
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEqual(t, "gzip", resp.Header.Get("Content-Encoding"))
}

func TestCompression_SmallResponse(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Hi") // Small response
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req)

	// Assert - might not be compressed (too small)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestBestSpeedCompression(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewBestSpeedCompression())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(strings.Repeat("Test Data ", 200))
	})

	// Test
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
}

func TestCompression_NoAcceptEncoding(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	// Test without Accept-Encoding header
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)

	// Assert - should NOT be compressed
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEqual(t, "gzip", resp.Header.Get("Content-Encoding"))

	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, "Hello World", string(body))
}

func TestCompression_JSONResponse(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/api/data", func(c *fiber.Ctx) error {
		// Large JSON response
		data := make([]map[string]string, 100)
		for i := 0; i < 100; i++ {
			data[i] = map[string]string{
				"id":    "test",
				"name":  "Test User",
				"email": "test@example.com",
			}
		}
		return c.JSON(data)
	})

	// Test
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

	// Verify compressed size is smaller
	compressedSize := resp.ContentLength

	// Get uncompressed size for comparison
	reader, _ := gzip.NewReader(resp.Body)
	uncompressed, _ := io.ReadAll(reader)
	reader.Close()

	assert.True(t, int64(len(uncompressed)) > compressedSize || compressedSize == -1)
}

func TestCompression_BinaryData(t *testing.T) {
	// Setup
	app := fiber.New()
	app.Use(NewCompression())
	app.Get("/binary", func(c *fiber.Ctx) error {
		// Binary data (less compressible)
		data := bytes.Repeat([]byte{0, 1, 2, 3, 4, 5}, 200)
		return c.Send(data)
	})

	// Test
	req := httptest.NewRequest("GET", "/binary", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := app.Test(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
