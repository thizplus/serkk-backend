package main

import (
	"fmt"
	"log"

	"gofiber-template/pkg/config"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check conversation
	conversationID := "f5c4d023-5ab7-4159-8816-40d231464d2b"

	var conversation struct {
		ID            uuid.UUID
		User1ID       uuid.UUID
		User2ID       uuid.UUID
		LastMessageID *uuid.UUID
	}

	err = db.Table("conversations").Where("id = ?", conversationID).First(&conversation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			fmt.Printf("❌ Conversation not found: %s\n", conversationID)
		} else {
			fmt.Printf("❌ Error querying conversation: %v\n", err)
		}
		return
	}

	fmt.Printf("✓ Conversation found\n")
	fmt.Printf("  ID: %s\n", conversation.ID)
	fmt.Printf("  User1ID: %s\n", conversation.User1ID)
	fmt.Printf("  User2ID: %s\n", conversation.User2ID)
	fmt.Printf("  LastMessageID: %v\n", conversation.LastMessageID)

	// Get user info
	var users []struct {
		ID       uuid.UUID
		Username string
	}
	err = db.Table("users").Where("id IN ?", []uuid.UUID{conversation.User1ID, conversation.User2ID}).Find(&users).Error
	if err != nil {
		fmt.Printf("⚠ Warning: Failed to get user info: %v\n", err)
	} else {
		fmt.Println("\n✓ Participants:")
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Username, user.ID)
		}
	}
}
