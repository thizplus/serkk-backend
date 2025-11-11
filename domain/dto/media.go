package dto

import (
	"time"

	"github.com/google/uuid"
)

// MediaUploadResponse - Response for media upload
type MediaUploadResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	FileName  string    `json:"fileName"`
	MimeType  string    `json:"mimeType"`
	Size      int64     `json:"size"`
	URL       string    `json:"url"`
	Thumbnail string    `json:"thumbnail,omitempty"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	Duration  float64   `json:"duration,omitempty"` // For videos
	CreatedAt time.Time `json:"createdAt"`
}

// MediaResponse - Response for a single media
type MediaResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"userId,omitempty"` // Owner of the media
	Type       string     `json:"type"`             // "image", "video", "file"
	FileName   string     `json:"fileName"`
	MimeType   string     `json:"mimeType"`
	Size       int64      `json:"size"`
	URL        string     `json:"url"`
	Thumbnail  string     `json:"thumbnail,omitempty"`
	Width      int        `json:"width,omitempty"`
	Height     int        `json:"height,omitempty"`
	Duration   float64    `json:"duration,omitempty"`   // For videos
	SourceType *string    `json:"sourceType,omitempty"` // "post", "message", "reel", "comment"
	SourceID   *uuid.UUID `json:"sourceId,omitempty"`   // ID of source entity
	CreatedAt  time.Time  `json:"createdAt"`
}

// MediaListResponse - Response for listing media
type MediaListResponse struct {
	Media []MediaResponse `json:"media"`
	Meta  PaginationMeta  `json:"meta"`
}

// MediaUploadRequest - Request for uploading media (multipart form)
type MediaUploadRequest struct {
	Type string `form:"type" validate:"required,oneof=image video"` // Validated from form
	// File will be handled separately via c.FormFile("file")
}

// VideoEncodingStatusResponse - Response for video encoding status
// NOTE: Currently not used as we don't encode videos (R2 direct play)
// Keeping for potential future Cloudflare Stream integration
type VideoEncodingStatusResponse struct {
	MediaID   uuid.UUID `json:"mediaId"`
	URL       string    `json:"url"`
	Thumbnail string    `json:"thumbnail,omitempty"`
	Width     int       `json:"width,omitempty"`
	Height    int       `json:"height,omitempty"`
	Duration  float64   `json:"duration,omitempty"`
}

// Note: Helper functions MediaToMediaUploadResponse and MediaToMediaResponse
// are implemented in the service layer to avoid circular dependencies with models package
