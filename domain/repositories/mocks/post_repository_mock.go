package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
)

// MockPostRepository is a mock implementation of PostRepository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) Update(ctx context.Context, id uuid.UUID, post *models.Post) error {
	args := m.Called(ctx, id, post)
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

func (m *MockPostRepository) Search(ctx context.Context, query string, offset, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, query, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
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

func (m *MockPostRepository) UpdatePostDTO(ctx context.Context, postID uuid.UUID, req dto.UpdatePostRequest) (*models.Post, error) {
	args := m.Called(ctx, postID, req)
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

func (m *MockPostRepository) List(ctx context.Context, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	args := m.Called(ctx, offset, limit, sortBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) ListByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, authorID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) ListByTag(ctx context.Context, tagName string, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	args := m.Called(ctx, tagName, offset, limit, sortBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) ListByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy) ([]*models.Post, error) {
	args := m.Called(ctx, tagID, offset, limit, sortBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) GetCrossposts(ctx context.Context, postID uuid.UUID, offset, limit int) ([]*models.Post, error) {
	args := m.Called(ctx, postID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) CountByAuthor(ctx context.Context, authorID uuid.UUID) (int64, error) {
	args := m.Called(ctx, authorID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostRepository) IncrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostRepository) DecrementCommentCount(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostRepository) UpdateVoteCount(ctx context.Context, postID uuid.UUID, voteChange int) error {
	args := m.Called(ctx, postID, voteChange)
	return args.Error(0)
}

func (m *MockPostRepository) AttachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error {
	args := m.Called(ctx, postID, mediaIDs)
	return args.Error(0)
}

func (m *MockPostRepository) DetachMedia(ctx context.Context, postID uuid.UUID, mediaIDs []uuid.UUID) error {
	args := m.Called(ctx, postID, mediaIDs)
	return args.Error(0)
}

func (m *MockPostRepository) GetPostsByMediaID(ctx context.Context, mediaID uuid.UUID) ([]*models.Post, error) {
	args := m.Called(ctx, mediaID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostRepository) SyncTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	args := m.Called(ctx, postID, tagIDs)
	return args.Error(0)
}
