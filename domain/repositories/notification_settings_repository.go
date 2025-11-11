package repositories

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

type NotificationSettingsRepository interface {
	// Create (default settings for new user)
	Create(ctx context.Context, settings *models.NotificationSettings) error

	// Get settings
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.NotificationSettings, error)

	// Update settings
	Update(ctx context.Context, userID uuid.UUID, settings *models.NotificationSettings) error

	// Check if user wants to receive specific notification type
	ShouldNotify(ctx context.Context, userID uuid.UUID, notificationType string) (bool, error)
}
