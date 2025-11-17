package repositories

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/pkg/utils"
)

type SavedPostRepository interface {
	// Save/Unsave post
	SavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error
	UnsavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error

	// Check if saved
	IsSaved(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (bool, error)

	// Get saved posts (offset-based, deprecated)
	GetSavedPosts(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Post, error)
	CountSavedPosts(ctx context.Context, userID uuid.UUID) (int64, error)

	// Get saved posts with cursor (cursor-based pagination)
	GetSavedPostsWithCursor(ctx context.Context, userID uuid.UUID, cursor *utils.PostCursor, limit int) ([]*models.Post, error)

	// Batch check (for checking multiple posts at once)
	GetSavedStatus(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}
