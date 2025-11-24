package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/utils"
)

// UserProfileHandler handles user profile operations
// Combines data from users_identity (Auth Service) and user_profiles (Backend Service)
type UserProfileHandler struct {
	usersIdentityService *services.UsersIdentityService
	userProfileService   services.UserProfileService
}

// NewUserProfileHandler creates a new user profile handler
func NewUserProfileHandler(
	usersIdentityService *services.UsersIdentityService,
	userProfileService services.UserProfileService,
) *UserProfileHandler {
	return &UserProfileHandler{
		usersIdentityService: usersIdentityService,
		userProfileService:   userProfileService,
	}
}

// GetMe returns the current authenticated user's complete profile
// GET /api/v1/users/me
func (h *UserProfileHandler) GetMe(c *fiber.Ctx) error {
	// Get user ID from JWT token (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	ctx := c.Context()

	// Get basic user info from users_identity
	identity, err := h.usersIdentityService.GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get social profile data from user_profiles
	profile, err := h.userProfileService.GetProfile(ctx, userID)
	if err != nil {
		// Profile might not exist yet (e.g., newly registered user)
		// Return basic identity info without profile data
		return utils.SuccessResponse(c, dto.IdentityWithProfileToUserResponse(identity, nil), "User profile retrieved")
	}

	// Combine users_identity and user_profiles data
	response := dto.IdentityWithProfileToUserResponse(identity, profile)
	return utils.SuccessResponse(c, response, "User profile retrieved")
}

// GetProfile returns a user's public profile by username
// GET /api/v1/profiles/:username
func (h *UserProfileHandler) GetProfile(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username is required",
		})
	}

	ctx := c.Context()

	// Get user from users_identity
	identity, err := h.usersIdentityService.GetByUsername(ctx, username)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get profile from user_profiles
	profile, err := h.userProfileService.GetProfile(ctx, identity.ID)
	if err != nil {
		// Return basic info if profile doesn't exist
		return utils.SuccessResponse(c, dto.IdentityWithProfileToUserResponse(identity, nil), "User profile retrieved")
	}

	// Combine data
	response := dto.IdentityWithProfileToUserResponse(identity, profile)

	// Add isFollowing if user is authenticated
	if userID, ok := c.Locals("userID").(uuid.UUID); ok && userID != uuid.Nil {
		// TODO: Check if current user is following this profile
		// This requires FollowService
		isFollowing := false
		response.IsFollowing = &isFollowing
	}

	return utils.SuccessResponse(c, response, "User profile retrieved")
}

// UpdateProfile updates the current user's complete profile
// PATCH /api/v1/users/profile
// Now accepts displayName and avatar (moved from Auth Service for simpler updates)
func (h *UserProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	// Get user ID from JWT token
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse request body
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	ctx := c.Context()

	// Update user_profiles (now includes displayName and avatar)
	profile, err := h.userProfileService.UpdateProfile(
		ctx,
		userID,
		req.DisplayName, // ⭐ New: now handled in user_profiles
		req.Avatar,      // ⭐ New: now handled in user_profiles
		req.Bio,
		req.Location,
		req.Website,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	// Get user identity (for email, username)
	identity, err := h.usersIdentityService.GetByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	// Combine users_identity and user_profiles data
	// Note: displayName and avatar now come from user_profiles
	response := dto.IdentityWithProfileToUserResponse(identity, profile)
	return utils.SuccessResponse(c, response, "Profile updated successfully")
}
