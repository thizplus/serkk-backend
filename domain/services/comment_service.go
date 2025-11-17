package services

import (
	"context"
	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/repositories"
)

type CommentService interface {
	// Create and manage comments
	CreateComment(ctx context.Context, userID uuid.UUID, req *dto.CreateCommentRequest) (*dto.CommentResponse, error)
	GetComment(ctx context.Context, commentID uuid.UUID, userID *uuid.UUID) (*dto.CommentResponse, error)
	UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, req *dto.UpdateCommentRequest) (*dto.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error

	// List comments (offset-based, deprecated)
	ListCommentsByPost(ctx context.Context, postID uuid.UUID, offset, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListResponse, error)
	ListCommentsByAuthor(ctx context.Context, authorID uuid.UUID, offset, limit int, userID *uuid.UUID) (*dto.CommentListResponse, error)
	ListReplies(ctx context.Context, parentID uuid.UUID, offset, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListResponse, error)

	// List comments with cursor (cursor-based pagination)
	ListCommentsByPostWithCursor(ctx context.Context, postID uuid.UUID, cursor string, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListCursorResponse, error)
	ListCommentsByAuthorWithCursor(ctx context.Context, authorID uuid.UUID, cursor string, limit int, userID *uuid.UUID) (*dto.CommentListCursorResponse, error)
	ListRepliesWithCursor(ctx context.Context, parentID uuid.UUID, cursor string, limit int, sortBy repositories.CommentSortBy, userID *uuid.UUID) (*dto.CommentListCursorResponse, error)

	// Tree structure
	GetCommentTree(ctx context.Context, postID uuid.UUID, maxDepth int, userID *uuid.UUID) (*dto.CommentTreeResponse, error)
	GetParentChain(ctx context.Context, commentID uuid.UUID, userID *uuid.UUID) ([]*dto.CommentResponse, error)
}
