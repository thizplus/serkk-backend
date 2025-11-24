package routes

import (
	"github.com/gofiber/fiber/v2"
	"gofiber-template/interfaces/api/handlers"
	"gofiber-template/interfaces/api/middleware"
)

// SetupUserProfileRoutes sets up user profile routes
// These routes handle user profile operations (combining users_cache + user_profiles)
func SetupUserProfileRoutes(api fiber.Router, h *handlers.Handlers) {
	// Public routes
	profiles := api.Group("/profiles")
	{
		// Get user profile by username (public)
		profiles.Get("/:username", h.UserProfileHandler.GetProfile)
	}

	// Protected routes (require authentication)
	users := api.Group("/users")
	users.Use(middleware.Protected()) // Require valid JWT token
	{
		// Get current user's complete profile
		users.Get("/me", h.UserProfileHandler.GetMe)

		// Update current user's social profile (bio, location, website)
		// Note: displayName and avatar should be updated via Auth Service
		users.Patch("/profile", h.UserProfileHandler.UpdateProfile)
	}
}
