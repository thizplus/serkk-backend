package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:5174",
		},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
		},
		MaxAge: 3600, // 1 hour
	}
}

// ProductionCORSConfig returns production CORS configuration
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// NewCORS creates CORS middleware with environment-aware config
func NewCORS() fiber.Handler {
	// Check for production origins from environment
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		// Fallback to FRONTEND_URL if CORS_ALLOWED_ORIGINS not set
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL != "" {
			allowedOriginsEnv = frontendURL
		}
	}

	// If environment variables are set, use production config
	if allowedOriginsEnv != "" {
		origins := strings.Split(allowedOriginsEnv, ",")
		// Trim spaces from origins
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		return NewProductionCORS(origins)
	}

	// Otherwise use default (development) config
	config := DefaultCORSConfig()
	return NewCORSWithConfig(config)
}

// NewCORSWithConfig creates CORS middleware with custom config
func NewCORSWithConfig(config CORSConfig) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(config.AllowOrigins, ","),
		AllowMethods:     strings.Join(config.AllowMethods, ","),
		AllowHeaders:     strings.Join(config.AllowHeaders, ","),
		AllowCredentials: config.AllowCredentials,
		ExposeHeaders:    strings.Join(config.ExposeHeaders, ","),
		MaxAge:           config.MaxAge,
	})
}

// NewProductionCORS creates CORS middleware for production
func NewProductionCORS(allowedOrigins []string) fiber.Handler {
	config := ProductionCORSConfig(allowedOrigins)
	return NewCORSWithConfig(config)
}
