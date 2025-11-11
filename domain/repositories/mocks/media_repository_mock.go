package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/models"
)

// MockMediaRepository is a mock implementation of MediaRepository
type MockMediaRepository struct {
	mock.Mock
}

func (m *MockMediaRepository) Create(ctx context.Context, media *models.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Media, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Media, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaRepository) GetByVideoID(ctx context.Context, videoID string) (*models.Media, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Media), args.Error(1)
}

func (m *MockMediaRepository) ListByUser(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Media, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaRepository) ListByType(ctx context.Context, userID uuid.UUID, mediaType string, offset, limit int) ([]*models.Media, error) {
	args := m.Called(ctx, userID, mediaType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}

func (m *MockMediaRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMediaRepository) Update(ctx context.Context, media *models.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) IncrementUsageCount(ctx context.Context, mediaID uuid.UUID) error {
	args := m.Called(ctx, mediaID)
	return args.Error(0)
}

func (m *MockMediaRepository) DecrementUsageCount(ctx context.Context, mediaID uuid.UUID) error {
	args := m.Called(ctx, mediaID)
	return args.Error(0)
}

func (m *MockMediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMediaRepository) GetUnusedMedia(ctx context.Context, olderThan int) ([]*models.Media, error) {
	args := m.Called(ctx, olderThan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Media), args.Error(1)
}
