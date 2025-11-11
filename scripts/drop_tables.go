package main

import (
	"fmt"
	"log"
	"os"

	"gofiber-template/pkg/config"
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

	fmt.Println("Connected to database")

	// Drop all tables
	tables := []string{
		"post_tags",
		"post_media",
		"votes",
		"follows",
		"saved_posts",
		"notifications",
		"notification_settings",
		"push_subscriptions",
		"tags",
		"search_histories",
		"messages",
		"blocks",
		"conversations",
		"tasks",
		"files",
		"jobs",
		"comments",
		"posts",
		"media",
		"users",
	}

	for _, table := range tables {
		result := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if result.Error != nil {
			fmt.Printf("Warning: Failed to drop table %s: %v\n", table, result.Error)
		} else {
			fmt.Printf("✓ Dropped table: %s\n", table)
		}
	}

	fmt.Println("\nAll tables dropped successfully!")
	fmt.Println("Run the application again to recreate tables with correct schema")

	os.Exit(0)
}
