package services

import (
	"context"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

type UserProfileService interface {
	// GetProfile retrieves a user's profile
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error)

	// UpdateProfile updates a user's complete profile (displayName, avatar, bio, location, website)
	// displayName and avatar moved from Auth Service for simpler profile updates
	UpdateProfile(ctx context.Context, userID uuid.UUID, displayName, avatar, bio, location, website string) (*models.UserProfile, error)

	// UpdateKarma updates user karma (called when post/comment gets upvoted/downvoted)
	UpdateKarma(ctx context.Context, userID uuid.UUID, delta int) error

	// GetTopUsersByKarma retrieves top users by karma score
	GetTopUsersByKarma(ctx context.Context, limit int) ([]*models.UserProfile, error)
}
