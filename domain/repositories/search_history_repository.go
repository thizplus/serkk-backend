package repositories

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

// PopularSearchResult - Result for popular searches with count
type PopularSearchResult struct {
	Query string
	Count int64
}

type SearchHistoryRepository interface {
	// Create search history entry
	Create(ctx context.Context, history *models.SearchHistory) error

	// Get user's search history
	ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.SearchHistory, error)

	// Get popular searches (global) - deprecated, use GetPopularSearchesWithCount
	GetPopularSearches(ctx context.Context, limit int) ([]string, error)
	// Get popular searches with count (recommended)
	GetPopularSearchesWithCount(ctx context.Context, limit int) ([]PopularSearchResult, error)

	// Get recent searches by type
	ListByUserAndType(ctx context.Context, userID uuid.UUID, searchType string, limit int) ([]*models.SearchHistory, error)

	// Clear history
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}
