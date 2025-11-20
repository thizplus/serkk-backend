package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
	"gofiber-template/interfaces/api/routes"
	websocketHandler "gofiber-template/interfaces/api/websocket"
	"gofiber-template/pkg/di"
	pkgMiddleware "gofiber-template/pkg/middleware"

	_ "gofiber-template/docs" // Import generated docs
)

// @title GoFiber Social Media API
// @version 1.0
// @description Social media platform backend API with real-time chat, posts, comments, and notifications
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
	// Initialize DI container
	container := di.NewContainer()

	// Initialize all dependencies
	if err := container.Initialize(); err != nil {
		log.Fatal("Failed to initialize container:", err)
	}

	// Setup graceful shutdown
	setupGracefulShutdown(container)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler(),
		AppName:      container.GetConfig().App.Name,
		BodyLimit:    300 * 1024 * 1024, // 300 MB
	})

	// Setup middleware
	app.Use(middleware.LoggerMiddleware())

	// Compression middleware (should be early in chain)
	app.Use(pkgMiddleware.NewCompression())

	// Monitoring middlewares
	app.Use(pkgMiddleware.NewRequestLogger(container.GetLogger()))
	app.Use(pkgMiddleware.NewMetricsMiddleware(container.GetMetrics()))

	// Security middlewares
	app.Use(pkgMiddleware.NewSecurityHeaders())
	app.Use(pkgMiddleware.NewCORS())
	app.Use(pkgMiddleware.NewGlobalRateLimiter())

	// Create handlers from services
	services := container.GetHandlerServices()

	// Create WebSocket Handlers
	chatWSHandler := websocketHandler.NewChatWebSocketHandler(container.ChatHub)
	notificationWSHandler := websocketHandler.NewNotificationWebSocketHandler(container.NotificationHub)

	h := handlers.NewHandlers(services, container.GetConfig(), chatWSHandler, notificationWSHandler, container.ChatHub, container.NotificationHub, container.ConversationRepository, container.MediaUploadService, container.R2Storage, container.MediaRepository, container.RedisService, container.FeedCacheService, container.DB)

	// Create monitoring handlers
	healthHandler := handlers.NewHealthHandler(container.GetHealthChecker())
	metricsHandler := handlers.NewMetricsHandler(container.GetMetrics())

	// Setup health and metrics endpoints (before other routes)
	app.Get("/health", healthHandler.Check)
	app.Get("/health/live", healthHandler.Live)
	app.Get("/health/ready", healthHandler.Ready)
	app.Get("/metrics", metricsHandler.GetMetrics)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Setup routes
	routes.SetupRoutes(app, h)

	// Start server
	port := container.GetConfig().App.Port
	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("🌍 Environment: %s", container.GetConfig().App.Env)
	log.Printf("📚 Health check: http://localhost:%s/health", port)
	log.Printf("📖 API docs: http://localhost:%s/swagger/index.html", port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", port)
	log.Printf("💬 WebSocket Chat: ws://localhost:%s/ws/chat", port)
	log.Printf("🔔 WebSocket Notifications: ws://localhost:%s/ws/notifications", port)

	// Start server in goroutine to allow graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("\n🛑 Gracefully shutting down...")
	if err := app.Shutdown(); err != nil {
		log.Printf("❌ Error shutting down server: %v", err)
	}

	// Cleanup resources
	if err := container.Cleanup(); err != nil {
		log.Printf("❌ Error during cleanup: %v", err)
	}

	log.Println("👋 Shutdown complete")
}

func setupGracefulShutdown(container *di.Container) {
	// This function is now handled inline in main()
	// Kept for backward compatibility if needed
}
