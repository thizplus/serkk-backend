package models

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	FollowerID uuid.UUID      `gorm:"primaryKey;index"`
	Follower   UsersIdentity `gorm:"foreignKey:FollowerID"`

	FollowingID uuid.UUID      `gorm:"primaryKey;index"`
	Following   UsersIdentity `gorm:"foreignKey:FollowingID"`

	CreatedAt time.Time `gorm:"index"`
}

func (Follow) TableName() string {
	return "follows"
}
