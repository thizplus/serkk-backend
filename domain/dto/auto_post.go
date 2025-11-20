package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAutoPostSettingRequest - Request for creating auto-post settings
type CreateAutoPostSettingRequest struct {
	BotUserID    uuid.UUID `json:"botUserId" validate:"required,uuid"`
	IsEnabled    bool      `json:"isEnabled"`
	CronSchedule string    `json:"cronSchedule" validate:"required,min=9,max=100"` // Cron expression
	Model        string    `json:"model" validate:"omitempty,oneof=gpt-4 gpt-4-turbo gpt-3.5-turbo gpt-4o gpt-4o-mini"`
	Topics       []string  `json:"topics" validate:"required,min=1,max=50,dive,min=3,max=500"`
	MaxTokens    int       `json:"maxTokens" validate:"omitempty,min=100,max=4000"`
}

// UpdateAutoPostSettingRequest - Request for updating auto-post settings
type UpdateAutoPostSettingRequest struct {
	IsEnabled    *bool     `json:"isEnabled"`
	CronSchedule *string   `json:"cronSchedule" validate:"omitempty,min=9,max=100"`
	Model        *string   `json:"model" validate:"omitempty,oneof=gpt-4 gpt-4-turbo gpt-3.5-turbo gpt-4o gpt-4o-mini"`
	Topics       *[]string `json:"topics" validate:"omitempty,min=1,max=50,dive,min=3,max=500"`
	MaxTokens    *int      `json:"maxTokens" validate:"omitempty,min=100,max=4000"`
}

// AutoPostSettingResponse - Response for auto-post settings
type AutoPostSettingResponse struct {
	ID                  uuid.UUID  `json:"id"`
	BotUserID           uuid.UUID  `json:"botUserId"`
	BotUser             *UserResponse `json:"botUser,omitempty"`
	IsEnabled           bool       `json:"isEnabled"`
	CronSchedule        string     `json:"cronSchedule"`
	Model               string     `json:"model"`
	Topics              []string   `json:"topics"`
	MaxTokens           int        `json:"maxTokens"`
	TotalPostsGenerated int        `json:"totalPostsGenerated"`
	LastGeneratedAt     *time.Time `json:"lastGeneratedAt,omitempty"`
	LastError           *string    `json:"lastError,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

// AutoPostLogResponse - Response for auto-post log
type AutoPostLogResponse struct {
	ID             uuid.UUID         `json:"id"`
	SettingID      uuid.UUID         `json:"settingId"`
	PostID         *uuid.UUID        `json:"postId,omitempty"`
	Post           *PostResponse     `json:"post,omitempty"`
	Topic          string            `json:"topic"`
	GeneratedTitle string            `json:"generatedTitle"`
	Status         string            `json:"status"` // pending, success, failed
	ErrorMessage   *string           `json:"errorMessage,omitempty"`
	TokensUsed     int               `json:"tokensUsed"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

// ListAutoPostLogsRequest - Request for listing auto-post logs
type ListAutoPostLogsRequest struct {
	SettingID *uuid.UUID `query:"settingId" validate:"omitempty,uuid"`
	Status    *string    `query:"status" validate:"omitempty,oneof=pending success failed"`
	Limit     int        `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int        `query:"offset" validate:"omitempty,min=0"`
}

// AutoPostLogListResponse - Response for listing auto-post logs
type AutoPostLogListResponse struct {
	Logs []AutoPostLogResponse `json:"logs"`
	Meta PaginationMeta        `json:"meta"`
}

// TriggerAutoPostRequest - Request for manually triggering an auto-post
type TriggerAutoPostRequest struct {
	Topic *string `json:"topic" validate:"omitempty,min=3,max=500"` // If empty, randomly select from configured topics
}

// TriggerAutoPostResponse - Response for triggered auto-post
type TriggerAutoPostResponse struct {
	Success bool              `json:"success"`
	Post    *PostResponse     `json:"post,omitempty"`
	Log     *AutoPostLogResponse `json:"log,omitempty"`
	Error   *string           `json:"error,omitempty"`
}
