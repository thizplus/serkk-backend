package dto

import "github.com/google/uuid"

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data,omitempty"`
	Meta    PaginationMeta `json:"meta"`
	Error   string         `json:"error,omitempty"`
}

// PaginationMeta for offset-based pagination
// For accurate result counts, use Total. For estimated/partial results, use HasMore.
type PaginationMeta struct {
	Total   *int64 `json:"total,omitempty"`   // Exact total count (only if counted from DB)
	HasMore *bool  `json:"hasMore,omitempty"` // Whether more results exist (for search/filtered lists)
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
}

// CursorPaginationMeta - Metadata for cursor-based pagination
// Used for feeds, posts, comments, and other real-time content
type CursorPaginationMeta struct {
	NextCursor *string `json:"nextCursor,omitempty"` // Encoded cursor for next page
	HasMore    bool    `json:"hasMore"`              // Whether there are more items
	Limit      int     `json:"limit"`                // Items per page
}

type IDRequest struct {
	ID uuid.UUID `json:"id" validate:"required" param:"id"`
}

type BaseEntity struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
}
