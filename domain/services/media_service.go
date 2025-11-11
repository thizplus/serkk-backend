package services

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"mime/multipart"
)

type MediaService interface {
	// Upload media
	UploadImage(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)
	UploadVideo(ctx context.Context, userID uuid.UUID, file *multipart.FileHeader) (*dto.MediaUploadResponse, error)

	// Create video with source tracking (for polymorphic video encoding)
	CreateVideo(ctx context.Context, userID uuid.UUID, sourceType string, sourceID *uuid.UUID, file multipart.File, filename string) (*dto.MediaResponse, error)
	UpdateSourceID(ctx context.Context, mediaID uuid.UUID, sourceID uuid.UUID) error

	// Get media
	GetMedia(ctx context.Context, mediaID uuid.UUID) (*dto.MediaResponse, error)
	GetUserMedia(ctx context.Context, userID uuid.UUID, offset, limit int) (*dto.MediaListResponse, error)
	GetUserMediaByType(ctx context.Context, userID uuid.UUID, mediaType string, offset, limit int) (*dto.MediaListResponse, error)

	// Delete media
	DeleteMedia(ctx context.Context, mediaID uuid.UUID, userID uuid.UUID) error

	// Cleanup unused media (for background jobs)
	CleanupUnusedMedia(ctx context.Context, olderThanDays int) (int, error)

	// Video encoding status
	UpdateVideoEncodingStatus(ctx context.Context, videoID string, status string, progress int, width int, height int, duration int) error
	GetEncodingStatus(ctx context.Context, mediaID uuid.UUID) (*dto.VideoEncodingStatusResponse, error)
	GetMediaByVideoID(ctx context.Context, videoID string) (*dto.MediaResponse, error)
}
