package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gorm.io/gorm"
)

type autoPostSettingRepository struct {
	db *gorm.DB
}

type autoPostLogRepository struct {
	db *gorm.DB
}

// NewAutoPostSettingRepository creates a new auto-post setting repository
func NewAutoPostSettingRepository(db *gorm.DB) repositories.AutoPostSettingRepository {
	return &autoPostSettingRepository{db: db}
}

// NewAutoPostLogRepository creates a new auto-post log repository
func NewAutoPostLogRepository(db *gorm.DB) repositories.AutoPostLogRepository {
	return &autoPostLogRepository{db: db}
}

// ========== AutoPostSettingRepository Implementation ==========

func (r *autoPostSettingRepository) Create(ctx context.Context, setting *models.AutoPostSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *autoPostSettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AutoPostSetting, error) {
	var setting models.AutoPostSetting
	err := r.db.WithContext(ctx).
		Preload("BotUser").
		First(&setting, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (r *autoPostSettingRepository) GetByBotUserID(ctx context.Context, botUserID uuid.UUID) (*models.AutoPostSetting, error) {
	var setting models.AutoPostSetting
	err := r.db.WithContext(ctx).
		Preload("BotUser").
		First(&setting, "bot_user_id = ?", botUserID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (r *autoPostSettingRepository) Update(ctx context.Context, setting *models.AutoPostSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *autoPostSettingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.AutoPostSetting{}, "id = ?", id).Error
}

func (r *autoPostSettingRepository) List(ctx context.Context, offset, limit int) ([]*models.AutoPostSetting, int64, error) {
	var settings []*models.AutoPostSetting
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&models.AutoPostSetting{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get settings
	err := r.db.WithContext(ctx).
		Preload("BotUser").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&settings).Error

	if err != nil {
		return nil, 0, err
	}

	return settings, total, nil
}

func (r *autoPostSettingRepository) GetAllEnabled(ctx context.Context) ([]*models.AutoPostSetting, error) {
	var settings []*models.AutoPostSetting
	err := r.db.WithContext(ctx).
		Preload("BotUser").
		Where("is_enabled = ?", true).
		Find(&settings).Error

	if err != nil {
		return nil, err
	}

	return settings, nil
}

func (r *autoPostSettingRepository) IncrementPostCount(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.AutoPostSetting{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"total_posts_generated": gorm.Expr("total_posts_generated + 1"),
			"last_generated_at":     now,
			"last_error":            nil, // Clear error on success
		}).Error
}

func (r *autoPostSettingRepository) UpdateLastError(ctx context.Context, id uuid.UUID, errorMsg *string) error {
	return r.db.WithContext(ctx).
		Model(&models.AutoPostSetting{}).
		Where("id = ?", id).
		Update("last_error", errorMsg).Error
}

// ========== AutoPostLogRepository Implementation ==========

func (r *autoPostLogRepository) Create(ctx context.Context, log *models.AutoPostLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *autoPostLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AutoPostLog, error) {
	var log models.AutoPostLog
	err := r.db.WithContext(ctx).
		Preload("Setting").
		Preload("Post").
		First(&log, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &log, nil
}

func (r *autoPostLogRepository) Update(ctx context.Context, log *models.AutoPostLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

func (r *autoPostLogRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.AutoPostLog{}, "id = ?", id).Error
}

func (r *autoPostLogRepository) ListBySettingID(ctx context.Context, settingID uuid.UUID, offset, limit int) ([]*models.AutoPostLog, int64, error) {
	var logs []*models.AutoPostLog
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&models.AutoPostLog{}).
		Where("setting_id = ?", settingID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get logs
	err := r.db.WithContext(ctx).
		Preload("Post").
		Preload("Post.Author").
		Where("setting_id = ?", settingID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *autoPostLogRepository) ListByStatus(ctx context.Context, status string, offset, limit int) ([]*models.AutoPostLog, int64, error) {
	var logs []*models.AutoPostLog
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).
		Model(&models.AutoPostLog{}).
		Where("status = ?", status).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get logs
	err := r.db.WithContext(ctx).
		Preload("Post").
		Preload("Post.Author").
		Where("status = ?", status).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *autoPostLogRepository) List(ctx context.Context, offset, limit int) ([]*models.AutoPostLog, int64, error) {
	var logs []*models.AutoPostLog
	var total int64

	// Count total
	if err := r.db.WithContext(ctx).Model(&models.AutoPostLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get logs
	err := r.db.WithContext(ctx).
		Preload("Post").
		Preload("Post.Author").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error

	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *autoPostLogRepository) CountBySettingID(ctx context.Context, settingID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AutoPostLog{}).
		Where("setting_id = ?", settingID).
		Count(&count).Error

	return count, err
}

func (r *autoPostLogRepository) CountByStatus(ctx context.Context, settingID uuid.UUID, status string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.AutoPostLog{}).
		Where("setting_id = ? AND status = ?", settingID, status).
		Count(&count).Error

	return count, err
}
