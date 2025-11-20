package serviceimpl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/ai"
)

type autoPostServiceImpl struct {
	settingRepo  repositories.AutoPostSettingRepository
	logRepo      repositories.AutoPostLogRepository
	postService  services.PostService
	aiService    ai.OpenAIService
	logger       zerolog.Logger
}

func NewAutoPostService(
	settingRepo repositories.AutoPostSettingRepository,
	logRepo repositories.AutoPostLogRepository,
	postService services.PostService,
	aiService ai.OpenAIService,
	logger zerolog.Logger,
) services.AutoPostService {
	return &autoPostServiceImpl{
		settingRepo:  settingRepo,
		logRepo:      logRepo,
		postService:  postService,
		aiService:    aiService,
		logger:       logger,
	}
}

// ========== Settings Management ==========

func (s *autoPostServiceImpl) CreateSetting(ctx context.Context, req *dto.CreateAutoPostSettingRequest) (*dto.AutoPostSettingResponse, error) {
	// Convert topics to JSON
	topicsJSON, err := json.Marshal(req.Topics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal topics: %w", err)
	}

	model := req.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1500
	}

	setting := &models.AutoPostSetting{
		BotUserID:    req.BotUserID,
		IsEnabled:    req.IsEnabled,
		CronSchedule: req.CronSchedule,
		Model:        model,
		Topics:       topicsJSON,
		MaxTokens:    maxTokens,
		Temperature:  "0.8",
	}

	if err := s.settingRepo.Create(ctx, setting); err != nil {
		return nil, err
	}

	return s.toSettingResponse(setting, req.Topics), nil
}

func (s *autoPostServiceImpl) GetSetting(ctx context.Context, id uuid.UUID) (*dto.AutoPostSettingResponse, error) {
	setting, err := s.settingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.New("setting not found")
	}

	topics, err := s.parseTopics(setting.Topics)
	if err != nil {
		return nil, err
	}

	return s.toSettingResponse(setting, topics), nil
}

func (s *autoPostServiceImpl) GetSettingByBotUserID(ctx context.Context, botUserID uuid.UUID) (*dto.AutoPostSettingResponse, error) {
	setting, err := s.settingRepo.GetByBotUserID(ctx, botUserID)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.New("setting not found")
	}

	topics, err := s.parseTopics(setting.Topics)
	if err != nil {
		return nil, err
	}

	return s.toSettingResponse(setting, topics), nil
}

func (s *autoPostServiceImpl) UpdateSetting(ctx context.Context, id uuid.UUID, req *dto.UpdateAutoPostSettingRequest) (*dto.AutoPostSettingResponse, error) {
	setting, err := s.settingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.New("setting not found")
	}

	// Update fields
	if req.IsEnabled != nil {
		setting.IsEnabled = *req.IsEnabled
	}
	if req.CronSchedule != nil {
		setting.CronSchedule = *req.CronSchedule
	}
	if req.Model != nil {
		setting.Model = *req.Model
	}
	if req.Topics != nil {
		topicsJSON, err := json.Marshal(*req.Topics)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal topics: %w", err)
		}
		setting.Topics = topicsJSON
	}
	if req.MaxTokens != nil {
		setting.MaxTokens = *req.MaxTokens
	}

	if err := s.settingRepo.Update(ctx, setting); err != nil {
		return nil, err
	}

	topics, err := s.parseTopics(setting.Topics)
	if err != nil {
		return nil, err
	}

	return s.toSettingResponse(setting, topics), nil
}

func (s *autoPostServiceImpl) DeleteSetting(ctx context.Context, id uuid.UUID) error {
	return s.settingRepo.Delete(ctx, id)
}

func (s *autoPostServiceImpl) ListSettings(ctx context.Context, offset, limit int) ([]*dto.AutoPostSettingResponse, int64, error) {
	settings, total, err := s.settingRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.AutoPostSettingResponse, len(settings))
	for i, setting := range settings {
		topics, err := s.parseTopics(setting.Topics)
		if err != nil {
			s.logger.Warn().Err(err).Msg("Failed to parse topics")
			topics = []string{}
		}
		responses[i] = s.toSettingResponse(setting, topics)
	}

	return responses, total, nil
}

func (s *autoPostServiceImpl) EnableSetting(ctx context.Context, id uuid.UUID) error {
	setting, err := s.settingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.New("setting not found")
	}

	setting.IsEnabled = true
	return s.settingRepo.Update(ctx, setting)
}

func (s *autoPostServiceImpl) DisableSetting(ctx context.Context, id uuid.UUID) error {
	setting, err := s.settingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if setting == nil {
		return errors.New("setting not found")
	}

	setting.IsEnabled = false
	return s.settingRepo.Update(ctx, setting)
}

// ========== Auto-Post Generation ==========

func (s *autoPostServiceImpl) GenerateAndPost(ctx context.Context, settingID uuid.UUID, customTopic *string) (*dto.TriggerAutoPostResponse, error) {
	// Get setting
	setting, err := s.settingRepo.GetByID(ctx, settingID)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.New("setting not found")
	}

	// Parse topics
	topics, err := s.parseTopics(setting.Topics)
	if err != nil {
		return nil, err
	}

	// Select topic
	var selectedTopic string
	if customTopic != nil && *customTopic != "" {
		selectedTopic = *customTopic
	} else if len(topics) > 0 {
		// Randomly select a topic
		rand.Seed(time.Now().UnixNano())
		selectedTopic = topics[rand.Intn(len(topics))]
	} else {
		return nil, errors.New("no topics configured")
	}

	// Create log entry
	log := &models.AutoPostLog{
		SettingID:      settingID,
		Topic:          selectedTopic,
		GeneratedTitle: "",
		Status:         "pending",
	}

	if err := s.logRepo.Create(ctx, log); err != nil {
		return nil, err
	}

	// Generate content using AI
	s.logger.Info().
		Str("topic", selectedTopic).
		Str("model", setting.Model).
		Msg("Generating AI content")

	content, err := s.aiService.GeneratePostContent(ctx, selectedTopic)
	if err != nil {
		// Update log with error
		errMsg := err.Error()
		log.Status = "failed"
		log.ErrorMessage = &errMsg
		_ = s.logRepo.Update(ctx, log)

		// Update setting error
		_ = s.settingRepo.UpdateLastError(ctx, settingID, &errMsg)

		return &dto.TriggerAutoPostResponse{
			Success: false,
			Error:   &errMsg,
			Log:     s.toLogResponse(log),
		}, nil
	}

	log.GeneratedTitle = content.Title

	// Create post
	createReq := &dto.CreatePostRequest{
		Title:   content.Title,
		Content: content.Content,
		Tags:    content.Tags,
		IsDraft: false,
	}

	post, err := s.postService.CreatePost(ctx, setting.BotUserID, createReq)
	if err != nil {
		// Update log with error
		errMsg := err.Error()
		log.Status = "failed"
		log.ErrorMessage = &errMsg
		_ = s.logRepo.Update(ctx, log)

		// Update setting error
		_ = s.settingRepo.UpdateLastError(ctx, settingID, &errMsg)

		return &dto.TriggerAutoPostResponse{
			Success: false,
			Error:   &errMsg,
			Log:     s.toLogResponse(log),
		}, nil
	}

	// Update log with success
	log.Status = "success"
	log.PostID = &post.ID
	_ = s.logRepo.Update(ctx, log)

	// Update setting statistics
	_ = s.settingRepo.IncrementPostCount(ctx, settingID)

	s.logger.Info().
		Str("post_id", post.ID.String()).
		Str("topic", selectedTopic).
		Msg("Auto-post created successfully")

	return &dto.TriggerAutoPostResponse{
		Success: true,
		Post:    post,
		Log:     s.toLogResponse(log),
	}, nil
}

func (s *autoPostServiceImpl) ProcessAllEnabledSettings(ctx context.Context) error {
	settings, err := s.settingRepo.GetAllEnabled(ctx)
	if err != nil {
		return err
	}

	s.logger.Info().Int("count", len(settings)).Msg("Processing enabled auto-post settings")

	for _, setting := range settings {
		_, err := s.GenerateAndPost(ctx, setting.ID, nil)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("setting_id", setting.ID.String()).
				Msg("Failed to generate and post")
			continue
		}
	}

	return nil
}

// ========== Logs ==========

func (s *autoPostServiceImpl) GetLog(ctx context.Context, logID uuid.UUID) (*dto.AutoPostLogResponse, error) {
	log, err := s.logRepo.GetByID(ctx, logID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, errors.New("log not found")
	}

	return s.toLogResponse(log), nil
}

func (s *autoPostServiceImpl) ListLogs(ctx context.Context, req *dto.ListAutoPostLogsRequest) (*dto.AutoPostLogListResponse, error) {
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}

	var logs []*models.AutoPostLog
	var total int64
	var err error

	if req.SettingID != nil {
		logs, total, err = s.logRepo.ListBySettingID(ctx, *req.SettingID, req.Offset, limit)
	} else if req.Status != nil {
		logs, total, err = s.logRepo.ListByStatus(ctx, *req.Status, req.Offset, limit)
	} else {
		logs, total, err = s.logRepo.List(ctx, req.Offset, limit)
	}

	if err != nil {
		return nil, err
	}

	responses := make([]dto.AutoPostLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = *s.toLogResponse(log)
	}

	return &dto.AutoPostLogListResponse{
		Logs: responses,
		Meta: dto.PaginationMeta{
			Offset: req.Offset,
			Limit:  limit,
			Total:  &total,
		},
	}, nil
}

func (s *autoPostServiceImpl) DeleteLog(ctx context.Context, logID uuid.UUID) error {
	return s.logRepo.Delete(ctx, logID)
}

// ========== Helper Methods ==========

func (s *autoPostServiceImpl) parseTopics(topicsJSON []byte) ([]string, error) {
	var topics []string
	if err := json.Unmarshal(topicsJSON, &topics); err != nil {
		return nil, fmt.Errorf("failed to parse topics: %w", err)
	}
	return topics, nil
}

func (s *autoPostServiceImpl) toSettingResponse(setting *models.AutoPostSetting, topics []string) *dto.AutoPostSettingResponse {
	resp := &dto.AutoPostSettingResponse{
		ID:                  setting.ID,
		BotUserID:           setting.BotUserID,
		IsEnabled:           setting.IsEnabled,
		CronSchedule:        setting.CronSchedule,
		Model:               setting.Model,
		Topics:              topics,
		MaxTokens:           setting.MaxTokens,
		TotalPostsGenerated: setting.TotalPostsGenerated,
		LastGeneratedAt:     setting.LastGeneratedAt,
		LastError:           setting.LastError,
		CreatedAt:           setting.CreatedAt,
		UpdatedAt:           setting.UpdatedAt,
	}

	if setting.BotUser.ID != uuid.Nil {
		resp.BotUser = &dto.UserResponse{
			ID:       setting.BotUser.ID,
			Username: setting.BotUser.Username,
			Email:    setting.BotUser.Email,
		}
	}

	return resp
}

func (s *autoPostServiceImpl) toLogResponse(log *models.AutoPostLog) *dto.AutoPostLogResponse {
	resp := &dto.AutoPostLogResponse{
		ID:             log.ID,
		SettingID:      log.SettingID,
		PostID:         log.PostID,
		Topic:          log.Topic,
		GeneratedTitle: log.GeneratedTitle,
		Status:         log.Status,
		ErrorMessage:   log.ErrorMessage,
		TokensUsed:     log.TokensUsed,
		CreatedAt:      log.CreatedAt,
		UpdatedAt:      log.UpdatedAt,
	}

	// Add post if available
	if log.Post != nil && log.Post.ID != uuid.Nil {
		resp.Post = &dto.PostResponse{
			ID:      log.Post.ID,
			Title:   log.Post.Title,
			Content: log.Post.Content,
		}
	}

	return resp
}
