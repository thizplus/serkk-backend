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

	fmt.Println("🔍 ตรวจสอบ post_media duplicates...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Check specific post
	postID := "6d678ddb-2822-441a-8b71-a6de8918dd19"

	type MediaRecord struct {
		PostID       string `gorm:"column:post_id"`
		MediaID      string `gorm:"column:media_id"`
		DisplayOrder int    `gorm:"column:display_order"`
		Count        int    `gorm:"column:count"`
	}

	var records []MediaRecord
	err = db.Raw(`
		SELECT
			post_id::text,
			media_id::text,
			display_order,
			COUNT(*) as count
		FROM post_media
		WHERE post_id = ?
		GROUP BY post_id, media_id, display_order
		ORDER BY display_order
	`, postID).Scan(&records).Error

	if err != nil {
		log.Fatalf("Failed to check duplicates: %v", err)
	}

	fmt.Printf("\n📊 Post ID: %s\n", postID)
	fmt.Printf("Total unique combinations: %d\n\n", len(records))

	for i, r := range records {
		status := "✅"
		if r.Count > 1 {
			status = "🔴 DUPLICATE"
		}
		fmt.Printf("%s %2d. Media: %s | Order: %d | Count: %d\n",
			status, i+1, r.MediaID[:8], r.DisplayOrder, r.Count)
	}

	// Check all posts with duplicates
	type DuplicatePost struct {
		PostID           string `gorm:"column:post_id"`
		MediaID          string `gorm:"column:media_id"`
		DuplicateCount   int    `gorm:"column:duplicate_count"`
		TotalRecords     int    `gorm:"column:total_records"`
	}

	var duplicates []DuplicatePost
	err = db.Raw(`
		SELECT
			post_id::text,
			media_id::text,
			COUNT(*) as duplicate_count,
			SUM(COUNT(*)) OVER (PARTITION BY post_id) as total_records
		FROM post_media
		GROUP BY post_id, media_id
		HAVING COUNT(*) > 1
		ORDER BY duplicate_count DESC
		LIMIT 20
	`).Scan(&duplicates).Error

	if err != nil {
		log.Fatalf("Failed to check all duplicates: %v", err)
	}

	if len(duplicates) > 0 {
		fmt.Printf("\n\n🔴 พบ posts ที่มี duplicate media:\n\n")
		for i, d := range duplicates {
			fmt.Printf("  %2d. Post: %s | Media: %s | Duplicates: %dx (Total: %d records)\n",
				i+1, d.PostID[:8], d.MediaID[:8], d.DuplicateCount, d.TotalRecords)
		}

		// Count total duplicate records
		type Stats struct {
			TotalDuplicates int `gorm:"column:total_duplicates"`
			PostsAffected   int `gorm:"column:posts_affected"`
		}

		var stats Stats
		err = db.Raw(`
			WITH duplicates AS (
				SELECT post_id, media_id, COUNT(*) - 1 as dup_count
				FROM post_media
				GROUP BY post_id, media_id
				HAVING COUNT(*) > 1
			)
			SELECT
				COALESCE(SUM(dup_count), 0)::int as total_duplicates,
				COUNT(DISTINCT post_id)::int as posts_affected
			FROM duplicates
		`).Scan(&stats).Error

		if err == nil {
			fmt.Printf("\n📊 สถิติ duplicates:\n")
			fmt.Printf("  - Posts affected: %d\n", stats.PostsAffected)
			fmt.Printf("  - Total duplicate records to delete: %d\n", stats.TotalDuplicates)
		}
	} else {
		fmt.Printf("\n✅ ไม่พบ duplicate records\n")
	}
}
