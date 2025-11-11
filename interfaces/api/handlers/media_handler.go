package handlers

import (
		apperrors "gofiber-template/pkg/errors"
"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/utils"
)

type MediaHandler struct {
	mediaService services.MediaService
}

func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// UploadImage uploads an image file
func (h *MediaHandler) UploadImage(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	// Get file from form
	file, err := c.FormFile("image")
	if err != nil {
		return utils.ValidationErrorResponse(c, "Image file is required")
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" && contentType != "image/webp" {
		return utils.ValidationErrorResponse(c, "Invalid image type. Supported: jpeg, png, gif, webp")
	}

	// Validate file size (10MB max)
	if file.Size > 10*1024*1024 {
		return utils.ValidationErrorResponse(c, "Image size must be less than 10MB")
	}

	media, err := h.mediaService.UploadImage(c.Context(), userID, file)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to upload image").WithInternal(err))
	}

	return utils.SuccessResponse(c, media, "Image uploaded successfully")
}

// UploadVideo uploads a video file
func (h *MediaHandler) UploadVideo(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	// Get file from form
	file, err := c.FormFile("video")
	if err != nil {
		return utils.ValidationErrorResponse(c, "Video file is required")
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "video/mp4" && contentType != "video/webm" && contentType != "video/ogg" {
		return utils.ValidationErrorResponse(c, "Invalid video type. Supported: mp4, webm, ogg")
	}

	// Validate file size (300MB max)
	if file.Size > 300*1024*1024 {
		return utils.ValidationErrorResponse(c, "Video size must be less than 300MB")
	}

	media, err := h.mediaService.UploadVideo(c.Context(), userID, file)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to upload video").WithInternal(err))
	}

	return utils.SuccessResponse(c, media, "Video uploaded successfully")
}

// GetMedia retrieves a media file by ID
func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	mediaID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid media ID")
	}

	media, err := h.mediaService.GetMedia(c.Context(), mediaID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage("Media not found").WithInternal(err))
	}

	return utils.SuccessResponse(c, media, "Media retrieved successfully")
}

// GetUserMedia retrieves all media uploaded by a user
func (h *MediaHandler) GetUserMedia(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid user ID")
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	media, err := h.mediaService.GetUserMedia(c.Context(), userID, offset, limit)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to retrieve user media").WithInternal(err))
	}

	return utils.SuccessResponse(c, media, "User media retrieved successfully")
}

// DeleteMedia deletes a media file
func (h *MediaHandler) DeleteMedia(c *fiber.Ctx) error {
	userID := c.Locals("userID").(uuid.UUID)

	mediaID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid media ID")
	}

	err = h.mediaService.DeleteMedia(c.Context(), userID, mediaID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to delete media").WithInternal(err))
	}

	return utils.SuccessResponse(c, nil, "Media deleted successfully")
}

// GetEncodingStatus retrieves video encoding status
func (h *MediaHandler) GetEncodingStatus(c *fiber.Ctx) error {
	mediaID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.ValidationErrorResponse(c, "Invalid media ID")
	}

	status, err := h.mediaService.GetEncodingStatus(c.Context(), mediaID)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrNotFound.WithMessage("Failed to get encoding status").WithInternal(err))
	}

	return utils.SuccessResponse(c, status, "Encoding status retrieved successfully")
}
