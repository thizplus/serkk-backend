package repositories

import (
	"context"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

type UserProfileRepository interface {
	// Create creates a new user profile
	Create(ctx context.Context, profile *models.UserProfile) error

	// GetByUserID retrieves a user profile by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)

	// Update updates a user profile
	Update(ctx context.Context, userID uuid.UUID, profile *models.UserProfile) error

	// UpdateKarma updates the karma score
	UpdateKarma(ctx context.Context, userID uuid.UUID, delta int) error

	// UpdateFollowerCount updates the followers count
	UpdateFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error

	// UpdateFollowingCount updates the following count
	UpdateFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error

	// GetTopUsersByKarma retrieves top users by karma
	GetTopUsersByKarma(ctx context.Context, limit int) ([]*models.UserProfile, error)

	// Delete deletes a user profile
	Delete(ctx context.Context, userID uuid.UUID) error
}
