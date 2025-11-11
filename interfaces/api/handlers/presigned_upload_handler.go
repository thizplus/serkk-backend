package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/infrastructure/storage"
)

type PresignedUploadHandler struct {
	r2Storage storage.R2Storage
	mediaRepo repositories.MediaRepository
}

func NewPresignedUploadHandler(r2Storage storage.R2Storage, mediaRepo repositories.MediaRepository) *PresignedUploadHandler {
	return &PresignedUploadHandler{
		r2Storage: r2Storage,
		mediaRepo: mediaRepo,
	}
}

// PresignedUploadRequest represents the request to generate a presigned upload URL
type PresignedUploadRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"`
	FileSize    int64  `json:"fileSize" validate:"required,min=1"`
	MediaType   string `json:"mediaType" validate:"required,oneof=image video file"`
}

// PresignedUploadResponse represents the response with presigned upload URL
type PresignedUploadResponse struct {
	UploadURL string    `json:"uploadUrl"` // Presigned URL for frontend to upload
	FileURL   string    `json:"fileUrl"`   // Final public URL after upload
	FileKey   string    `json:"fileKey"`   // R2 object key
	MediaID   uuid.UUID `json:"mediaId"`   // Media ID for tracking
	ExpiresAt time.Time `json:"expiresAt"` // When the presigned URL expires
}

// GeneratePresignedUploadURL generates a presigned upload URL for direct frontend upload
// @Summary Generate presigned upload URL
// @Description Generates a presigned URL for frontend to upload files directly to R2
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PresignedUploadRequest true "Upload request"
// @Success 200 {object} PresignedUploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/media/presigned-upload [post]
func (h *PresignedUploadHandler) GeneratePresignedUploadURL(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse request
	var req PresignedUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate file size limits
	maxSizes := map[string]int64{
		"image": 20 * 1024 * 1024,  // 20MB
		"video": 500 * 1024 * 1024, // 500MB
		"file":  100 * 1024 * 1024, // 100MB
	}

	maxSize, exists := maxSizes[req.MediaType]
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid media type",
		})
	}

	if req.FileSize > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("%s size exceeds maximum allowed (%d MB)", req.MediaType, maxSize/(1024*1024)),
		})
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(req.Filename))
	if !h.isValidExtension(req.MediaType, ext) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("invalid file extension for %s", req.MediaType),
		})
	}

	// Generate unique file key
	mediaID := uuid.New()
	fileKey := h.generateFileKey(userID, mediaID, req.MediaType, ext)

	// Generate presigned upload URL (valid for 15 minutes)
	expiresIn := 15 * time.Minute
	uploadURL, err := h.r2Storage.GeneratePresignedUploadURL(c.Context(), fileKey, req.ContentType, expiresIn)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate upload URL",
		})
	}

	// Get final public URL
	fileURL := h.r2Storage.GetPublicURL(fileKey)

	return c.JSON(PresignedUploadResponse{
		UploadURL: uploadURL,
		FileURL:   fileURL,
		FileKey:   fileKey,
		MediaID:   mediaID,
		ExpiresAt: time.Now().Add(expiresIn),
	})
}

// GenerateBatchPresignedUploadURLs generates multiple presigned upload URLs in a single request
// @Summary Generate batch presigned upload URLs
// @Description Generates multiple presigned URLs for frontend to upload files directly to R2 (max 200 files)
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchPresignedUploadRequest true "Batch upload request"
// @Success 200 {object} BatchPresignedUploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/media/presigned-upload-batch [post]
func (h *PresignedUploadHandler) GenerateBatchPresignedUploadURLs(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse request
	var req BatchPresignedUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate batch size
	if len(req.Files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "files array cannot be empty",
		})
	}

	if len(req.Files) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "maximum 200 files per batch request",
		})
	}

	// File size limits
	maxSizes := map[string]int64{
		"image": 20 * 1024 * 1024,  // 20MB
		"video": 500 * 1024 * 1024, // 500MB
		"file":  100 * 1024 * 1024, // 100MB
	}

	// Process each file request
	uploads := make([]PresignedUploadResponse, 0, len(req.Files))
	expiresIn := 15 * time.Minute
	expiresAt := time.Now().Add(expiresIn)

	for i, fileReq := range req.Files {
		// Validate media type
		maxSize, exists := maxSizes[fileReq.MediaType]
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid media type at index %d: %s", i, fileReq.MediaType),
			})
		}

		// Validate file size
		if fileReq.FileSize > maxSize {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("%s size exceeds maximum at index %d (%d MB allowed)", fileReq.MediaType, i, maxSize/(1024*1024)),
			})
		}

		// Validate file extension
		ext := strings.ToLower(filepath.Ext(fileReq.Filename))
		if !h.isValidExtension(fileReq.MediaType, ext) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("invalid file extension at index %d for %s: %s", i, fileReq.MediaType, ext),
			})
		}

		// Generate unique file key
		mediaID := uuid.New()
		fileKey := h.generateFileKey(userID, mediaID, fileReq.MediaType, ext)

		// Generate presigned upload URL
		uploadURL, err := h.r2Storage.GeneratePresignedUploadURL(c.Context(), fileKey, fileReq.ContentType, expiresIn)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   fmt.Sprintf("failed to generate upload URL at index %d", i),
				"details": err.Error(),
			})
		}

		// Get final public URL
		fileURL := h.r2Storage.GetPublicURL(fileKey)

		uploads = append(uploads, PresignedUploadResponse{
			UploadURL: uploadURL,
			FileURL:   fileURL,
			FileKey:   fileKey,
			MediaID:   mediaID,
			ExpiresAt: expiresAt,
		})
	}

	return c.JSON(BatchPresignedUploadResponse{
		Uploads: uploads,
		Total:   len(uploads),
	})
}

// BatchPresignedUploadRequest represents a batch request to generate multiple presigned upload URLs
type BatchPresignedUploadRequest struct {
	Files []PresignedUploadRequest `json:"files" validate:"required,min=1,max=200,dive"`
}

// BatchPresignedUploadResponse represents the batch response with multiple presigned upload URLs
type BatchPresignedUploadResponse struct {
	Uploads []PresignedUploadResponse `json:"uploads"`
	Total   int                       `json:"total"`
}

// ConfirmUploadRequest represents the request to confirm upload completion
type ConfirmUploadRequest struct {
	MediaID     uuid.UUID `json:"mediaId" validate:"required"`
	FileKey     string    `json:"fileKey" validate:"required"`
	FileSize    int64     `json:"fileSize" validate:"required"`
	ContentType string    `json:"contentType,omitempty"` // MIME type
	SourceType  string    `json:"sourceType,omitempty"`  // "post", "message", "reel", etc.
	SourceID    uuid.UUID `json:"sourceId,omitempty"`    // ID of the source entity
	Width       int       `json:"width,omitempty"`       // For images/videos
	Height      int       `json:"height,omitempty"`      // For images/videos
	Duration    float64   `json:"duration,omitempty"`    // For videos (seconds)
	Thumbnail   string    `json:"thumbnail,omitempty"`   // For videos
}

// ConfirmUpload confirms that the frontend has successfully uploaded the file
// This creates the media record in the database
// @Summary Confirm upload completion
// @Description Creates media record after successful frontend upload to R2
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ConfirmUploadRequest true "Confirm upload request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/media/confirm-upload [post]
func (h *PresignedUploadHandler) ConfirmUpload(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse request
	var req ConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate mediaId and fileKey
	if req.MediaID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid media ID",
		})
	}

	if req.FileKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "file key is required",
		})
	}

	// Extract file extension and type from fileKey
	ext := strings.ToLower(filepath.Ext(req.FileKey))
	extWithoutDot := strings.TrimPrefix(ext, ".")

	// Determine media type from fileKey path
	var mediaType string
	if strings.HasPrefix(req.FileKey, "images/") {
		mediaType = "image"
	} else if strings.HasPrefix(req.FileKey, "videos/") {
		mediaType = "video"
	} else {
		mediaType = "file"
	}

	// Get file URL from R2
	fileURL := h.r2Storage.GetPublicURL(req.FileKey)

	// Create media record in database
	media := &models.Media{
		ID:        req.MediaID,
		UserID:    userID,
		Type:      mediaType,
		FileName:  filepath.Base(req.FileKey),
		Extension: extWithoutDot,
		MimeType:  req.ContentType,
		Size:      req.FileSize,
		URL:       fileURL,
		Width:     req.Width,
		Height:    req.Height,
		Duration:  req.Duration,
		Thumbnail: req.Thumbnail,
		CreatedAt: time.Now(),
	}

	// Save to database
	ctx := context.Background()
	err := h.mediaRepo.Create(ctx, media)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "failed to create media record",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"mediaId": req.MediaID,
		"message": "upload confirmed",
		"fileUrl": fileURL,
	})
}

// BatchConfirmUploadRequest represents a batch request to confirm multiple uploads
type BatchConfirmUploadRequest struct {
	Uploads []ConfirmUploadRequest `json:"uploads" validate:"required,min=1,max=200,dive"`
}

// BatchConfirmUploadResponse represents the batch response for confirmed uploads
type BatchConfirmUploadResponse struct {
	Successful   []BatchConfirmResult `json:"successful"`
	Failed       []BatchConfirmResult `json:"failed"`
	Total        int                  `json:"total"`
	SuccessCount int                  `json:"successCount"`
	FailCount    int                  `json:"failCount"`
}

// BatchConfirmResult represents the result of a single file confirmation
type BatchConfirmResult struct {
	MediaID uuid.UUID `json:"mediaId"`
	FileURL string    `json:"fileUrl,omitempty"`
	FileKey string    `json:"fileKey"`
	Success bool      `json:"success"`
	Error   string    `json:"error,omitempty"`
}

// BatchConfirmUpload confirms multiple uploads at once
// @Summary Batch confirm upload completion
// @Description Creates multiple media records after successful frontend uploads to R2
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BatchConfirmUploadRequest true "Batch confirm upload request"
// @Success 200 {object} BatchConfirmUploadResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/upload/confirm-batch [post]
func (h *PresignedUploadHandler) BatchConfirmUpload(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("userID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse request
	var req BatchConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Validate batch size
	if len(req.Uploads) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uploads array cannot be empty",
		})
	}

	if len(req.Uploads) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "maximum 200 uploads per batch request",
		})
	}

	// Process each upload confirmation
	successful := make([]BatchConfirmResult, 0)
	failed := make([]BatchConfirmResult, 0)
	ctx := context.Background()

	for i, uploadReq := range req.Uploads {
		result := BatchConfirmResult{
			MediaID: uploadReq.MediaID,
			FileKey: uploadReq.FileKey,
		}

		// Validate mediaId and fileKey
		if uploadReq.MediaID == uuid.Nil {
			result.Success = false
			result.Error = fmt.Sprintf("invalid media ID at index %d", i)
			failed = append(failed, result)
			continue
		}

		if uploadReq.FileKey == "" {
			result.Success = false
			result.Error = fmt.Sprintf("file key is required at index %d", i)
			failed = append(failed, result)
			continue
		}

		// Extract file extension and type from fileKey
		ext := strings.ToLower(filepath.Ext(uploadReq.FileKey))
		extWithoutDot := strings.TrimPrefix(ext, ".")

		// Determine media type from fileKey path
		var mediaType string
		if strings.HasPrefix(uploadReq.FileKey, "images/") {
			mediaType = "image"
		} else if strings.HasPrefix(uploadReq.FileKey, "videos/") {
			mediaType = "video"
		} else {
			mediaType = "file"
		}

		// Get file URL from R2
		fileURL := h.r2Storage.GetPublicURL(uploadReq.FileKey)

		// Create media record in database
		media := &models.Media{
			ID:        uploadReq.MediaID,
			UserID:    userID,
			Type:      mediaType,
			FileName:  filepath.Base(uploadReq.FileKey),
			Extension: extWithoutDot,
			MimeType:  uploadReq.ContentType,
			Size:      uploadReq.FileSize,
			URL:       fileURL,
			Width:     uploadReq.Width,
			Height:    uploadReq.Height,
			Duration:  uploadReq.Duration,
			Thumbnail: uploadReq.Thumbnail,
			CreatedAt: time.Now(),
		}

		// Save to database
		err := h.mediaRepo.Create(ctx, media)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to create media record: %v", err)
			failed = append(failed, result)
			continue
		}

		// Success
		result.Success = true
		result.FileURL = fileURL
		successful = append(successful, result)
	}

	return c.JSON(BatchConfirmUploadResponse{
		Successful:   successful,
		Failed:       failed,
		Total:        len(req.Uploads),
		SuccessCount: len(successful),
		FailCount:    len(failed),
	})
}

// Helper functions

// generateFileKey generates a unique S3 key for the file
func (h *PresignedUploadHandler) generateFileKey(userID uuid.UUID, mediaID uuid.UUID, mediaType string, ext string) string {
	// Format: {mediaType}/{userID}/{mediaID}{ext}
	// Example: images/550e8400-e29b-41d4-a716-446655440000/123e4567-e89b-12d3-a456-426614174000.jpg
	return fmt.Sprintf("%ss/%s/%s%s", mediaType, userID.String(), mediaID.String(), ext)
}

// isValidExtension checks if the file extension is valid for the media type
func (h *PresignedUploadHandler) isValidExtension(mediaType string, ext string) bool {
	validExtensions := map[string][]string{
		"image": {".jpg", ".jpeg", ".png", ".gif", ".webp"},
		"video": {".mp4", ".mov", ".avi", ".webm"},
		"file":  {".pdf", ".doc", ".docx", ".txt", ".zip", ".rar"},
	}

	extensions, exists := validExtensions[mediaType]
	if !exists {
		return false
	}

	for _, validExt := range extensions {
		if ext == validExt {
			return true
		}
	}
	return false
}
