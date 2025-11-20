package main

import (
	"context"
	"fmt"
	"log"

	"gofiber-template/domain/models"
	"gofiber-template/pkg/config"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()

	fmt.Println("🔧 Fixing display_order to be sequential...")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	// Fix display_order to be sequential (0, 1, 2, 3, ...)
	result, err := sqlDB.Exec(`
		UPDATE post_media
		SET display_order = subq.new_order - 1
		FROM (
			SELECT
				post_id,
				media_id,
				ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as new_order
			FROM post_media
		) subq
		WHERE post_media.post_id = subq.post_id
		AND post_media.media_id = subq.media_id
	`)
	if err != nil {
		log.Fatalf("Failed to fix display_order: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ Updated %d records\n\n", rowsAffected)

	// Test query to see if it fixes the issue
	postID, _ := uuid.Parse("6d678ddb-2822-441a-8b71-a6de8918dd19")

	fmt.Println("🔍 Testing preload after fix...")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	var post models.Post
	err = db.WithContext(context.Background()).
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.Order("post_media.display_order ASC")
		}).
		Where("id = ?", postID).
		First(&post).Error

	if err != nil {
		log.Fatalf("Failed to query post: %v", err)
	}

	fmt.Printf("Post ID: %s\n", post.ID)
	fmt.Printf("Media count: %d\n\n", len(post.Media))

	if len(post.Media) <= 20 {
		fmt.Println("Media files:")
		for i, media := range post.Media {
			fmt.Printf("  %2d. %s (Type: %s)\n", i+1, media.ID, media.Type)
		}
	}

	if len(post.Media) > 20 {
		fmt.Printf("🔴 ERROR: Still returning %d media files (expected 8)!\n", len(post.Media))
		fmt.Println("\nThis is a GORM bug. Let me check the actual SQL query...")
	} else {
		fmt.Printf("✅ SUCCESS: Returning %d media files\n", len(post.Media))
	}
}
