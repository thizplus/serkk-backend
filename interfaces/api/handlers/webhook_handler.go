package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"gofiber-template/domain/dto"
	"gofiber-template/domain/services"
	"gofiber-template/infrastructure/redis"
	"gofiber-template/infrastructure/websocket"
	"gofiber-template/pkg/utils"
)

type WebhookHandler struct {
	mediaService     services.MediaService
	postService      services.PostService
	messageService   services.MessageService
	redisService     *redis.RedisService
	notificationHub  *websocket.NotificationHub
}

func NewWebhookHandler(mediaService services.MediaService, postService services.PostService, messageService services.MessageService, redisService *redis.RedisService, notificationHub *websocket.NotificationHub) *WebhookHandler {
	return &WebhookHandler{
		mediaService:    mediaService,
		postService:     postService,
		messageService:  messageService,
		redisService:    redisService,
		notificationHub: notificationHub,
	}
}

// BunnyStreamWebhook handles video encoding status callbacks from Bunny Stream
// Bunny Stream sends webhook when video status changes
func (h *WebhookHandler) BunnyStreamWebhook(c *fiber.Ctx) error {
	// Parse the webhook payload
	var payload BunnyStreamWebhookPayload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Failed to parse Bunny webhook: %v", err)
		return utils.ValidationErrorResponse(c, "Invalid webhook payload")
	}

	// Log incoming webhook
	log.Printf("Received Bunny Stream webhook - VideoGUID: %s, LibraryID: %d, Status: %d, Progress: %d%%",
		payload.VideoGUID, payload.VideoLibraryID, payload.Status, payload.EncodeProgress)

	// Process the webhook asynchronously (don't block Bunny's webhook call)
	go h.processBunnyWebhook(context.Background(), payload)

	// Respond immediately to Bunny (they expect quick 200 OK)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Webhook received",
	})
}

// BunnyStreamWebhookPayload represents the payload sent by Bunny Stream
type BunnyStreamWebhookPayload struct {
	VideoLibraryID int64  `json:"VideoLibraryId"` // PascalCase from Bunny!
	VideoGUID      string `json:"VideoGuid"`      // PascalCase from Bunny!
	Status         int    `json:"Status"`         // PascalCase from Bunny!
	/*
	   Status codes from Bunny Stream:
	   0 = Queued
	   1 = Processing
	   2 = Encoding
	   3 = Finished (Ready to play)
	   4 = Resolution Finished
	   5 = Failed
	   6 = PresignedUploadStarted
	   7 = PresignedUploadFinished
	   8 = PresignedUploadFailed
	   9 = CaptionsGenerated
	   10 = TitleOrDescriptionGenerated
	*/

	// Optional fields (may not be in minimal callback)
	EncodeProgress       int    `json:"EncodeProgress,omitempty"`       // 0-100
	Width                int    `json:"Width,omitempty"`
	Height               int    `json:"Height,omitempty"`
	AvailableResolutions string `json:"AvailableResolutions,omitempty"`
	Length               int    `json:"Length,omitempty"`               // Duration in seconds
	Title                string `json:"Title,omitempty"`
	ThumbnailCount       int    `json:"ThumbnailCount,omitempty"`
}

// calculateProgress estimates encoding progress based on Bunny Stream status
func calculateProgress(status int, encodeProgress int) int {
	// If Bunny provides EncodeProgress, use it
	if encodeProgress > 0 && encodeProgress <= 100 {
		return encodeProgress
	}

	// Otherwise, estimate based on status
	switch status {
	case 0: // Queued
		return 5
	case 1: // Processing
		return 15
	case 2: // Encoding
		// If no progress provided, assume mid-encoding
		return 50
	case 6: // PresignedUploadStarted
		return 10
	case 7: // PresignedUploadFinished
		return 25
	case 3, 4, 9, 10: // Finished, Resolution Finished, etc.
		return 100
	case 5, 8: // Failed
		return 0
	default:
		return 0
	}
}

// processBunnyWebhook processes the webhook in background
func (h *WebhookHandler) processBunnyWebhook(ctx context.Context, payload BunnyStreamWebhookPayload) {
	videoID := payload.VideoGUID

	// Map Bunny status to our encoding status
	var encodingStatus string
	switch payload.Status {
	case 0, 1, 2, 6, 7: // Queued, Processing, Encoding, PresignedUploadStarted, PresignedUploadFinished
		encodingStatus = "processing"
	case 3, 4, 9, 10: // Finished, Resolution Finished, CaptionsGenerated, TitleOrDescriptionGenerated
		encodingStatus = "completed"
	case 5, 8: // Failed, PresignedUploadFailed
		encodingStatus = "failed"
	default:
		log.Printf("⚠️  Unknown Bunny Stream status code: %d", payload.Status)
		encodingStatus = "processing" // Default to processing for unknown statuses
	}

	// Update encoding status via media service
	err := h.mediaService.UpdateVideoEncodingStatus(ctx, videoID, encodingStatus, payload.EncodeProgress, payload.Width, payload.Height, payload.Length)
	if err != nil {
		log.Printf("Failed to update video encoding status for VideoID %s: %v", videoID, err)
		return
	}

	// Get media info to get user ID and media ID
	media, err := h.mediaService.GetMediaByVideoID(ctx, videoID)
	if err != nil {
		log.Printf("Failed to get media by VideoID %s: %v", videoID, err)
		return
	}

	// Update Redis queue status
	h.redisService.UpdateVideoEncodingStatus(ctx, media.ID, encodingStatus, payload.EncodeProgress, "")

	// Clear job from Redis if completed or failed
	if encodingStatus == "completed" || encodingStatus == "failed" {
		h.redisService.ClearVideoEncodingJob(ctx, media.ID)
	}

	// Broadcast WebSocket event to user
	if media != nil {
		// Calculate progress based on status and payload
		// Bunny Stream doesn't always send EncodeProgress, so we estimate it
		progress := calculateProgress(payload.Status, payload.EncodeProgress)
		h.broadcastEncodingUpdate(media, encodingStatus, progress)
	}

	// Source-specific actions based on sourceType
	if encodingStatus == "completed" && media != nil {
		sourceType := ""
		if media.SourceType != nil {
			sourceType = *media.SourceType
		}

		switch sourceType {
		case "message":
			// Update message JSONB with encoding status
			if media.SourceID != nil {
				err := h.messageService.UpdateMessageVideoStatus(ctx, *media.SourceID, media)
				if err != nil {
					log.Printf("Failed to update message video status for message %s: %v", media.SourceID, err)
				}
			}
		case "reel":
			// TODO: Handle reel video completion
			log.Printf("Reel video encoding completed: mediaID=%s", media.ID)
		case "comment":
			// TODO: Handle comment video completion
			log.Printf("Comment video encoding completed: mediaID=%s", media.ID)
		default:
			// ⭐ Many-to-Many pattern (Post, Media Library)
			// source_type = NULL means this is a Post video (standalone media)
			err := h.postService.PublishDraftPostsWithMedia(ctx, media.ID)
			if err != nil {
				log.Printf("Failed to auto-publish draft posts for media %s: %v", media.ID, err)
			} else {
				log.Printf("✅ Checked draft posts for media %s (Many-to-Many pattern)", media.ID)
			}
		}
	}

	log.Printf("Successfully updated video encoding status - VideoID: %s, Status: %s, Progress: %d%%",
		videoID, encodingStatus, payload.EncodeProgress)
}

// broadcastEncodingUpdate sends WebSocket event to user about encoding progress
func (h *WebhookHandler) broadcastEncodingUpdate(media *dto.MediaResponse, status string, progress int) {
	if h.notificationHub == nil {
		log.Printf("⚠️  NotificationHub not available, skipping WebSocket broadcast")
		return
	}

	// Prepare payload
	payload := map[string]interface{}{
		"mediaId":  media.ID.String(),
		"status":   status,
		"progress": progress,
		"message":  getStatusMessage(status),
	}

	// Add sourceType and sourceId if available
	if media.SourceType != nil {
		payload["sourceType"] = *media.SourceType
	}
	if media.SourceID != nil {
		payload["sourceId"] = media.SourceID.String()
	}

	// Add video metadata when completed
	// NOTE: Removed HLSURL as we no longer use Bunny Stream
	if status == "completed" {
		if media.Thumbnail != "" {
			payload["thumbnail"] = media.Thumbnail
		}
		if media.Width > 0 {
			payload["width"] = media.Width
			payload["height"] = media.Height
		}
		if media.Duration > 0 {
			payload["duration"] = int(media.Duration)
		}
	}

	// Create NotificationMessage
	message := &websocket.NotificationMessage{
		Type:    getEventType(status),
		Payload: payload,
	}

	// Send to user via NotificationHub
	h.notificationHub.SendToUser(media.UserID, message)

	log.Printf("📢 Video encoding notification sent: user=%s, mediaId=%s, sourceType=%v, status=%s, progress=%d%%",
		media.UserID, media.ID, media.SourceType, status, progress)
}

// getEventType returns WebSocket event type based on encoding status
func getEventType(status string) string {
	switch status {
	case "completed":
		return "video.encoding.completed"
	case "failed":
		return "video.encoding.failed"
	case "processing":
		return "video.encoding.progress"
	default:
		return "video.encoding.update"
	}
}

// getStatusMessage returns user-friendly message based on status
func getStatusMessage(status string) string {
	switch status {
	case "completed":
		return "Your video has been processed and is ready to view!"
	case "failed":
		return "Video processing failed. Please try uploading again."
	case "processing":
		return "Your video is being processed..."
	default:
		return "Video encoding status updated"
	}
}
