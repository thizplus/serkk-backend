package main

import (
	"fmt"
	"log"

	"gofiber-template/domain/models"
	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Testing migrations step by step...\n")

	models := []struct {
		name  string
		model interface{}
	}{
		{"User", &models.User{}},
		{"Post", &models.Post{}},
		{"Comment", &models.Comment{}},
		{"Media", &models.Media{}},
		{"PostMedia", &models.PostMedia{}},
		{"Vote", &models.Vote{}},
		{"Follow", &models.Follow{}},
		{"SavedPost", &models.SavedPost{}},
		{"Notification", &models.Notification{}},
		{"NotificationSettings", &models.NotificationSettings{}},
		{"PushSubscription", &models.PushSubscription{}},
		{"Tag", &models.Tag{}},
		{"SearchHistory", &models.SearchHistory{}},
		{"Conversation", &models.Conversation{}},
		{"Block", &models.Block{}},
		{"Message", &models.Message{}},
		{"Task", &models.Task{}},
		{"File", &models.File{}},
		{"Job", &models.Job{}},
	}

	for i, m := range models {
		fmt.Printf("[%d/%d] Migrating %s...", i+1, len(models), m.name)
		err := db.AutoMigrate(m.model)
		if err != nil {
			fmt.Printf(" ❌ FAILED\n")
			log.Fatalf("Error: %v", err)
		}
		fmt.Printf(" ✓\n")
	}

	fmt.Println("\n✅ All migrations successful!")
}
