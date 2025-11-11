package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/models"
)

// MockFollowRepository is a mock implementation of FollowRepository
type MockFollowRepository struct {
	mock.Mock
}

func (m *MockFollowRepository) Follow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	args := m.Called(ctx, followerID, followingID)
	return args.Error(0)
}

func (m *MockFollowRepository) Unfollow(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) error {
	args := m.Called(ctx, followerID, followingID)
	return args.Error(0)
}

func (m *MockFollowRepository) IsFollowing(ctx context.Context, followerID uuid.UUID, followingID uuid.UUID) (bool, error) {
	args := m.Called(ctx, followerID, followingID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFollowRepository) GetFollowers(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockFollowRepository) CountFollowers(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFollowRepository) GetFollowing(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockFollowRepository) CountFollowing(ctx context.Context, userID uuid.UUID) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFollowRepository) GetFollowStatus(ctx context.Context, followerID uuid.UUID, userIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	args := m.Called(ctx, followerID, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]bool), args.Error(1)
}

func (m *MockFollowRepository) GetMutualFollows(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*models.User, error) {
	args := m.Called(ctx, userID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockFollowRepository) UpdateFollowerCount(ctx context.Context, userID uuid.UUID, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}

func (m *MockFollowRepository) UpdateFollowingCount(ctx context.Context, userID uuid.UUID, delta int) error {
	args := m.Called(ctx, userID, delta)
	return args.Error(0)
}
