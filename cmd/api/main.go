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

	h := handlers.NewHandlers(services, container.GetConfig(), chatWSHandler, notificationWSHandler, container.ChatHub, container.NotificationHub, container.ConversationRepository, container.MediaUploadService, container.R2Storage, container.MediaRepository, container.RedisService)

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
	log.Printf("📖 API docs: http://localhost:%s/api/v1", port)
	log.Printf("💬 WebSocket Chat: ws://localhost:%s/ws/chat", port)
	log.Printf("🔔 WebSocket Notifications: ws://localhost:%s/ws/notifications", port)

	log.Fatal(app.Listen(":" + port))
}

func setupGracefulShutdown(container *di.Container) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("\n🛑 Gracefully shutting down...")

		if err := container.Cleanup(); err != nil {
			log.Printf("❌ Error during cleanup: %v", err)
		}

		log.Println("👋 Shutdown complete")
		os.Exit(0)
	}()
}