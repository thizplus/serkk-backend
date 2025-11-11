package serviceimpl

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	repomocks "gofiber-template/domain/repositories/mocks"
	servicemocks "gofiber-template/domain/services/mocks"
	"gofiber-template/pkg/testutil"
)

func setupPostService() (
	*PostServiceImpl,
	*repomocks.MockPostRepository,
	*repomocks.MockUserRepository,
	*repomocks.MockVoteRepository,
	*repomocks.MockSavedPostRepository,
	*servicemocks.MockTagService,
	*repomocks.MockMediaRepository,
) {
	mockPostRepo := new(repomocks.MockPostRepository)
	mockUserRepo := new(repomocks.MockUserRepository)
	mockVoteRepo := new(repomocks.MockVoteRepository)
	mockSavedPostRepo := new(repomocks.MockSavedPostRepository)
	mockTagService := new(servicemocks.MockTagService)
	mockMediaRepo := new(repomocks.MockMediaRepository)

	service := &PostServiceImpl{
		postRepo:        mockPostRepo,
		userRepo:        mockUserRepo,
		voteRepo:        mockVoteRepo,
		savedPostRepo:   mockSavedPostRepo,
		tagService:      mockTagService,
		mediaRepo:       mockMediaRepo,
		notificationHub: nil,
	}

	return service, mockPostRepo, mockUserRepo, mockVoteRepo, mockSavedPostRepo, mockTagService, mockMediaRepo
}

func TestCreatePost_Success(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, mockVoteRepo, mockSavedPostRepo, _, _ := setupPostService()
	ctx := context.Background()
	userID := uuid.New()

	req := &dto.CreatePostRequest{
		Title:   "Test Post",
		Content: "Test content",
		IsDraft: false,
	}

	createdPost := testutil.CreateTestPostWithData(userID, req.Title, req.Content)

	mockPostRepo.On("Create", ctx, mock.AnythingOfType("*models.Post")).Return(nil)
	mockPostRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(createdPost, nil)
	mockVoteRepo.On("GetVote", ctx, userID, mock.AnythingOfType("uuid.UUID"), "post").Return(nil, errors.New("not found"))
	mockSavedPostRepo.On("IsSaved", ctx, userID, mock.AnythingOfType("uuid.UUID")).Return(false, nil)

	// Act
	result, err := service.CreatePost(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Title, result.Title)
	assert.Equal(t, req.Content, result.Content)
	mockPostRepo.AssertExpectations(t)
}

func TestCreatePost_WithTags(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, mockVoteRepo, mockSavedPostRepo, mockTagService, _ := setupPostService()
	ctx := context.Background()
	userID := uuid.New()

	tags := []string{"golang", "testing"}
	tagIDs := []uuid.UUID{uuid.New(), uuid.New()}

	req := &dto.CreatePostRequest{
		Title:   "Test Post with Tags",
		Content: "Test content",
		Tags:    tags,
	}

	createdPost := testutil.CreateTestPostWithData(userID, req.Title, req.Content)

	mockPostRepo.On("Create", ctx, mock.AnythingOfType("*models.Post")).Return(nil)
	mockTagService.On("GetOrCreateTags", ctx, tags).Return(tagIDs, nil)
	mockPostRepo.On("AttachTags", ctx, mock.AnythingOfType("uuid.UUID"), tagIDs).Return(nil)
	mockPostRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(createdPost, nil)
	mockVoteRepo.On("GetVote", ctx, userID, mock.AnythingOfType("uuid.UUID"), "post").Return(nil, errors.New("not found"))
	mockSavedPostRepo.On("IsSaved", ctx, userID, mock.AnythingOfType("uuid.UUID")).Return(false, nil)

	// Act
	result, err := service.CreatePost(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockPostRepo.AssertExpectations(t)
	mockTagService.AssertExpectations(t)
}

func TestCreatePost_WithMedia(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, mockVoteRepo, mockSavedPostRepo, _, mockMediaRepo := setupPostService()
	ctx := context.Background()
	userID := uuid.New()

	mediaIDs := []uuid.UUID{uuid.New(), uuid.New()}

	req := &dto.CreatePostRequest{
		Title:    "Test Post with Media",
		Content:  "Test content",
		MediaIDs: mediaIDs,
	}

	createdPost := testutil.CreateTestPostWithData(userID, req.Title, req.Content)

	mockPostRepo.On("Create", ctx, mock.AnythingOfType("*models.Post")).Return(nil)
	mockPostRepo.On("AttachMedia", ctx, mock.AnythingOfType("uuid.UUID"), mediaIDs).Return(nil)
	mockMediaRepo.On("IncrementUsageCount", ctx, mediaIDs[0]).Return(nil)
	mockMediaRepo.On("IncrementUsageCount", ctx, mediaIDs[1]).Return(nil)
	mockPostRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(createdPost, nil)
	mockVoteRepo.On("GetVote", ctx, userID, mock.AnythingOfType("uuid.UUID"), "post").Return(nil, errors.New("not found"))
	mockSavedPostRepo.On("IsSaved", ctx, userID, mock.AnythingOfType("uuid.UUID")).Return(false, nil)

	// Act
	result, err := service.CreatePost(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockPostRepo.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
}

func TestCreatePost_AsDraft(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, mockVoteRepo, mockSavedPostRepo, _, _ := setupPostService()
	ctx := context.Background()
	userID := uuid.New()

	req := &dto.CreatePostRequest{
		Title:   "Draft Post",
		Content: "Draft content",
		IsDraft: true,
	}

	createdPost := testutil.CreateTestPostWithData(userID, req.Title, req.Content)
	createdPost.Status = "draft"

	mockPostRepo.On("Create", ctx, mock.AnythingOfType("*models.Post")).Return(nil)
	mockPostRepo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(createdPost, nil)
	mockVoteRepo.On("GetVote", ctx, userID, mock.AnythingOfType("uuid.UUID"), "post").Return(nil, errors.New("not found"))
	mockSavedPostRepo.On("IsSaved", ctx, userID, mock.AnythingOfType("uuid.UUID")).Return(false, nil)

	// Act
	result, err := service.CreatePost(ctx, userID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "draft", result.Status)
	mockPostRepo.AssertExpectations(t)
}

func TestGetPost_Success(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, _, _, _, _ := setupPostService()
	ctx := context.Background()

	post := testutil.CreateTestPost(uuid.New())

	mockPostRepo.On("GetByID", ctx, post.ID).Return(post, nil)

	// Act
	result, err := service.GetPost(ctx, post.ID, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, post.ID, result.ID)
	assert.Equal(t, post.Title, result.Title)
	mockPostRepo.AssertExpectations(t)
}

func TestGetPost_WithUserContext(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, mockVoteRepo, mockSavedPostRepo, _, _ := setupPostService()
	ctx := context.Background()
	userID := uuid.New()

	post := testutil.CreateTestPost(uuid.New())
	vote := &models.Vote{
		UserID:     userID,
		TargetID:   post.ID,
		TargetType: "post",
		VoteType:   "up",
	}

	mockPostRepo.On("GetByID", ctx, post.ID).Return(post, nil)
	mockVoteRepo.On("GetVote", ctx, userID, post.ID, "post").Return(vote, nil)
	mockSavedPostRepo.On("IsSaved", ctx, userID, post.ID).Return(true, nil)

	// Act
	result, err := service.GetPost(ctx, post.ID, &userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.UserVote)
	assert.Equal(t, "up", *result.UserVote)
	assert.NotNil(t, result.IsSaved)
	assert.True(t, *result.IsSaved)
	mockPostRepo.AssertExpectations(t)
	mockVoteRepo.AssertExpectations(t)
	mockSavedPostRepo.AssertExpectations(t)
}

func TestGetPost_NotFound(t *testing.T) {
	// Arrange
	service, mockPostRepo, _, _, _, _, _ := setupPostService()
	ctx := context.Background()
	postID := uuid.New()

	mockPostRepo.On("GetByID", ctx, postID).Return(nil, errors.New("not found"))

	// Act
	result, err := service.GetPost(ctx, postID, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockPostRepo.AssertExpectations(t)
}
