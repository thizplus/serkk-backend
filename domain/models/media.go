package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Media struct {
	ID     uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID uuid.UUID `gorm:"not null;index"`
	User   User      `gorm:"foreignKey:UserID"`

	// File info
	Type      string `gorm:"not null;index"` // image, video, file
	FileName  string `gorm:"not null"`
	Extension string `gorm:"type:varchar(10);index"` // pdf, doc, zip, etc. (without dot)
	MimeType  string `gorm:"not null"`
	Size      int64  `gorm:"not null"` // bytes

	// URLs (R2)
	URL       string `gorm:"not null"` // Full R2 public URL
	Thumbnail string // Thumbnail URL

	// Dimensions
	Width  int
	Height int

	// Video specific
	Duration float64 // seconds (for videos)

	// Polymorphic source tracking (for videos in different features)
	SourceType *string    `gorm:"type:varchar(50);index:idx_media_source"` // "post", "message", "reel", "comment", etc.
	SourceID   *uuid.UUID `gorm:"type:uuid;index:idx_media_source"`        // ID of the source entity

	// Usage tracking
	Posts      []Post `gorm:"many2many:post_media;"`
	UsageCount int    `gorm:"default:0"`

	// Timestamps
	CreatedAt time.Time `gorm:"index"`
}

func (Media) TableName() string {
	return "media"
}

// BeforeCreate hook to generate UUID before creating media
func (m *Media) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
