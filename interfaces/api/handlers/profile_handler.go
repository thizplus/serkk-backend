package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/services"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
)

type ProfileHandler struct {
	userService services.UserService
}

func NewProfileHandler(userService services.UserService) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
	}
}

// GetPublicProfile retrieves a user's public profile by username
func (h *ProfileHandler) GetPublicProfile(c *fiber.Ctx) error {
	username := c.Params("username")
	if username == "" {
		return utils.ValidationErrorResponse(c, "Username is required")
	}

	// Get currentUserID if authenticated (optional)
	var currentUserIDPtr *uuid.UUID
	if userID, ok := c.Locals("userID").(uuid.UUID); ok {
		currentUserIDPtr = &userID
	}

	profile, err := h.userService.GetPublicProfile(c.Context(), username, currentUserIDPtr)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage("User not found").WithInternal(err))
	}

	return utils.SuccessResponse(c, profile, "Profile retrieved successfully")
}
