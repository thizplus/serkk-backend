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

	fmt.Println("🔍 ตรวจสอบจำนวน media ในแต่ละ post...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Check posts with too many media
	type MediaCount struct {
		PostID     string `gorm:"column:post_id"`
		MediaCount int    `gorm:"column:media_count"`
	}

	var results []MediaCount
	err = db.Raw(`
		SELECT
			post_id::text,
			COUNT(*)::int as media_count
		FROM post_media
		GROUP BY post_id
		ORDER BY media_count DESC
		LIMIT 20
	`).Scan(&results).Error

	if err != nil {
		log.Fatalf("Failed to check media counts: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("✅ ไม่มี post ที่มี media")
		return
	}

	fmt.Printf("\n📊 Top 20 posts ที่มี media มากที่สุด:\n\n")
	for i, r := range results {
		emoji := "✅"
		if r.MediaCount > 20 {
			emoji = "🔴"
		} else if r.MediaCount > 10 {
			emoji = "⚠️ "
		}
		fmt.Printf("%s %2d. Post ID: %s → %d media files\n", emoji, i+1, r.PostID, r.MediaCount)
	}

	// Get statistics
	type Stats struct {
		PostsOver20  int `gorm:"column:posts_over_20"`
		PostsOver10  int `gorm:"column:posts_over_10"`
		TotalPosts   int `gorm:"column:total_posts"`
		MaxMedia     int `gorm:"column:max_media"`
		RecordsOver10 int `gorm:"column:records_over_10"`
	}

	var stats Stats
	err = db.Raw(`
		SELECT
			COUNT(DISTINCT CASE WHEN cnt > 20 THEN post_id END)::int as posts_over_20,
			COUNT(DISTINCT CASE WHEN cnt > 10 THEN post_id END)::int as posts_over_10,
			COUNT(DISTINCT post_id)::int as total_posts,
			COALESCE(MAX(cnt), 0)::int as max_media,
			COALESCE(SUM(CASE WHEN rn > 10 THEN 1 ELSE 0 END), 0)::int as records_over_10
		FROM (
			SELECT
				post_id,
				COUNT(*) as cnt,
				ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
			FROM post_media
			GROUP BY post_id, media_id, display_order
		) subq
	`).Scan(&stats).Error

	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	fmt.Printf("\n📊 สถิติทั้งหมด:\n")
	fmt.Printf("  - Total posts: %d\n", stats.TotalPosts)
	fmt.Printf("  - Posts with >10 media: %d 🔴\n", stats.PostsOver10)
	fmt.Printf("  - Posts with >20 media: %d 🔴🔴\n", stats.PostsOver20)
	fmt.Printf("  - Max media per post: %d\n", stats.MaxMedia)
	fmt.Printf("  - Records to be deleted (>10): %d\n", stats.RecordsOver10)

	if stats.PostsOver10 > 0 {
		fmt.Printf("\n⚠️  ต้องการลบ media ที่เกิน 10 ไฟล์!\n")
	} else {
		fmt.Printf("\n✅ ระบบปกติดี\n")
	}
}
