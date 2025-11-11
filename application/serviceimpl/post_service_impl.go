package serviceimpl

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/websocket"
)

type PostServiceImpl struct {
	postRepo        repositories.PostRepository
	userRepo        repositories.UserRepository
	voteRepo        repositories.VoteRepository
	savedPostRepo   repositories.SavedPostRepository
	tagService      services.TagService
	mediaRepo       repositories.MediaRepository
	notificationHub *websocket.NotificationHub
}

func NewPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	voteRepo repositories.VoteRepository,
	savedPostRepo repositories.SavedPostRepository,
	tagService services.TagService,
	mediaRepo repositories.MediaRepository,
	notificationHub *websocket.NotificationHub,
) services.PostService {
	return &PostServiceImpl{
		postRepo:        postRepo,
		userRepo:        userRepo,
		voteRepo:        voteRepo,
		savedPostRepo:   savedPostRepo,
		tagService:      tagService,
		mediaRepo:       mediaRepo,
		notificationHub: notificationHub,
	}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error) {
	// Determine post status (draft or published)
	status := "published" // Default

	// Check if user explicitly wants draft
	if req.IsDraft {
		status = "draft"
	} else if len(req.MediaIDs) > 0 {
		// Check if any attached video is still processing
		hasProcessingVideo, err := s.hasProcessingVideo(ctx, req.MediaIDs)
		if err != nil {
			return nil, err
		}
		if hasProcessingVideo {
			status = "draft" // Auto-draft if video is processing
		}
	}

	// Create post
	post := &models.Post{
		ID:           uuid.New(),
		Title:        req.Title,
		Content:      req.Content,
		AuthorID:     userID,
		Votes:        0,
		CommentCount: 0,
		Status:       status,
		IsDeleted:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Handle source post for crosspost
	if req.SourcePostID != nil {
		post.SourcePostID = req.SourcePostID
	}

	// Create post in database
	err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, err
	}

	// Handle tags
	if len(req.Tags) > 0 {
		tagIDs, err := s.tagService.GetOrCreateTags(ctx, req.Tags)
		if err != nil {
			return nil, err
		}
		err = s.postRepo.AttachTags(ctx, post.ID, tagIDs)
		if err != nil {
			return nil, err
		}
	}

	// Handle media attachments
	if len(req.MediaIDs) > 0 {
		err = s.postRepo.AttachMedia(ctx, post.ID, req.MediaIDs)
		if err != nil {
			return nil, err
		}
		// Increment usage count for each media
		for _, mediaID := range req.MediaIDs {
			_ = s.mediaRepo.IncrementUsageCount(ctx, mediaID)
		}
	}

	// Get full post with relations
	return s.GetPost(ctx, post.ID, &userID)
}

// hasProcessingVideo checks if any of the media IDs are videos with processing status
func (s *PostServiceImpl) hasProcessingVideo(ctx context.Context, mediaIDs []uuid.UUID) (bool, error) {
	// NOTE: We no longer encode videos (R2 direct play)
	// All videos are ready immediately after upload
	// This function always returns false now (no videos are processing)
	return false, nil
}

func (s *PostServiceImpl) GetPost(ctx context.Context, postID uuid.UUID, userID *uuid.UUID) (*dto.PostResponse, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	resp := dto.PostToPostResponse(post)

	// Add user-specific data if authenticated
	if userID != nil {
		// Get user's vote
		vote, _ := s.voteRepo.GetVote(ctx, *userID, postID, "post")
		if vote != nil {
			resp.UserVote = &vote.VoteType
		}

		// Check if saved
		isSaved, _ := s.savedPostRepo.IsSaved(ctx, *userID, postID)
		resp.IsSaved = &isSaved
	}

	return resp, nil
}

func (s *PostServiceImpl) UpdatePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID, req *dto.UpdatePostRequest) (*dto.PostResponse, error) {
	// Get existing post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	// Check ownership
	if post.AuthorID != userID {
		return nil, errors.New("unauthorized: not post owner")
	}

	// Update fields
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	post.UpdatedAt = time.Now()

	err = s.postRepo.Update(ctx, postID, post)
	if err != nil {
		return nil, err
	}

	// Update tags if provided
	if len(req.Tags) > 0 {
		tagIDs, err := s.tagService.GetOrCreateTags(ctx, req.Tags)
		if err != nil {
			return nil, err
		}
		err = s.postRepo.SyncTags(ctx, postID, tagIDs)
		if err != nil {
			return nil, err
		}
	}

	return s.GetPost(ctx, postID, &userID)
}

func (s *PostServiceImpl) DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error {
	// Get post
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	// Check ownership
	if post.AuthorID != userID {
		return errors.New("unauthorized: not post owner")
	}

	// Soft delete
	return s.postRepo.Delete(ctx, postID)
}

func (s *PostServiceImpl) ListPosts(ctx context.Context, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.List(ctx, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	count, err := s.postRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) ListPostsByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.ListByAuthor(ctx, authorID, offset, limit)
	if err != nil {
		return nil, err
	}

	count, err := s.postRepo.CountByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}

	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) ListPostsByTag(ctx context.Context, tagName string, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.ListByTag(ctx, tagName, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	// For tag-filtered lists, we'll use a simplified count
	count := int64(len(posts))
	if len(posts) == limit {
		count = int64(offset + limit + 1) // Approximate
	}

	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) ListPostsByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.ListByTagID(ctx, tagID, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	// For tag-filtered lists, we'll use a simplified count
	count := int64(len(posts))
	if len(posts) == limit {
		count = int64(offset + limit + 1) // Approximate
	}

	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) SearchPosts(ctx context.Context, query string, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, err
	}

	count := int64(len(posts))
	if len(posts) == limit {
		count = int64(offset + limit + 1)
	}

	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) CreateCrosspost(ctx context.Context, userID uuid.UUID, sourcePostID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error) {
	// Verify source post exists
	_, err := s.postRepo.GetByID(ctx, sourcePostID)
	if err != nil {
		return nil, errors.New("source post not found")
	}

	// Set source post ID
	req.SourcePostID = &sourcePostID

	return s.CreatePost(ctx, userID, req)
}

func (s *PostServiceImpl) GetCrossposts(ctx context.Context, postID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
	posts, err := s.postRepo.GetCrossposts(ctx, postID, offset, limit)
	if err != nil {
		return nil, err
	}

	count := int64(len(posts))
	return s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
}

func (s *PostServiceImpl) GetFeed(ctx context.Context, userID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy) (*dto.PostFeedResponse, error) {
	// For now, just return all posts sorted by the requested method
	// TODO: Implement personalized feed based on followed users
	posts, err := s.postRepo.List(ctx, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	count, err := s.postRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	listResp, err := s.buildPostListResponse(ctx, posts, count, offset, limit, &userID)
	if err != nil {
		return nil, err
	}

	return &dto.PostFeedResponse{
		Posts: listResp.Posts,
		Meta:  listResp.Meta,
	}, nil
}

// PublishDraftPostsWithMedia auto-publishes draft posts when all videos are ready
func (s *PostServiceImpl) PublishDraftPostsWithMedia(ctx context.Context, mediaID uuid.UUID) error {
	// Get all posts that contain this media
	posts, err := s.postRepo.GetPostsByMediaID(ctx, mediaID)
	if err != nil {
		return err
	}

	// Process each draft post
	for _, post := range posts {
		// Skip if not draft
		if post.Status != "draft" {
			continue
		}

		// ⭐ Reload media data to get latest encoding status
		// (Preloaded media might be stale)
		mediaIDs := make([]uuid.UUID, len(post.Media))
		for i, media := range post.Media {
			mediaIDs[i] = media.ID
		}

		// Get fresh media data from database
		freshMediaList := make([]*models.Media, 0, len(mediaIDs))
		for _, mediaID := range mediaIDs {
			freshMedia, err := s.mediaRepo.GetByID(ctx, mediaID)
			if err != nil {
				log.Printf("Warning: Failed to reload media %s: %v", mediaID, err)
				continue
			}
			freshMediaList = append(freshMediaList, freshMedia)
		}

		// NOTE: We no longer encode videos (R2 direct play)
		// All videos are ready immediately after upload
		// Check if all videos in this post are completed (using fresh data)
		allVideosReady := true // Always true now (no encoding)
		_ = freshMediaList      // Use variable to avoid unused error

		// Publish if all videos are ready (always true now)
		if allVideosReady {
			post.Status = "published"
			post.UpdatedAt = time.Now()
			err := s.postRepo.Update(ctx, post.ID, post)
			if err != nil {
				log.Printf("Failed to publish draft post %s: %v", post.ID, err)
				continue
			}
			log.Printf("✅ Auto-published draft post %s (all videos ready)", post.ID)

			// ⭐ Send WebSocket notification to post owner
			if s.notificationHub != nil {
				s.notificationHub.SendToUser(post.AuthorID, &websocket.NotificationMessage{
					Type: "post.auto_published",
					Payload: map[string]interface{}{
						"postId":      post.ID.String(),
						"status":      "published",
						"publishedAt": post.UpdatedAt.Format(time.RFC3339),
					},
				})
				log.Printf("📡 Sent WebSocket event 'post.auto_published' to user %s", post.AuthorID)
			}
		}
	}

	return nil
}

// Helper function to build post list response with user-specific data
func (s *PostServiceImpl) buildPostListResponse(ctx context.Context, posts []*models.Post, count int64, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
	responses := make([]dto.PostResponse, len(posts))

	// Collect all post IDs for batch operations
	postIDs := make([]uuid.UUID, len(posts))
	for i, post := range posts {
		postIDs[i] = post.ID
	}

	// Batch get user votes and saved status if authenticated
	var voteMap map[uuid.UUID]*models.Vote
	var savedMap map[uuid.UUID]bool
	if userID != nil {
		voteMap, _ = s.voteRepo.GetUserVotesForTargets(ctx, *userID, postIDs, "post")
		savedMap, _ = s.savedPostRepo.GetSavedStatus(ctx, *userID, postIDs)
	}

	// Build responses
	for i, post := range posts {
		resp := dto.PostToPostResponse(post)

		// Calculate hot score
		hoursSinceCreation := time.Since(post.CreatedAt).Hours()
		hotScore := float64(post.Votes) / math.Pow(hoursSinceCreation+2, 1.5)
		resp.HotScore = &hotScore

		// Add user-specific data
		if userID != nil {
			if vote, ok := voteMap[post.ID]; ok {
				resp.UserVote = &vote.VoteType
			}
			// Always set isSaved for authenticated users (true or false)
			isSaved := false
			if saved, ok := savedMap[post.ID]; ok {
				isSaved = saved
			}
			resp.IsSaved = &isSaved
		}

		responses[i] = *resp
	}

	return &dto.PostListResponse{
		Posts: responses,
		Meta: dto.PaginationMeta{
			Total:  count,
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

var _ services.PostService = (*PostServiceImpl)(nil)
