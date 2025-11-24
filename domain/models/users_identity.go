package models

import (
	"time"

	"github.com/google/uuid"
)

// UsersIdentity - Identity data จาก Auth Service (V2)
// เก็บเฉพาะ id, email, username (ไม่มี displayName, avatar, role)
type UsersIdentity struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Email     string     `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Username  string     `json:"username" gorm:"type:varchar(100);uniqueIndex;not null"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	SyncedAt  time.Time  `json:"synced_at"` // เวลาที่ sync จาก Auth Service event
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Relationship to UserProfile (contains displayName, avatar, karma, etc.)
	Profile *UserProfile `gorm:"foreignKey:UserID;references:ID" json:"profile,omitempty"`
}

// TableName specifies the table name
func (UsersIdentity) TableName() string {
	return "users_identity"
}
