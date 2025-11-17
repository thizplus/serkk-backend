package dto

import (
	"time"

	"github.com/google/uuid"
)

// SearchRequest - Request for searching
type SearchRequest struct {
	Query string `json:"query" validate:"required,min=1,max=255"`
	Type  string `json:"type" validate:"omitempty,oneof=post user tag all"` // Default: "all"
	Limit int    `json:"limit" validate:"omitempty,min=1,max=100"`
}

// SearchResponse - Response for search results (offset-based, deprecated)
type SearchResponse struct {
	Query string         `json:"query"`
	Type  string         `json:"type"`
	Posts []PostResponse `json:"posts,omitempty"`
	Users []UserResponse `json:"users,omitempty"`
	Tags  []TagResponse  `json:"tags,omitempty"`
	Meta  PaginationMeta `json:"meta"`
}

// SearchCursorResponse - Response for search with cursor pagination (posts only)
type SearchCursorResponse struct {
	Query string                `json:"query"`
	Posts []PostResponse        `json:"posts"`
	Meta  CursorPaginationMeta  `json:"meta"`
}

// SearchHistoryResponse - Response for search history
type SearchHistoryResponse struct {
	ID         uuid.UUID `json:"id"`
	Query      string    `json:"query"`
	Type       string    `json:"type,omitempty"`
	SearchedAt time.Time `json:"searchedAt"`
}

// SearchHistoryListResponse - Response for listing search history
type SearchHistoryListResponse struct {
	History []SearchHistoryResponse `json:"history"`
	Meta    PaginationMeta          `json:"meta"`
}

// PopularSearch - Single popular search item with count
type PopularSearch struct {
	Query string `json:"query"`
	Count int64  `json:"count"`
}

// PopularSearchesResponse - Response for popular searches
type PopularSearchesResponse struct {
	Searches []PopularSearch `json:"searches"`
}
