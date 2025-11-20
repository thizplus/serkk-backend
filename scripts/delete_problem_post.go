package main

import (
	"fmt"
	"log"

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
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		"disable",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	postID := "6d678ddb-2822-441a-8b71-a6de8918dd19"

	fmt.Printf("Deleting post %s...\n", postID[:8])

	result := db.Exec("DELETE FROM posts WHERE id = ?", postID)
	if result.Error != nil {
		log.Fatalf("Failed to delete post: %v", result.Error)
	}

	fmt.Printf("✅ Deleted post (rows affected: %d)\n", result.RowsAffected)
}
