package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	Title       string    `gorm:"not null"`
	Description string
	Status      string `gorm:"default:'pending'"`
	Priority    int    `gorm:"default:1"`
	DueDate     *time.Time
	UserID      uuid.UUID `gorm:"not null"`
	User        User      `gorm:"foreignKey:UserID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Task) TableName() string {
	return "tasks"
}

// BeforeCreate hook to generate UUID before creating Task
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
