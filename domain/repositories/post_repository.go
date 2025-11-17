package repositories

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/pkg/utils"
)

type PostSortBy string

const (
	SortByHot           PostSortBy = "hot"           // votes / (hours + 2)^1.5
	SortByNew           PostSortBy = "new"           // created_at DESC
	SortByTop           PostSortBy = "top"           // votes DESC
	SortByControversial PostSortBy = "controversial" // high engagement but mixed votes
)

type PostRepository interface {
	// Basic CRUD
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error)
	GetByClientPostID(ctx context.Context, clientPostID string) (*models.Post, error) // For idempotency check
	Update(ctx context.Context, id uuid.UUID, post *models.Post) error
	Delete(ctx context.Context, id uuid.UUID) error // Soft delete

	// List & Filter (offset-based, deprecated)
	List(ctx context.Context, offset, limit int, sortBy PostSortBy) ([]*models.Post, error)
	ListByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*models.Post, error)
	ListByTag(ctx context.Context, tagName string, offset, limit int, sortBy PostSortBy) ([]*models.Post, error)
	ListByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy PostSortBy) ([]*models.Post, error)

	// List with Cursor (cursor-based pagination)
	ListWithCursor(ctx context.Context, cursor *utils.PostCursor, limit int, sortBy PostSortBy) ([]*models.Post, error)
	ListByAuthorWithCursor(ctx context.Context, authorID uuid.UUID, cursor *utils.PostCursor, limit int) ([]*models.Post, error)
	ListByTagWithCursor(ctx context.Context, tagName string, cursor *utils.PostCursor, limit int, sortBy PostSortBy) ([]*models.Post, error)
	ListFollowingFeedWithCursor(ctx context.Context, userID uuid.UUID, cursor *utils.PostCursor, limit int) ([]*models.Post, error)

	// Search (offset-based, deprecated)
	Search(ctx context.Context, query string, offset, limit int) ([]*models.Post, error)
	// Search with cursor (recommended)
	SearchWithCursor(ctx context.Context, query string, cursor *utils.PostCursor, limit int) ([]*models.Post, error)

	// Crosspost
	GetCrossposts(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error)

	// Stats
	Count(ctx context.Context) (int64, error)
	CountByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error)

	// Comment count management
	IncrementCommentCount(ctx context.Context, postID uuid.UUID) error
	DecrementCommentCount(ctx context.Context, postID uuid.UUID) error

	// Vote management
	UpdateVoteCount(ctx context.Context, postID uuid.UUID, voteChange int) error

	// Media association
	AttachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error
	DetachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error
	GetPostsByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*models.Post, error)

	// Tag association
	AttachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error
	DetachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error
	SyncTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error
}
