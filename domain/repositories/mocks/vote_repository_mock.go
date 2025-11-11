package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/models"
)

// MockVoteRepository is a mock implementation of VoteRepository
type MockVoteRepository struct {
	mock.Mock
}

func (m *MockVoteRepository) Vote(ctx context.Context, vote *models.Vote) error {
	args := m.Called(ctx, vote)
	return args.Error(0)
}

func (m *MockVoteRepository) Unvote(ctx context.Context, userID uuid.UUID, targetID uuid.UUID, targetType string) error {
	args := m.Called(ctx, userID, targetID, targetType)
	return args.Error(0)
}

func (m *MockVoteRepository) GetVote(ctx context.Context, userID uuid.UUID, targetID uuid.UUID, targetType string) (*models.Vote, error) {
	args := m.Called(ctx, userID, targetID, targetType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vote), args.Error(1)
}

func (m *MockVoteRepository) HasVoted(ctx context.Context, userID uuid.UUID, targetID uuid.UUID, targetType string) (bool, error) {
	args := m.Called(ctx, userID, targetID, targetType)
	return args.Bool(0), args.Error(1)
}

func (m *MockVoteRepository) GetVoteCount(ctx context.Context, targetID uuid.UUID, targetType string) (int64, int64, error) {
	args := m.Called(ctx, targetID, targetType)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockVoteRepository) ListByUser(ctx context.Context, userID uuid.UUID, targetType string, offset, limit int) ([]*models.Vote, error) {
	args := m.Called(ctx, userID, targetType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Vote), args.Error(1)
}

func (m *MockVoteRepository) ListByTarget(ctx context.Context, targetID uuid.UUID, targetType string, offset, limit int) ([]*models.Vote, error) {
	args := m.Called(ctx, targetID, targetType, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Vote), args.Error(1)
}

func (m *MockVoteRepository) GetUserVotesForTargets(ctx context.Context, userID uuid.UUID, targetIDs []uuid.UUID, targetType string) (map[uuid.UUID]*models.Vote, error) {
	args := m.Called(ctx, userID, targetIDs, targetType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID]*models.Vote), args.Error(1)
}
