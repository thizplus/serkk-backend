package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
)

type usersIdentityRepositoryImpl struct {
	db *gorm.DB
}

// NewUsersIdentityRepository creates a new users identity repository
func NewUsersIdentityRepository(db *gorm.DB) repositories.UsersIdentityRepository {
	return &usersIdentityRepositoryImpl{db: db}
}

// Upsert inserts or updates user identity
func (r *usersIdentityRepositoryImpl) Upsert(ctx context.Context, user *models.UsersIdentity) error {
	user.SyncedAt = time.Now()

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"email", "username", "synced_at", "updated_at"}),
		}).
		Create(user).Error
}

// GetByID retrieves user identity by ID
func (r *usersIdentityRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*models.UsersIdentity, error) {
	var user models.UsersIdentity
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error
	return &user, err
}

// GetByEmail retrieves user identity by email
func (r *usersIdentityRepositoryImpl) GetByEmail(ctx context.Context, email string) (*models.UsersIdentity, error) {
	var user models.UsersIdentity
	err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	return &user, err
}

// GetByUsername retrieves user identity by username
func (r *usersIdentityRepositoryImpl) GetByUsername(ctx context.Context, username string) (*models.UsersIdentity, error) {
	var user models.UsersIdentity
	err := r.db.WithContext(ctx).
		Where("username = ? AND deleted_at IS NULL", username).
		First(&user).Error
	return &user, err
}

// SoftDelete performs soft delete on user identity
func (r *usersIdentityRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.UsersIdentity{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}
