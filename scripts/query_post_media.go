package main

import (
	"fmt"
	"log"

	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostMedia struct {
	PostID       string `gorm:"column:post_id"`
	MediaID      string `gorm:"column:media_id"`
	DisplayOrder int    `gorm:"column:display_order"`
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

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

	postID := "207af412-c890-4560-83b2-a5642933367e"

	var results []PostMedia
	err = db.Table("post_media").
		Where("post_id = ?", postID).
		Order("display_order ASC").
		Find(&results).Error

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	fmt.Printf("📊 Query Result for Post: %s\n", postID[:8]+"...")
	fmt.Println(string(make([]byte, 80)))
	fmt.Printf("\nTotal records: %d\n\n", len(results))

	if len(results) == 0 {
		fmt.Println("❌ No records found")
		return
	}

	fmt.Println("Records:")
	for i, record := range results {
		fmt.Printf("%d. post_id: %s\n", i+1, record.PostID[:8]+"...")
		fmt.Printf("   media_id: %s\n", record.MediaID[:8]+"...")
		fmt.Printf("   display_order: %d\n\n", record.DisplayOrder)
	}

	fmt.Println(string(make([]byte, 80)))
	fmt.Printf("\n✅ Database has %d media records for this post\n", len(results))
}
