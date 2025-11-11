package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	FileName  string    `gorm:"not null"`
	FileSize  int64
	MimeType  string
	URL       string `gorm:"not null"`
	CDNPath   string
	UserID    uuid.UUID `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (File) TableName() string {
	return "files"
}

// BeforeCreate hook to generate UUID before creating File
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
