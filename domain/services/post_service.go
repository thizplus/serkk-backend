package services

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/repositories"
)

type PostService interface {
	// Create and manage posts
	CreatePost(ctx context.Context, userID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error)
	GetPost(ctx context.Context, postID uuid.UUID, userID *uuid.UUID) (*dto.PostResponse, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID, req *dto.UpdatePostRequest) (*dto.PostResponse, error)
	DeletePost(ctx context.Context, postID uuid.UUID, userID uuid.UUID) error

	// List and filter posts (offset-based, deprecated)
	ListPosts(ctx context.Context, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error)
	ListPostsByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error)
	ListPostsByTag(ctx context.Context, tagName string, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error)
	ListPostsByTagID(ctx context.Context, tagID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListResponse, error)

	// List with Cursor (cursor-based pagination)
	ListPostsWithCursor(ctx context.Context, cursor string, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListCursorResponse, error)
	ListPostsByAuthorWithCursor(ctx context.Context, authorID uuid.UUID, cursor string, limit int, userID *uuid.UUID) (*dto.PostListCursorResponse, error)
	ListPostsByTagWithCursor(ctx context.Context, tagName string, cursor string, limit int, sortBy repositories.PostSortBy, userID *uuid.UUID) (*dto.PostListCursorResponse, error)
	GetFollowingFeedWithCursor(ctx context.Context, userID uuid.UUID, cursor string, limit int) (*dto.PostFeedCursorResponse, error)

	// Search (offset-based, deprecated)
	SearchPosts(ctx context.Context, query string, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error)
	// Search with cursor (recommended)
	SearchPostsWithCursor(ctx context.Context, query string, cursor string, limit int, userID *uuid.UUID) (*dto.PostListCursorResponse, error)

	// Crossposting
	CreateCrosspost(ctx context.Context, userID uuid.UUID, sourcePostID uuid.UUID, req *dto.CreatePostRequest) (*dto.PostResponse, error)
	GetCrossposts(ctx context.Context, postID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.PostListResponse, error)

	// Feed
	GetFeed(ctx context.Context, userID uuid.UUID, offset, limit int, sortBy repositories.PostSortBy) (*dto.PostFeedResponse, error)

	// Draft posts management
	PublishDraftPostsWithMedia(ctx context.Context, mediaID uuid.UUID) error
}
