package utils

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PostCursor represents a cursor for post pagination
// This cursor supports multiple sorting strategies (hot, new, top)
type PostCursor struct {
	// SortValue is the primary sort field (votes for top, hot_score for hot)
	// For "new" sorting, this is nil and we only use CreatedAt
	SortValue *float64 `json:"sort_value,omitempty"`

	// CreatedAt is used as a tie-breaker when SortValue is equal
	// For "new" sorting, this is the primary sort field
	CreatedAt time.Time `json:"created_at"`

	// ID is the final tie-breaker to ensure uniqueness
	// This is critical when multiple posts have the same SortValue and CreatedAt
	ID uuid.UUID `json:"id"`
}

// EncodePostCursor encodes a post cursor into a base64 URL-safe string
// This allows us to pass cursors in URL query parameters
func EncodePostCursor(sortValue *float64, createdAt time.Time, id uuid.UUID) (string, error) {
	cursor := PostCursor{
		SortValue: sortValue,
		CreatedAt: createdAt,
		ID:        id,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}

	// Encode to URL-safe base64
	encoded := base64.URLEncoding.EncodeToString(jsonBytes)
	return encoded, nil
}

// DecodePostCursor decodes a base64 cursor string back to PostCursor
// Returns nil if cursor string is empty (first page)
func DecodePostCursor(cursorStr string) (*PostCursor, error) {
	// Empty cursor means first page
	if cursorStr == "" {
		return nil, nil
	}

	// Decode from base64
	jsonBytes, err := base64.URLEncoding.DecodeString(cursorStr)
	if err != nil {
		return nil, err
	}

	// Unmarshal from JSON
	var cursor PostCursor
	if err := json.Unmarshal(jsonBytes, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// EncodePostCursorSimple creates a cursor for "new" sorting (no sort value needed)
// This is a convenience function for the most common case
func EncodePostCursorSimple(createdAt time.Time, id uuid.UUID) (string, error) {
	return EncodePostCursor(nil, createdAt, id)
}
