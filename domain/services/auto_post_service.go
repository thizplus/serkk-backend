package services

import (
	"context"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
)

type AutoPostService interface {
	// Settings Management
	CreateSetting(ctx context.Context, req *dto.CreateAutoPostSettingRequest) (*dto.AutoPostSettingResponse, error)
	GetSetting(ctx context.Context, id uuid.UUID) (*dto.AutoPostSettingResponse, error)
	GetSettingByBotUserID(ctx context.Context, botUserID uuid.UUID) (*dto.AutoPostSettingResponse, error)
	UpdateSetting(ctx context.Context, id uuid.UUID, req *dto.UpdateAutoPostSettingRequest) (*dto.AutoPostSettingResponse, error)
	DeleteSetting(ctx context.Context, id uuid.UUID) error
	ListSettings(ctx context.Context, offset, limit int) ([]*dto.AutoPostSettingResponse, int64, error)

	// Enable/Disable
	EnableSetting(ctx context.Context, id uuid.UUID) error
	DisableSetting(ctx context.Context, id uuid.UUID) error

	// Auto-Post Generation
	GenerateAndPost(ctx context.Context, settingID uuid.UUID, topic *string) (*dto.TriggerAutoPostResponse, error)
	ProcessAllEnabledSettings(ctx context.Context) error

	// Logs
	GetLog(ctx context.Context, logID uuid.UUID) (*dto.AutoPostLogResponse, error)
	ListLogs(ctx context.Context, req *dto.ListAutoPostLogsRequest) (*dto.AutoPostLogListResponse, error)
	DeleteLog(ctx context.Context, logID uuid.UUID) error
}
