package eventhandlers

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"

	"gofiber-template/domain/events"
	"gofiber-template/domain/models"
	"gofiber-template/domain/ports"
	"gofiber-template/domain/repositories"
	"gofiber-template/domain/services"
)

// AuthServiceEventHandler - รับ events จาก Auth Service (V2 - Minimal Identity Event)
type AuthServiceEventHandler struct {
	identityService *services.UsersIdentityService
	profileRepo     repositories.UserProfileRepository
}

// NewAuthServiceEventHandler creates a new auth service event handler
func NewAuthServiceEventHandler(
	identityService *services.UsersIdentityService,
	profileRepo repositories.UserProfileRepository,
) *AuthServiceEventHandler {
	return &AuthServiceEventHandler{
		identityService: identityService,
		profileRepo:     profileRepo,
	}
}

// HandleUserEvent handles all user events (created/updated/deleted)
// ใช้ action field แยก action แทนที่จะแยก subject
func (h *AuthServiceEventHandler) HandleUserEvent(ctx context.Context, msg *ports.EventMessage) error {
	start := time.Now()

	// Parse Auth Service V2 event format
	var event events.AuthServiceUserEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("❌ Failed to unmarshal Auth Service event: %v", err)
		return err
	}

	// Validate event
	if !event.IsValid() {
		log.Printf("⚠️  Invalid event data: %+v", event)
		return nil // Don't retry invalid events (Ack message)
	}

	log.Printf("📥 [Auth Service V2] %s - UserID: %s, Email: %s, Username: %s, RequestID: %s",
		event.Action, event.ID, event.Email, event.Username, event.RequestID)

	// Route to correct handler based on action
	var err error
	switch event.Action {
	case "created":
		err = h.handleCreated(ctx, event)
	case "updated":
		err = h.handleUpdated(ctx, event)
	case "deleted":
		err = h.handleDeleted(ctx, event)
	default:
		log.Printf("⚠️  Unknown action: %s", event.Action)
		return nil // Ack message (don't retry)
	}

	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Failed to sync %s (UserID: %s): %v [%dms]",
			event.Action, event.ID, err, duration.Milliseconds())
		return err // Trigger retry
	}

	log.Printf("✅ Successfully synced %s - UserID: %s [%dms]",
		event.Action, event.ID, duration.Milliseconds())

	return nil
}

func (h *AuthServiceEventHandler) handleCreated(ctx context.Context, event events.AuthServiceUserEvent) error {
	userID, err := uuid.Parse(event.ID)
	if err != nil {
		log.Printf("❌ Invalid UUID: %v", err)
		return err
	}

	// 1. Create identity record
	identity := &models.UsersIdentity{
		ID:       userID,
		Email:    event.Email,
		Username: event.Username,
		SyncedAt: time.Now(),
	}

	if err := h.identityService.Upsert(ctx, identity); err != nil {
		return err
	}

	log.Printf("✅ Created identity for user: %s (email: %s)", event.Username, event.Email)

	// 2. Create default profile (ใช้ username เป็น default display name)
	profile := &models.UserProfile{
		UserID:      userID,
		DisplayName: event.Username, // Default: ใช้ username
		Avatar:      "",             // Empty avatar
		Bio:         "",             // Empty bio
	}

	if err := h.profileRepo.Create(ctx, profile); err != nil {
		// ถ้า profile มีอยู่แล้ว (duplicate) ไม่ต้อง error
		// เพราะอาจเป็นการ replay event หรือ retry
		log.Printf("⚠️  Profile creation skipped (may already exist): %v", err)
	} else {
		log.Printf("✅ Created default profile for user: %s", event.Username)
	}

	return nil
}

func (h *AuthServiceEventHandler) handleUpdated(ctx context.Context, event events.AuthServiceUserEvent) error {
	userID, err := uuid.Parse(event.ID)
	if err != nil {
		return err
	}

	// Update identity only (email, username)
	// NOTE: V2 ไม่มี displayName, avatar ใน event
	identity := &models.UsersIdentity{
		ID:       userID,
		Email:    event.Email,
		Username: event.Username,
		SyncedAt: time.Now(),
	}

	if err := h.identityService.Upsert(ctx, identity); err != nil {
		return err
	}

	log.Printf("✅ Updated identity for user: %s (email: %s)", event.Username, event.Email)

	// NOTE: ไม่ update profile เพราะ displayName, avatar ไม่มีใน V2 event
	// User ต้องเรียก API ของเราเพื่อ update profile เอง
	return nil
}

func (h *AuthServiceEventHandler) handleDeleted(ctx context.Context, event events.AuthServiceUserEvent) error {
	userID, err := uuid.Parse(event.ID)
	if err != nil {
		return err
	}

	// Soft delete identity
	// NOTE: Profile จะถูก cascade delete ด้วย (ถ้าตั้ง FK constraint)
	if err := h.identityService.SoftDelete(ctx, userID); err != nil {
		return err
	}

	log.Printf("✅ Soft deleted user: %s (ID: %s)", event.Username, event.ID)
	return nil
}
