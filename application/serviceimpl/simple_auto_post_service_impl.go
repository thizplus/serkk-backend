package serviceimpl

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/pkg/ai"
	"gorm.io/gorm"
)

type QueueItem struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key"`
	BotUserID     uuid.UUID  `gorm:"type:uuid;not null"`
	Topic         string     `gorm:"type:text;not null"`
	Tone          string     `gorm:"type:varchar(50)"`
	Status        string     `gorm:"type:varchar(20)"`
	PostID        *uuid.UUID `gorm:"type:uuid"`
	GeneratedTitle *string   `gorm:"type:text"`
	ErrorMessage  *string    `gorm:"type:text"`
	TokensUsed    *int       `gorm:"type:integer"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	CompletedAt   *time.Time `gorm:"type:timestamp"`
}

func (QueueItem) TableName() string {
	return "auto_post_queue"
}

type simpleAutoPostServiceImpl struct {
	db            *gorm.DB
	openAIService ai.OpenAIService
	postService   services.PostService
}

func NewSimpleAutoPostService(
	db *gorm.DB,
	openAIService ai.OpenAIService,
	postService services.PostService,
) services.SimpleAutoPostService {
	return &simpleAutoPostServiceImpl{
		db:            db,
		openAIService: openAIService,
		postService:   postService,
	}
}

func (s *simpleAutoPostServiceImpl) ProcessNextTopic(ctx context.Context) error {
	// 1. หา topic ถัดไปที่ status = pending
	var item QueueItem
	err := s.db.WithContext(ctx).
		Where("status = ?", "pending").
		Order("created_at ASC").
		First(&item).Error

	if err == gorm.ErrRecordNotFound {
		log.Println("ℹ️  No pending topics in queue")
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to get next topic: %w", err)
	}

	log.Printf("📝 Processing topic: %s (tone: %s)", item.Topic, item.Tone)

	// 2. Generate content ด้วย AI
	content, err := s.openAIService.GeneratePostContentWithStyle(ctx, item.Topic, item.Tone, false)
	if err != nil {
		// Mark as failed
		errorMsg := err.Error()
		s.db.WithContext(ctx).Model(&item).Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": errorMsg,
			"completed_at":  time.Now(),
		})
		return fmt.Errorf("failed to generate content: %w", err)
	}

	log.Printf("✅ Generated: %s", content.Title)

	// 3. สร้างโพสต์
	createReq := &dto.CreatePostRequest{
		Title:   content.Title,
		Content: content.Content,
		Tags:    content.Tags,
	}

	createdPost, err := s.postService.CreatePost(ctx, item.BotUserID, createReq)
	if err != nil {
		// Mark as failed
		errorMsg := err.Error()
		s.db.WithContext(ctx).Model(&item).Updates(map[string]interface{}{
			"status":         "failed",
			"error_message":  errorMsg,
			"generated_title": content.Title,
			"tokens_used":    content.TotalTokens,
			"completed_at":   time.Now(),
		})
		return fmt.Errorf("failed to create post: %w", err)
	}

	log.Printf("🎉 Post created: %s", createdPost.ID)

	// 4. Update status = completed
	now := time.Now()
	err = s.db.WithContext(ctx).Model(&item).Updates(map[string]interface{}{
		"status":          "completed",
		"post_id":         createdPost.ID,
		"generated_title": content.Title,
		"tokens_used":     content.TotalTokens,
		"completed_at":    now,
	}).Error

	if err != nil {
		log.Printf("⚠️  Warning: failed to update queue item: %v", err)
	}

	log.Printf("✅ Topic completed successfully!")

	return nil
}
