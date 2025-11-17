package services

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
)

type SavedPostService interface {
	// Save/Unsave posts
	SavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*dto.SavedPostResponse, error)
	UnsavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error

	// Check if saved
	IsSaved(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (*dto.SaveStatusResponse, error)

	// Get saved posts (offset-based, deprecated)
	GetSavedPosts(ctx context.Context, userID uuid.UUID, offset, limit int) (*dto.SavedPostListResponse, error)

	// Get saved posts with cursor (cursor-based pagination)
	GetSavedPostsWithCursor(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*dto.SavedPostListCursorResponse, error)
}
