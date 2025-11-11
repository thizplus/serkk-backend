package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/services"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
)

type SavedPostHandler struct {
	savedPostService services.SavedPostService
}

func NewSavedPostHandler(savedPostService services.SavedPostService) *SavedPostHandler {
	return &SavedPostHandler{
		savedPostService: savedPostService,
	}
}

// SavePost saves a post
func (h *SavedPostHandler) SavePost(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	postID, err := uuid.Parse(c.Params("postId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid post ID")
	}

	savedPost, err := h.savedPostService.SavePost(c.Context(), userID, postID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to save post").WithInternal(err))
	}

	return utils.SuccessResponse(c, savedPost, "Post saved successfully")
}

// UnsavePost unsaves a post
func (h *SavedPostHandler) UnsavePost(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	postID, err := uuid.Parse(c.Params("postId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid post ID")
	}

	err = h.savedPostService.UnsavePost(c.Context(), userID, postID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to unsave post").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Post unsaved successfully")
}

// IsSaved checks if a post is saved
func (h *SavedPostHandler) IsSaved(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	postID, err := uuid.Parse(c.Params("postId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid post ID")
	}

	status, err := h.savedPostService.IsSaved(c.Context(), userID, postID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to check saved status").WithInternal(err))
	}

	return utils.SuccessResponse(c, status, "Saved status retrieved successfully")
}

// GetSavedPosts retrieves all saved posts
func (h *SavedPostHandler) GetSavedPosts(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	savedPosts, err := h.savedPostService.GetSavedPosts(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve saved posts").WithInternal(err))
	}

	return utils.SuccessResponse(c, savedPosts, "Saved posts retrieved successfully")
}
