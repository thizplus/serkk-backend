package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserProfile contains social-specific data for users
// Separated from UsersIdentity (auth data) for clean separation of concerns
// Now includes displayName and avatar (moved from Auth Service for simpler profile updates)
type UserProfile struct {
	// Foreign Key to UsersIdentity
	UserID uuid.UUID `gorm:"primaryKey;type:uuid" json:"userId"`

	// Profile Info (moved from Auth Service)
	DisplayName string `gorm:"size:100;not null;default:''" json:"displayName"`
	Avatar      string `gorm:"size:500;default:''" json:"avatar"`

	// Additional Profile Info
	Bio      string `gorm:"type:text" json:"bio"`
	Location string `gorm:"size:100" json:"location"`
	Website  string `gorm:"size:255" json:"website"`

	// Social Stats
	Karma          int `gorm:"default:0;index" json:"karma"`
	FollowersCount int `gorm:"default:0" json:"followersCount"`
	FollowingCount int `gorm:"default:0" json:"followingCount"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Virtual relation to UsersIdentity (no DB FK due to separate databases)
	User *UsersIdentity `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (UserProfile) TableName() string {
	return "user_profiles"
}

// BeforeCreate hook
func (up *UserProfile) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	up.CreatedAt = now
	up.UpdatedAt = now
	return nil
}

// BeforeUpdate hook
func (up *UserProfile) BeforeUpdate(tx *gorm.DB) error {
	up.UpdatedAt = time.Now()
	return nil
}
