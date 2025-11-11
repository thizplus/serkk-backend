package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tag struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"uniqueIndex;not null;type:varchar(50)"`
	PostCount int       `gorm:"default:0;index"`

	Posts []Post `gorm:"many2many:post_tags;"`

	CreatedAt time.Time
}

func (Tag) TableName() string {
	return "tags"
}

// BeforeCreate hook to generate UUID before creating Tag
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
