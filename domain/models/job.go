package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Job struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"not null"`
	CronExpr  string    `gorm:"not null"`
	Payload   string    `gorm:"type:jsonb"`
	Status    string    `gorm:"default:'active'"`
	LastRun   *time.Time
	NextRun   *time.Time
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Job) TableName() string {
	return "jobs"
}

// BeforeCreate hook to generate UUID before creating Job
func (j *Job) BeforeCreate(tx *gorm.DB) error {
	if j.ID == uuid.Nil {
		j.ID = uuid.New()
	}
	return nil
}
