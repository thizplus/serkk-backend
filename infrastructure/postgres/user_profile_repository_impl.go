package postgres

import (
	"context"

	"github.com/google/uuid"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gorm.io/gorm"
)

type UserProfileRepositoryImpl struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) repositories.UserProfileRepository {
	return &UserProfileRepositoryImpl{db: db}
}

func (r *UserProfileRepositoryImpl) Create(ctx context.Context, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *UserProfileRepositoryImpl) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserProfile, error) {
	var profile models.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *UserProfileRepositoryImpl) Update(ctx context.Context, userID uuid.UUID, profile *models.UserProfile) error {
	return r.db.WithContext(ctx).
		Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(profile).Error
}

func (r *UserProfileRepositoryImpl) UpdateKarma(ctx context.Context, userID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("karma", gorm.Expr("karma + ?", delta)).Error
}

func (r *UserProfileRepositoryImpl) UpdateFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("followers_count", gorm.Expr("followers_count + ?", delta)).Error
}

func (r *UserProfileRepositoryImpl) UpdateFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error {
	return r.db.WithContext(ctx).
		Model(&models.UserProfile{}).
		Where("user_id = ?", userID).
		UpdateColumn("following_count", gorm.Expr("following_count + ?", delta)).Error
}

func (r *UserProfileRepositoryImpl) GetTopUsersByKarma(ctx context.Context, limit int) ([]*models.UserProfile, error) {
	var profiles []*models.UserProfile
	err := r.db.WithContext(ctx).
		Preload("User"). // Load users_identity data (no need for .Profile here, it's circular)
		Order("karma DESC").
		Limit(limit).
		Find(&profiles).Error
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *UserProfileRepositoryImpl) Delete(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.UserProfile{}).Error
}
