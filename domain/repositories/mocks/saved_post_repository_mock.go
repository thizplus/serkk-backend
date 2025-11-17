package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/models"
	"gofiber-template/pkg/utils"
)

// MockSavedPostRepository is a mock implementation of SavedPostRepository
type MockSavedPostRepository struct {
	mock.Mock
}

func (m *MockSavedPostRepository) SavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockSavedPostRepository) UnsavePost(ctx context.Context, userID uuid.UUID, postID uuid.UUID) error {
	args := m.Called(ctx, userID, postID)
	return args.Error(0)
}

func (m *MockSavedPostRepository) IsSaved(ctx context.Context, userID uuid.UUID, postID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, postID)
	return args.Bool(0), args.Error(1)
}

func (m *MockSavedPostRepository) GetSavedPosts(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockSavedPostRepository) CountSavedPosts(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSavedPostRepository) GetSavedStatus(ctx context.Context, userID uuid.UUID, postIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	args := m.Called(ctx, userID, postIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]bool), args.Error(1)
}

func (m *MockSavedPostRepository) GetSavedPostsWithCursor(ctx context.Context, userID uuid.UUID, cursor *utils.PostCursor, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, userID, cursor, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}
