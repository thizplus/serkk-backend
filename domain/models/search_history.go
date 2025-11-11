package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SearchHistory struct {
	ID     uuid.UUID `gorm:"primaryKey;type:uuid"`
	UserID uuid.UUID `gorm:"not null;index"`
	User   User      `gorm:"foreignKey:UserID"`

	Query string `gorm:"not null"`
	Type  string // 'posts', 'users', 'all'

	SearchedAt time.Time `gorm:"index"`
}

func (SearchHistory) TableName() string {
	return "search_history"
}

// BeforeCreate hook to generate UUID before creating SearchHistory
func (sh *SearchHistory) BeforeCreate(tx *gorm.DB) error {
	if sh.ID == uuid.Nil {
		sh.ID = uuid.New()
	}
	return nil
}
