package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AutoPostSetting stores configuration for AI-powered auto-posting
type AutoPostSetting struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid"`

	// Bot User (the account that will create auto-posts)
	BotUserID uuid.UUID `gorm:"not null;index;uniqueIndex:idx_auto_post_bot_user"`
	BotUser   User      `gorm:"foreignKey:BotUserID"`

	// Settings
	IsEnabled    bool   `gorm:"default:false;index"`
	CronSchedule string `gorm:"not null;default:'0 * * * *'"` // Default: every hour (minute 0)
	Model        string `gorm:"type:varchar(50);default:'gpt-4o-mini'"`

	// Topic Configuration (JSON array of topics/prompts)
	Topics datatypes.JSON `gorm:"type:jsonb;not null"` // ["topic1", "topic2", "topic3"]

	// Post Generation Settings
	MaxTokens   int    `gorm:"default:1500"`
	Temperature string `gorm:"type:varchar(10);default:'0.8'"`

	// Content Style & Variation
	Tone              string         `gorm:"type:varchar(50);default:'neutral'"` // neutral, casual, professional, humorous, controversial
	EnableVariations  bool           `gorm:"default:true"`                       // Enable title/content variations
	VariationStyle    datatypes.JSON `gorm:"type:jsonb"`                         // Settings for variations (emoji, punchlines, etc.)

	// Moderation & Approval
	RequireApproval   bool   `gorm:"default:false"`                  // Require manual approval before posting
	SensitiveTopics   datatypes.JSON `gorm:"type:jsonb"`              // Topics that require approval

	// Batch Generation
	BatchSize         int    `gorm:"default:1"`                      // Number of posts to generate per batch
	UseBatchMode      bool   `gorm:"default:false"`                  // Enable batch generation

	// Statistics
	TotalPostsGenerated int        `gorm:"default:0"`
	LastGeneratedAt     *time.Time `gorm:"index"`
	LastError           *string    `gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AutoPostSetting) TableName() string {
	return "auto_post_settings"
}

// BeforeCreate hook to generate UUID before creating AutoPostSetting
func (a *AutoPostSetting) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// AutoPostLog stores history of auto-generated posts
type AutoPostLog struct {
	ID uuid.UUID `gorm:"primaryKey;type:uuid"`

	// References
	SettingID uuid.UUID        `gorm:"not null;index"`
	Setting   AutoPostSetting  `gorm:"foreignKey:SettingID"`
	PostID    *uuid.UUID       `gorm:"index"`
	Post      *Post            `gorm:"foreignKey:PostID"`

	// Generation Details
	Topic          string `gorm:"type:varchar(500);not null"`
	GeneratedTitle string `gorm:"type:varchar(300)"`
	Status         string `gorm:"type:varchar(20);not null;default:'pending';index"` // pending, pending_approval, approved, success, failed, rejected
	ErrorMessage   *string `gorm:"type:text"`

	// AI Usage Stats
	TokensUsed         int `gorm:"default:0"`
	PromptTokens       int `gorm:"default:0"`
	CompletionTokens   int `gorm:"default:0"`

	// Metadata & Analytics
	Metadata           datatypes.JSON `gorm:"type:jsonb"` // Store tone, variation_type, engagement predictions, etc.
	TitleVariationUsed *string        `gorm:"type:varchar(500)"` // Which title variation was used

	// Approval Workflow
	ApprovedBy         *uuid.UUID `gorm:"index"` // User who approved
	ApprovedAt         *time.Time
	RejectedBy         *uuid.UUID `gorm:"index"` // User who rejected
	RejectedAt         *time.Time
	RejectionReason    *string    `gorm:"type:text"`

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (AutoPostLog) TableName() string {
	return "auto_post_logs"
}

// BeforeCreate hook to generate UUID before creating AutoPostLog
func (a *AutoPostLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
