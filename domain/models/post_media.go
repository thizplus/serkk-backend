package models

import (
	"time"

	"github.com/google/uuid"
)

// PostMedia represents the junction table between posts and media
// with support for ordering (display_order) to maintain media sequence
type PostMedia struct {
	PostID       uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	MediaID      uuid.UUID `gorm:"primaryKey;type:uuid;not null"`
	DisplayOrder int       `gorm:"not null;index;default:0"` // Order of media in post (0-indexed)
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`

	// Relationships (optional, for eager loading)
	Post  Post  `gorm:"foreignKey:PostID;references:ID"`
	Media Media `gorm:"foreignKey:MediaID;references:ID"`
}

func (PostMedia) TableName() string {
	return "post_media"
}
