package services

import (
	"context"

	"github.com/google/uuid"

	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
)

// UsersIdentityService - Service สำหรับจัดการ users identity
type UsersIdentityService struct {
	repo repositories.UsersIdentityRepository
}

// NewUsersIdentityService creates a new users identity service
func NewUsersIdentityService(repo repositories.UsersIdentityRepository) *UsersIdentityService {
	return &UsersIdentityService{repo: repo}
}

// Upsert inserts or updates user identity
func (s *UsersIdentityService) Upsert(ctx context.Context, user *models.UsersIdentity) error {
	return s.repo.Upsert(ctx, user)
}

// GetByID retrieves user identity by ID
func (s *UsersIdentityService) GetByID(ctx context.Context, id uuid.UUID) (*models.UsersIdentity, error) {
	return s.repo.GetByID(ctx, id)
}

// GetByEmail retrieves user identity by email
func (s *UsersIdentityService) GetByEmail(ctx context.Context, email string) (*models.UsersIdentity, error) {
	return s.repo.GetByEmail(ctx, email)
}

// GetByUsername retrieves user identity by username
func (s *UsersIdentityService) GetByUsername(ctx context.Context, username string) (*models.UsersIdentity, error) {
	return s.repo.GetByUsername(ctx, username)
}

// SoftDelete performs soft delete on user identity
func (s *UsersIdentityService) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id)
}
