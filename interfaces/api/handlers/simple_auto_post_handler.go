package handlers

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	apperrors "gofiber-template/pkg/errors"
	"gofiber-template/pkg/utils"
	"gorm.io/gorm"
)

type SimpleAutoPostHandler struct {
	db *gorm.DB
}

func NewSimpleAutoPostHandler(db *gorm.DB) *SimpleAutoPostHandler {
	return &SimpleAutoPostHandler{
		db: db,
	}
}

type QueueItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BotUserID  uuid.UUID `gorm:"type:uuid;not null"`
	Topic      string    `gorm:"type:text;not null"`
	Tone       string    `gorm:"type:varchar(50);default:'neutral'"`
	Status     string    `gorm:"type:varchar(20);default:'pending'"`
}

func (QueueItem) TableName() string {
	return "auto_post_queue"
}

type UploadCSVResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	TotalImported int    `json:"totalImported"`
	TotalFailed   int    `json:"totalFailed"`
}

// UploadCSV godoc
// @Summary Upload CSV file to import topics
// @Description Upload a CSV file with topics to add to auto-post queue
// @Tags simple-auto-post
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV file (format: category,topic,tone)"
// @Param botUserId formData string true "Bot User ID (UUID)"
// @Success 200 {object} UploadCSVResponse
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /simple-auto-post/upload [post]
func (h *SimpleAutoPostHandler) UploadCSV(c *fiber.Ctx) error {
	// Get bot user ID from form
	botUserIDStr := c.FormValue("botUserId")
	if botUserIDStr == "" {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("botUserId is required"))
	}

	botUserID, err := uuid.Parse(botUserIDStr)
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Invalid botUserId format"))
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("File is required"))
	}

	// Check file extension
	if file.Header.Get("Content-Type") != "text/csv" &&
	   !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Only CSV files are allowed"))
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to open file"))
	}
	defer fileReader.Close()

	// Parse CSV
	reader := csv.NewReader(fileReader)

	successCount := 0
	failCount := 0
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return utils.ErrorResponse(c, apperrors.ErrBadRequest.WithMessage("Failed to parse CSV: " + err.Error()))
		}

		lineNum++

		// Skip header row
		if lineNum == 1 {
			continue
		}

		// Validate record length
		if len(record) < 2 {
			failCount++
			continue
		}

		// Parse record
		var topic, tone string
		if len(record) >= 3 {
			// Format: category,topic,tone
			topic = record[1]
			tone = record[2]
		} else {
			// Format: topic,tone
			topic = record[0]
			tone = record[1]
		}

		// Skip empty topics
		if topic == "" {
			failCount++
			continue
		}

		// Default tone
		if tone == "" {
			tone = "neutral"
		}

		// Insert to database
		item := &QueueItem{
			BotUserID: botUserID,
			Topic:     topic,
			Tone:      tone,
			Status:    "pending",
		}

		if err := h.db.Create(item).Error; err != nil {
			// Log error for debugging
			if failCount == 0 {
				// Log only first error to avoid spam
				c.Context().Logger().Printf("Failed to insert topic: %v", err)
			}
			failCount++
		} else {
			successCount++
		}
	}

	return c.JSON(UploadCSVResponse{
		Success:       true,
		Message:       "CSV imported successfully",
		TotalImported: successCount,
		TotalFailed:   failCount,
	})
}

// GetQueueStatus godoc
// @Summary Get queue status
// @Description Get statistics about the auto-post queue
// @Tags simple-auto-post
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /simple-auto-post/queue/status [get]
func (h *SimpleAutoPostHandler) GetQueueStatus(c *fiber.Ctx) error {
	var totalCount int64
	var pendingCount int64
	var completedCount int64
	var failedCount int64

	// Count total
	if err := h.db.Model(&QueueItem{}).Count(&totalCount).Error; err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to count total"))
	}

	// Count by status
	h.db.Model(&QueueItem{}).Where("status = ?", "pending").Count(&pendingCount)
	h.db.Model(&QueueItem{}).Where("status = ?", "completed").Count(&completedCount)
	h.db.Model(&QueueItem{}).Where("status = ?", "failed").Count(&failedCount)

	return c.JSON(fiber.Map{
		"total":     totalCount,
		"pending":   pendingCount,
		"completed": completedCount,
		"failed":    failedCount,
	})
}

// ListQueue godoc
// @Summary List queue items
// @Description Get paginated list of queue items
// @Tags simple-auto-post
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (pending/completed/failed)"
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /simple-auto-post/queue [get]
func (h *SimpleAutoPostHandler) ListQueue(c *fiber.Ctx) error {
	status := c.Query("status")
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	var items []QueueItem
	query := h.db.Model(&QueueItem{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to fetch queue"))
	}

	var total int64
	query.Count(&total)

	return c.JSON(fiber.Map{
		"items": items,
		"meta": fiber.Map{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// SetupTable godoc
// @Summary Setup auto-post queue table
// @Description Create the auto_post_queue table if it doesn't exist
// @Tags simple-auto-post
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.Response
// @Security BearerAuth
// @Router /simple-auto-post/setup [post]
func (h *SimpleAutoPostHandler) SetupTable(c *fiber.Ctx) error {
	// Check if table exists
	if h.db.Migrator().HasTable(&QueueItem{}) {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Table already exists",
		})
	}

	// Create table
	if err := h.db.AutoMigrate(&QueueItem{}); err != nil {
		return utils.ErrorResponse(c, apperrors.ErrInternal.WithMessage("Failed to create table: "+err.Error()))
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Table created successfully",
	})
}
