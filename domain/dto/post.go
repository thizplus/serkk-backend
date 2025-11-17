package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreatePostRequest - Request for creating a new post
type CreatePostRequest struct {
	ClientPostID   *string     `json:"clientPostId" validate:"omitempty,min=1,max=255"`   // Client-generated unique ID for idempotency
	IdempotencyKey *string     `json:"idempotencyKey" validate:"omitempty,min=1,max=255"` // Idempotency key for caching responses
	Title          string      `json:"title" validate:"required,min=1,max=300"`
	Content        string      `json:"content" validate:"required,min=1,max=40000"`
	MediaIDs       []uuid.UUID `json:"mediaIds" validate:"omitempty,max=10,dive,uuid"` // Max 10 media files per post
	Tags           []string    `json:"tags" validate:"omitempty,max=5,dive,min=1,max=50"`
	SourcePostID   *uuid.UUID  `json:"sourcePostId" validate:"omitempty,uuid"` // For crossposting
	IsDraft        bool        `json:"isDraft"`                                // true = save as draft (for video encoding)
}

// UpdatePostRequest - Request for updating a post
type UpdatePostRequest struct {
	Title   string   `json:"title" validate:"omitempty,min=1,max=300"`
	Content string   `json:"content" validate:"omitempty,min=1,max=40000"`
	Tags    []string `json:"tags" validate:"omitempty,max=5,dive,min=1,max=50"`
}

// ListPostsRequest - Request for listing posts with cursor pagination
type ListPostsRequest struct {
	Sort   string `query:"sort" validate:"omitempty,oneof=new top hot"`      // Sort order: new, top, hot
	Tag    string `query:"tag" validate:"omitempty,min=1,max=50"`            // Filter by tag
	Cursor string `query:"cursor" validate:"omitempty"`                      // Cursor for pagination
	Limit  int    `query:"limit" validate:"omitempty,min=1,max=100"`         // Items per page (default: 20)

	// Legacy offset-based params (for backward compatibility)
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// ListPostsByAuthorRequest - Request for listing posts by author
type ListPostsByAuthorRequest struct {
	AuthorID uuid.UUID `param:"authorId" validate:"required,uuid"`
	Cursor   string    `query:"cursor" validate:"omitempty"`
	Limit    int       `query:"limit" validate:"omitempty,min=1,max=100"`

	// Legacy offset-based params
	Offset int `query:"offset" validate:"omitempty,min=0"`
}

// PostResponse - Response for a single post
type PostResponse struct {
	ID           uuid.UUID       `json:"id"`
	Title        string          `json:"title"`
	Content      string          `json:"content"`
	Author       UserResponse    `json:"author"`
	Votes        int             `json:"votes"`
	CommentCount int             `json:"commentCount"`
	Type         string          `json:"type"`                 // "text", "image", "gallery", "video"
	Media        []MediaResponse `json:"media,omitempty"`
	Tags         []TagResponse   `json:"tags,omitempty"`
	SourcePost   *PostResponse   `json:"sourcePost,omitempty"` // For crossposts
	Status       string          `json:"status"`               // "draft" or "published"
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`

	// User-specific fields (when authenticated)
	UserVote *string  `json:"userVote,omitempty"` // "up", "down", or null
	IsSaved  *bool    `json:"isSaved,omitempty"`  // true/false
	HotScore *float64 `json:"hotScore,omitempty"` // For debugging/sorting
}

// PostListResponse - Response for listing posts (offset-based, deprecated)
// Use PostListCursorResponse for new implementations
type PostListResponse struct {
	Posts []PostResponse `json:"posts"`
	Meta  PaginationMeta `json:"meta"`
}

// PostListCursorResponse - Response for listing posts with cursor pagination
type PostListCursorResponse struct {
	Posts []PostResponse       `json:"posts"`
	Meta  CursorPaginationMeta `json:"meta"`
}

// PostFeedResponse - Response for feed with mixed content types (offset-based, deprecated)
// Use PostFeedCursorResponse for new implementations
type PostFeedResponse struct {
	Posts []PostResponse `json:"posts"`
	Meta  PaginationMeta `json:"meta"`
}

// PostFeedCursorResponse - Response for feed with cursor pagination
type PostFeedCursorResponse struct {
	Posts []PostResponse       `json:"posts"`
	Meta  CursorPaginationMeta `json:"meta"`
}

// PostSummaryResponse - Lightweight post info for nested responses (comments, etc.)
type PostSummaryResponse struct {
	ID        uuid.UUID    `json:"id"`
	Title     string       `json:"title"`
	Author    UserResponse `json:"author"`
	CreatedAt time.Time    `json:"createdAt"`
}
