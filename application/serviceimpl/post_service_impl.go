package serviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/redis"
	"gofiber-template/infrastructure/websocket"
	"gofiber-template/pkg/utils"
)

type PostServiceImpl struct {
	postRepo        repositories.PostRepository
	userRepo        repositories.UserRepository
	voteRepo        repositories.VoteRepository
	savedPostRepo   repositories.SavedPostRepository
	tagService      services.TagService
	mediaRepo       repositories.MediaRepository
	notificationHub *websocket.NotificationHub
	redisService    *redis.RedisService
	feedCache       *redis.FeedCacheService
}

func NewPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	voteRepo repositories.VoteRepository,
	savedPostRepo repositories.SavedPostRepository,
	tagService services.TagService,
	mediaRepo repositories.MediaRepository,
	notificationHub *websocket.NotificationHub,
	redisService *redis.RedisService,
	feedCache *redis.FeedCacheService,
) services.PostService {
	return &PostServiceImpl{
		postRepo:        postRepo,
		userRepo:        userRepo,
		voteRepo:        voteRepo,
		savedPostRepo:   savedPostRepo,
		tagService:      tagService,
		mediaRepo:       mediaRepo,
		notificationHub: notificationHub,
		redisService:    redisService,
		feedCache:       feedCache,
	}
}

func (s *PostServiceImpl) CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error) {
	// ============================================
	// STEP 1: Generate or use provided idempotency keys
	// ============================================
	var clientPostID string
	var idempotencyKey string

	if req.ClientPostID != nil && *req.ClientPostID != "" {
		clientPostID = *req.ClientPostID
	} else {
		// Generate backend client post ID if not provided
		clientPostID = utils.GenerateClientPostID()
		log.Printf("[IDEMPOTENCY] Generated backend clientPostID: %s", clientPostID)
	}

	if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
		idempotencyKey = *req.IdempotencyKey
	} else {
		// Generate backend idempotency key if not provided
		idempotencyKey = utils.GenerateIdempotencyKey()
		log.Printf("[IDEMPOTENCY] Generated backend idempotencyKey: %s", idempotencyKey)
	}

	// ============================================
	// STEP 2: Check idempotency cache (Redis)
	// ============================================
	if s.redisService != nil {
		cachedResponse, err := s.redisService.GetIdempotencyCache(ctx, idempotencyKey)
		if err == nil && cachedResponse != nil {
			// Cache hit - unmarshal and return cached response
			var response dto.PostResponse
			if err := json.Unmarshal(cachedResponse, &response); err == nil {
				log.Printf("[IDEMPOTENCY] Cache hit for idempotencyKey: %s", idempotencyKey)
				return &response, nil
			}
		}
	}

	// ============================================
	// STEP 3: Check database for existing post with clientPostID
	// ============================================
	existingPost, err := s.postRepo.GetByClientPostID(ctx, clientPostID)
	if err == nil && existingPost != nil {
		// Post already exists - return existing post
		log.Printf("[IDEMPOTENCY] Post already exists with clientPostID: %s (postID: %s)", clientPostID, existingPost.ID)

		response, err := s.GetPost(ctx, existingPost.ID, &userID)
		if err != nil {
			return nil, err
		}

		// Cache the response for future requests
		if s.redisService != nil {
			_ = s.redisService.SetIdempotencyCache(ctx, idempotencyKey, response, 24*time.Hour)
		}

		return response, nil
	}

	// ============================================
	// STEP 4: Validate media ownership (if provided)
	// ============================================
	if len(req.MediaIDs) > 0 {
		for _, mediaID := range req.MediaIDs {
			media, err := s.mediaRepo.GetByID(ctx, mediaID)
			if err != nil {
				return nil, errors.New("some media files not found")
			}
			if media.UserID != userID {
				return nil, errors.New("some media files not owned by you")
			}
		}
	}

	// ============================================
	// STEP 5: Create new post
	// ============================================

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

	// Determine post type based on media
	postType, err := s.determinePostType(ctx, req.MediaIDs)
	if err != nil {
		return nil, err
	}

	// Create post
	post := &models.Post{
		ID:           uuid.New(),
		Title:        req.Title,
		Content:      req.Content,
		AuthorID:     userID,
		Votes:        0,
		CommentCount: 0,
		Type:         postType,
		Status:       status,
		ClientPostID: &clientPostID, // ✅ Set clientPostID
		IsDeleted:    false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Handle source post for crosspost
	if req.SourcePostID != nil {
		post.SourcePostID = req.SourcePostID
	}

	// ============================================
	// STEP 6: Create post in database with race condition handling
	// ============================================
	err = s.postRepo.Create(ctx, post)
	if err != nil {
		// Check if it's a duplicate key error (race condition)
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			log.Printf("[RACE CONDITION] Duplicate clientPostID detected: %s", clientPostID)

			// Try to get the existing post
			existingPost, getErr := s.postRepo.GetByClientPostID(ctx, clientPostID)
			if getErr == nil && existingPost != nil {
				response, err := s.GetPost(ctx, existingPost.ID, &userID)
				if err != nil {
					return nil, err
				}

				// Cache the response
				if s.redisService != nil {
					_ = s.redisService.SetIdempotencyCache(ctx, idempotencyKey, response, 24*time.Hour)
				}

				return response, nil
			}
		}
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

	// ============================================
	// STEP 7: Get full post with relations
	// ============================================
	response, err := s.GetPost(ctx, post.ID, &userID)
	if err != nil {
		return nil, err
	}

	// ============================================
	// STEP 8: Cache the response (24 hours TTL)
	// ============================================
	if s.redisService != nil {
		if err := s.redisService.SetIdempotencyCache(ctx, idempotencyKey, response, 24*time.Hour); err != nil {
			log.Printf("[IDEMPOTENCY] Failed to cache response: %v", err)
		} else {
			log.Printf("[IDEMPOTENCY] Cached response for idempotencyKey: %s", idempotencyKey)
		}
	}

	// ============================================
	// STEP 9: Invalidate feed caches (new post created)
	// ============================================
	if s.feedCache != nil {
		if err := s.feedCache.InvalidateAllFeeds(ctx); err != nil {
			log.Printf("[CACHE] Failed to invalidate feed caches: %v", err)
		} else {
			log.Printf("[CACHE] Feed caches invalidated after post creation")
		}
	}

	return response, nil
}

// hasProcessingVideo checks if any of the media IDs are videos with processing status
func (s *PostServiceImpl) hasProcessingVideo(ctx context.Context, mediaIDs []uuid.UUID) (bool, error) {
	// NOTE: We no longer encode videos (R2 direct play)
	// All videos are ready immediately after upload
	// This function always returns false now (no videos are processing)
	return false, nil
}

// determinePostType determines the post type based on attached media
func (s *PostServiceImpl) determinePostType(ctx context.Context, mediaIDs []uuid.UUID) (string, error) {
	// No media = text post
	if len(mediaIDs) == 0 {
		return "text", nil
	}

	// Fetch media to check their types
	hasVideo := false
	imageCount := 0

	for _, mediaID := range mediaIDs {
		media, err := s.mediaRepo.GetByID(ctx, mediaID)
		if err != nil {
			return "", err
		}

		if media.Type == "video" {
			hasVideo = true
		} else if media.Type == "image" {
			imageCount++
		}
	}

	// Video takes priority
	if hasVideo {
		return "video", nil
	}

	// Multiple images = gallery
	if imageCount > 1 {
		return "gallery", nil
	}

	// Single image
	if imageCount == 1 {
		return "image", nil
	}

	// Default to text if no recognized media type
	return "text", nil
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

	// Invalidate feed caches (post updated)
	if s.feedCache != nil {
		if err := s.feedCache.InvalidateAllFeeds(ctx); err != nil {
			log.Printf("[CACHE] Failed to invalidate feed caches: %v", err)
		} else {
			log.Printf("[CACHE] Feed caches invalidated after post update")
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
	err = s.postRepo.Delete(ctx, postID)
	if err != nil {
		return err
	}

	// Invalidate feed caches (post deleted)
	if s.feedCache != nil {
		if err := s.feedCache.InvalidateAllFeeds(ctx); err != nil {
			log.Printf("[CACHE] Failed to invalidate feed caches: %v", err)
		} else {
			log.Printf("[CACHE] Feed caches invalidated after post deletion")
		}
	}

	return nil
}

func (s *PostServiceImpl) ListPosts(ctx context.Context, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error) {
	// STEP 1: Build cache key
	var cacheKey string
	if s.feedCache != nil {
		page := offset / limit
		cacheKey = s.feedCache.BuildFeedCacheKey(sortBy, page, limit)
	}

	// STEP 2: Try to get from cache (skip if userID present - personalized data)
	if userID == nil && s.feedCache != nil {
		cachedPosts, err := s.feedCache.GetCachedFeed(ctx, cacheKey)
		if err == nil && cachedPosts != nil {
			// Cache hit! Return cached data
			log.Printf("[CACHE HIT] Returning %d cached posts from key: %s", len(cachedPosts), cacheKey)
			count, _ := s.postRepo.Count(ctx)
			return &dto.PostListResponse{
				Posts: cachedPosts,
				Meta: dto.PaginationMeta{
					Total:  &count,
					Offset: offset,
					Limit:  limit,
				},
			}, nil
		}
	}

	// STEP 3: Cache miss - query database
	posts, err := s.postRepo.List(ctx, offset, limit, sortBy)
	if err != nil {
		return nil, err
	}

	count, err := s.postRepo.Count(ctx)
	if err != nil {
		return nil, err
	}

	response, err := s.buildPostListResponse(ctx, posts, count, offset, limit, userID)
	if err != nil {
		return nil, err
	}

	// STEP 4: Cache the result (skip personalized data)
	if userID == nil && s.feedCache != nil && response != nil {
		ttl := s.feedCache.GetFeedTTL(sortBy)
		if err := s.feedCache.CacheFeed(ctx, cacheKey, response.Posts, ttl); err != nil {
			log.Printf("[CACHE] Failed to cache feed: %v", err)
		}
	}

	return response, nil
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
	// Fetch limit+1 to determine if there are more results
	posts, err := s.postRepo.ListByTag(ctx, tagName, offset, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	// Determine hasMore by checking if we got more than requested limit
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Trim to actual limit
	}

	return s.buildPostListResponseWithHasMore(ctx, posts, hasMore, offset, limit, userID)
}

func (s *PostServiceImpl) ListPostsByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error) {
	// Fetch limit+1 to determine if there are more results
	posts, err := s.postRepo.ListByTagID(ctx, tagID, offset, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	// Determine hasMore by checking if we got more than requested limit
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Trim to actual limit
	}

	return s.buildPostListResponseWithHasMore(ctx, posts, hasMore, offset, limit, userID)
}

func (s *PostServiceImpl) SearchPosts(ctx context.Context, query string, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
	// Fetch limit+1 to determine if there are more results
	posts, err := s.postRepo.Search(ctx, query, offset, limit+1)
	if err != nil {
		return nil, err
	}

	// Determine hasMore by checking if we got more than requested limit
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Trim to actual limit
	}

	return s.buildPostListResponseWithHasMore(ctx, posts, hasMore, offset, limit, userID)
}

// SearchPostsWithCursor searches posts with cursor-based pagination
func (s *PostServiceImpl) SearchPostsWithCursor(ctx context.Context, query string, cursorStr string, limit int, userID *uuid.UUID) (*dto.PostListCursorResponse, error) {
	// Decode cursor
	cursor, err := utils.DecodePostCursor(cursorStr)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to determine if there are more pages
	posts, err := s.postRepo.SearchWithCursor(ctx, query, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	// Search always sorts by "new" (created_at DESC)
	return s.buildPostListCursorResponse(ctx, posts, limit, repositories.SortByNew, userID)
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
	// Fetch limit+1 to determine if there are more results
	posts, err := s.postRepo.GetCrossposts(ctx, postID, offset, limit+1)
	if err != nil {
		return nil, err
	}

	// Determine hasMore by checking if we got more than requested limit
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Trim to actual limit
	}

	return s.buildPostListResponseWithHasMore(ctx, posts, hasMore, offset, limit, userID)
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

// ListPostsWithCursor returns posts with cursor-based pagination
func (s *PostServiceImpl) ListPostsWithCursor(ctx context.Context, cursorStr string, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListCursorResponse, error) {
	// Decode cursor
	cursor, err := utils.DecodePostCursor(cursorStr)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to determine if there are more pages
	posts, err := s.postRepo.ListWithCursor(ctx, cursor, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	return s.buildPostListCursorResponse(ctx, posts, limit, sortBy, userID)
}

// ListPostsByAuthorWithCursor returns posts by author with cursor pagination
func (s *PostServiceImpl) ListPostsByAuthorWithCursor(ctx context.Context, authorID uuid.UUID, cursorStr string, limit int, userID *uuid.UUID) (*dto.PostListCursorResponse, error) {
	// Decode cursor
	cursor, err := utils.DecodePostCursor(cursorStr)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to determine if there are more pages
	posts, err := s.postRepo.ListByAuthorWithCursor(ctx, authorID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	return s.buildPostListCursorResponse(ctx, posts, limit, repositories.SortByNew, userID)
}

// ListPostsByTagWithCursor returns posts by tag with cursor pagination
func (s *PostServiceImpl) ListPostsByTagWithCursor(ctx context.Context, tagName string, cursorStr string, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListCursorResponse, error) {
	// Decode cursor
	cursor, err := utils.DecodePostCursor(cursorStr)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to determine if there are more pages
	posts, err := s.postRepo.ListByTagWithCursor(ctx, tagName, cursor, limit+1, sortBy)
	if err != nil {
		return nil, err
	}

	return s.buildPostListCursorResponse(ctx, posts, limit, sortBy, userID)
}

// GetFollowingFeedWithCursor returns posts from followed users with cursor pagination
func (s *PostServiceImpl) GetFollowingFeedWithCursor(ctx context.Context, userID uuid.UUID, cursorStr string, limit int) (*dto.PostFeedCursorResponse, error) {
	// Decode cursor
	cursor, err := utils.DecodePostCursor(cursorStr)
	if err != nil {
		return nil, errors.New("invalid cursor")
	}

	// Fetch limit+1 to determine if there are more pages
	posts, err := s.postRepo.ListFollowingFeedWithCursor(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, err
	}

	listResp, err := s.buildPostListCursorResponse(ctx, posts, limit, repositories.SortByNew, &userID)
	if err != nil {
		return nil, err
	}

	return &dto.PostFeedCursorResponse{
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
		_ = freshMediaList     // Use variable to avoid unused error

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
			Total:  &count,
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

// Helper function to build post list response with hasMore (for search/filtered lists)
func (s *PostServiceImpl) buildPostListResponseWithHasMore(ctx context.Context, posts []*models.Post, hasMore bool, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error) {
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
			HasMore: &hasMore,
			Offset:  offset,
			Limit:   limit,
		},
	}, nil
}

// Helper function to build cursor-based post list response
func (s *PostServiceImpl) buildPostListCursorResponse(ctx context.Context, posts []*models.Post, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListCursorResponse, error) {
	// Determine if there are more pages (limit+1 pattern)
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit] // Trim to actual limit
	}

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
	responses := make([]dto.PostResponse, len(posts))
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

	// Generate next cursor from last item if there are more pages
	var nextCursor *string
	if hasMore && len(posts) > 0 {
		lastPost := posts[len(posts)-1]

		// Determine sort value based on sort type
		var sortValue *float64
		switch sortBy {
		case repositories.SortByTop:
			// For top sorting, use votes
			votes := float64(lastPost.Votes)
			sortValue = &votes
		case repositories.SortByHot:
			// For hot sorting, use hot score
			hoursSinceCreation := time.Since(lastPost.CreatedAt).Hours()
			hotScore := float64(lastPost.Votes) / math.Pow(hoursSinceCreation+2, 1.5)
			sortValue = &hotScore
		case repositories.SortByNew:
			// For new sorting, no sort value needed (only created_at and id)
			sortValue = nil
		default:
			sortValue = nil
		}

		// Encode cursor
		encoded, err := utils.EncodePostCursor(sortValue, lastPost.CreatedAt, lastPost.ID)
		if err != nil {
			return nil, err
		}
		nextCursor = &encoded
	}

	return &dto.PostListCursorResponse{
		Posts: responses,
		Meta: dto.CursorPaginationMeta{
			NextCursor: nextCursor,
			HasMore:    hasMore,
			Limit:      limit,
		},
	}, nil
}

var _ services.PostService = (*PostServiceImpl)(nil)
