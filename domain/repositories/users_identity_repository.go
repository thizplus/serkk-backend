package repositories

import (
	"context"

	"github.com/google/uuid"

	"gofiber-template/domain/models"
)

// UsersIdentityRepository - Repository สำหรับ users_identity table
type UsersIdentityRepository interface {
	// Upsert - Insert หรือ Update user identity
	Upsert(ctx context.Context, user *models.UsersIdentity) error

	// GetByID - ดึง user identity ตาม ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.UsersIdentity, error)

	// GetByEmail - ดึง user identity ตาม email
	GetByEmail(ctx context.Context, email string) (*models.UsersIdentity, error)

	// GetByUsername - ดึง user identity ตาม username
	GetByUsername(ctx context.Context, username string) (*models.UsersIdentity, error)

	// SoftDelete - Soft delete user identity
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
