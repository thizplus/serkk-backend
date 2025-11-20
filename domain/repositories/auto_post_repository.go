package repositories

import (
	"context"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
)

type AutoPostSettingRepository interface {
	// Auto-Post Settings CRUD
	Create(ctx context.Context, setting *models.AutoPostSetting) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AutoPostSetting, error)
	GetByBotUserID(ctx context.Context, botUserID uuid.UUID) (*models.AutoPostSetting, error)
	Update(ctx context.Context, setting *models.AutoPostSetting) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.AutoPostSetting, int64, error)

	// Get all enabled settings
	GetAllEnabled(ctx context.Context) ([]*models.AutoPostSetting, error)

	// Update statistics
	IncrementPostCount(ctx context.Context, id uuid.UUID) error
	UpdateLastError(ctx context.Context, id uuid.UUID, errorMsg *string) error
}

type AutoPostLogRepository interface {
	// Auto-Post Log CRUD
	Create(ctx context.Context, log *models.AutoPostLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AutoPostLog, error)
	Update(ctx context.Context, log *models.AutoPostLog) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List logs with filters
	ListBySettingID(ctx context.Context, settingID uuid.UUID, offset, limit int) ([]*models.AutoPostLog, int64, error)
	ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.AutoPostLog, int64, error)
	List(ctx context.Context, offset, limit int) ([]*models.AutoPostLog, int64, error)

	// Statistics
	CountBySettingID(ctx context.Context, settingID uuid.UUID) (int64, error)
	CountByStatus(ctx context.Context, settingID uuid.UUID, status string) (int64, error)
}
