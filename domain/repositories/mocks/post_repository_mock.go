package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
)

// MockPostRepository is a mock implementation of PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) FindByIDWithRelations(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Post, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) GetByAuthor(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*models.Post, int64, error) {
	args := m.Called(ctx, authorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) GetByTag(ctx context.Context, tagID uuid.UUID, limit, offset int) ([]*models.Post, int64, error) {
	args := m.Called(ctx, tagID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) GetPopularPosts(ctx context.Context, duration time.Duration, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, duration, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) Search(ctx context.Context, query string, limit, offset int) ([]*models.Post, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Get(1).(int64), args.Error(2)
}

func (m *MockPostRepository) AttachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	args := m.Called(ctx, postID, tagIDs)
	return args.Error(0)
}

func (m *MockPostRepository) DetachTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	args := m.Called(ctx, postID, tagIDs)
	return args.Error(0)
}

func (m *MockPostRepository) GetTags(ctx context.Context, postID uuid.UUID) ([]*models.Tag, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tag), args.Error(1)
}

func (m *MockPostRepository) IncrementVotes(ctx context.Context, postID uuid.UUID, value int) error {
	args := m.Called(ctx, postID, value)
	return args.Error(0)
}

func (m *MockPostRepository) UpdatePostDTO(ctx context.Context, postID uuid.UUID, updateDTO dto.UpdatePostDTO) (*models.Post, error) {
	args := m.Called(ctx, postID, updateDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) GetFeed(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Post, int64, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Get(1).(int64), args.Error(2)
}
