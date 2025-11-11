package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/dto"
)

// MockTagService is a mock implementation of TagService
type MockTagService struct {
	mock.Mock
}

func (m *MockTagService) GetTag(ctx context.Context, tagID uuid.UUID) (*dto.TagResponse, error) {
	args := m.Called(ctx, tagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TagResponse), args.Error(1)
}

func (m *MockTagService) GetTagByName(ctx context.Context, name string) (*dto.TagResponse, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TagResponse), args.Error(1)
}

func (m *MockTagService) ListTags(ctx context.Context, offset, limit int) (*dto.TagListResponse, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TagListResponse), args.Error(1)
}

func (m *MockTagService) GetPopularTags(ctx context.Context, limit int) (*dto.PopularTagsResponse, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PopularTagsResponse), args.Error(1)
}

func (m *MockTagService) SearchTags(ctx context.Context, query string, limit int) (*dto.TagListResponse, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TagListResponse), args.Error(1)
}

func (m *MockTagService) GetOrCreateTags(ctx context.Context, tagNames []string) ([]uuid.UUID, error) {
	args := m.Called(ctx, tagNames)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uuid.UUID), args.Error(1)
}
