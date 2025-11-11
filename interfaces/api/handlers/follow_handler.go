package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/services"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
)

type FollowHandler struct {
	followService services.FollowService
}

func NewFollowHandler(followService services.FollowService) *FollowHandler {
	return &FollowHandler{
		followService: followService,
	}
}

// Follow follows a user
func (h *FollowHandler) Follow(c *fiber.Ctx) error {
	followerID := c.Locals("userID").(uuid.UUID)

	followingID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	follow, err := h.followService.Follow(c.Context(), followerID, followingID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to follow user").WithInternal(err))
	}

	return utils.SuccessResponse(c, follow, "Followed successfully")
}

// Unfollow unfollows a user
func (h *FollowHandler) Unfollow(c *fiber.Ctx) error {
	followerID := c.Locals("userID").(uuid.UUID)

	followingID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	err = h.followService.Unfollow(c.Context(), followerID, followingID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to unfollow user").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Unfollowed successfully")
}

// IsFollowing checks if current user is following another user
func (h *FollowHandler) IsFollowing(c *fiber.Ctx) error {
	followerID := c.Locals("userID").(uuid.UUID)

	followingID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	status, err := h.followService.IsFollowing(c.Context(), followerID, followingID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to check follow status").WithInternal(err))
	}

	return utils.SuccessResponse(c, status, "Follow status retrieved successfully")
}

// GetFollowers retrieves followers of a user
func (h *FollowHandler) GetFollowers(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Get current user ID if authenticated (optional)
	var currentUserIDPtr *uuid.UUID
	if currentUserID, ok := c.Locals("userID").(uuid.UUID); ok {
		currentUserIDPtr = &currentUserID
	}

	followers, err := h.followService.GetFollowers(c.Context(), userID, offset, limit, currentUserIDPtr)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve followers").WithInternal(err))
	}

	return utils.SuccessResponse(c, followers, "Followers retrieved successfully")
}

// GetFollowing retrieves users that a user is following
func (h *FollowHandler) GetFollowing(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Get current user ID if authenticated (optional)
	var currentUserIDPtr *uuid.UUID
	if currentUserID, ok := c.Locals("userID").(uuid.UUID); ok {
		currentUserIDPtr = &currentUserID
	}

	following, err := h.followService.GetFollowing(c.Context(), userID, offset, limit, currentUserIDPtr)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve following").WithInternal(err))
	}

	return utils.SuccessResponse(c, following, "Following retrieved successfully")
}

// GetMutualFollows retrieves mutual follows (friends)
func (h *FollowHandler) GetMutualFollows(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	mutuals, err := h.followService.GetMutualFollows(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve mutual follows").WithInternal(err))
	}

	return utils.SuccessResponse(c, mutuals, "Mutual follows retrieved successfully")
}
