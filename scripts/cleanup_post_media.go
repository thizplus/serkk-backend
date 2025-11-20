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

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	fmt.Println("🔍 ขั้นตอนที่ 1: ตรวจสอบปัญหา post_media...")
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
		HAVING COUNT(*) > 10
		ORDER BY media_count DESC
		LIMIT 20
	`).Scan(&results).Error

	if err != nil {
		log.Fatalf("Failed to check media counts: %v", err)
	}

	if len(results) == 0 {
		fmt.Println("✅ ไม่พบ post ที่มี media เกิน 10 ไฟล์")
		fmt.Println("\n✨ ระบบปกติดี ไม่ต้องทำความสะอาด")
		return
	}

	fmt.Printf("\n⚠️  พบ %d posts ที่มี media เกิน 10 ไฟล์:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("  %d. Post ID: %s → %d media files\n", i+1, r.PostID, r.MediaCount)
	}

	// Get statistics
	type Stats struct {
		TotalPosts         int `gorm:"column:total_posts"`
		PostsWithTooMany   int `gorm:"column:posts_with_too_many"`
		TotalPostMedia     int `gorm:"column:total_post_media"`
		RecordsToBeDeleted int `gorm:"column:records_to_be_deleted"`
	}

	var stats Stats
	err = db.Raw(`
		WITH stats AS (
			SELECT
				COUNT(DISTINCT post_id) as total_posts,
				COUNT(*) as total_post_media
			FROM post_media
		),
		too_many AS (
			SELECT COUNT(*) as cnt
			FROM (
				SELECT post_id
				FROM post_media
				GROUP BY post_id
				HAVING COUNT(*) > 10
			) subq
		),
		to_delete AS (
			SELECT COUNT(*) as cnt
			FROM (
				SELECT
					post_id,
					media_id,
					ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
				FROM post_media
			) ranked
			WHERE rn > 10
		)
		SELECT
			stats.total_posts,
			too_many.cnt as posts_with_too_many,
			stats.total_post_media,
			to_delete.cnt as records_to_be_deleted
		FROM stats, too_many, to_delete
	`).Scan(&stats).Error

	if err != nil {
		log.Fatalf("Failed to get statistics: %v", err)
	}

	fmt.Printf("\n📊 สถิติ:\n")
	fmt.Printf("  - Total posts with media: %d\n", stats.TotalPosts)
	fmt.Printf("  - Posts with >10 media: %d\n", stats.PostsWithTooMany)
	fmt.Printf("  - Total post_media records: %d\n", stats.TotalPostMedia)
	fmt.Printf("  - Records to be deleted: %d\n\n", stats.RecordsToBeDeleted)

	// Ask for confirmation
	fmt.Print("❓ ต้องการดำเนินการลบ media ที่เกิน 10 ไฟล์ใช่หรือไม่? (yes/no): ")
	var answer string
	fmt.Scanln(&answer)

	if answer != "yes" && answer != "y" && answer != "Y" {
		fmt.Println("\n❌ ยกเลิกการดำเนินการ")
		return
	}

	fmt.Println("\n🧹 ขั้นตอนที่ 2: ทำความสะอาดข้อมูล...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	// Execute cleanup (steps 2-5 from SQL script)
	fmt.Println("\n1. สำรองข้อมูลก่อนลบ...")
	_, err = sqlDB.Exec(`
		CREATE TABLE IF NOT EXISTS post_media_backup AS
		SELECT * FROM post_media WHERE 1=0;

		INSERT INTO post_media_backup
		SELECT pm.*
		FROM post_media pm
		INNER JOIN (
			SELECT
				post_id,
				media_id,
				ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
			FROM post_media
		) ranked
		ON pm.post_id = ranked.post_id
		AND pm.media_id = ranked.media_id
		WHERE ranked.rn > 10;
	`)
	if err != nil {
		log.Fatalf("Failed to backup data: %v", err)
	}
	fmt.Println("   ✅ สำรองข้อมูลสำเร็จ")

	fmt.Println("\n2. ลบ media ที่เกิน 10 ตัว...")
	result, err := sqlDB.Exec(`
		DELETE FROM post_media
		WHERE (post_id, media_id) IN (
			SELECT post_id, media_id
			FROM (
				SELECT
					post_id,
					media_id,
					ROW_NUMBER() OVER (PARTITION BY post_id ORDER BY display_order, media_id) as rn
				FROM post_media
			) ranked
			WHERE rn > 10
		)
	`)
	if err != nil {
		log.Fatalf("Failed to delete excess media: %v", err)
	}

	rowsDeleted, _ := result.RowsAffected()
	fmt.Printf("   ✅ ลบข้อมูล %d records สำเร็จ\n", rowsDeleted)

	fmt.Println("\n3. จัดเรียง display_order ใหม่...")
	_, err = sqlDB.Exec(`
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
		log.Fatalf("Failed to reorder display_order: %v", err)
	}
	fmt.Println("   ✅ จัดเรียงลำดับสำเร็จ")

	// Verify results
	fmt.Println("\n✅ ขั้นตอนที่ 3: ตรวจสอบผลลัพธ์...")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	var postsWithTooMany int
	err = db.Raw(`
		SELECT COUNT(*)
		FROM (
			SELECT post_id
			FROM post_media
			GROUP BY post_id
			HAVING COUNT(*) > 10
		) subq
	`).Scan(&postsWithTooMany).Error

	if err != nil {
		log.Fatalf("Failed to verify: %v", err)
	}

	if postsWithTooMany == 0 {
		fmt.Println("\n🎉 สำเร็จ! ไม่มี post ที่มี media เกิน 10 ไฟล์แล้ว")
	} else {
		fmt.Printf("\n⚠️  ยังมี %d posts ที่มี media เกิน 10 ไฟล์อยู่\n", postsWithTooMany)
	}

	// Final statistics
	type FinalStats struct {
		Metric string `gorm:"column:metric"`
		Count  string `gorm:"column:count"`
	}

	var finalStats []FinalStats
	err = db.Raw(`
		SELECT
			'Total posts' as metric,
			COUNT(DISTINCT post_id)::text as count
		FROM post_media
		UNION ALL
		SELECT
			'Max media per post' as metric,
			MAX(cnt)::text as count
		FROM (
			SELECT COUNT(*) as cnt
			FROM post_media
			GROUP BY post_id
		) subq
		UNION ALL
		SELECT
			'Avg media per post' as metric,
			ROUND(AVG(cnt)::numeric, 2)::text as count
		FROM (
			SELECT COUNT(*) as cnt
			FROM post_media
			GROUP BY post_id
		) subq
	`).Scan(&finalStats).Error

	if err == nil {
		fmt.Println("\n📊 สถิติหลังทำความสะอาด:")
		for _, s := range finalStats {
			fmt.Printf("  - %s: %s\n", s.Metric, s.Count)
		}
	}

	fmt.Println("\n✨ เสร็จสิ้น!")
	fmt.Println("\n💡 หมายเหตุ:")
	fmt.Println("  - ข้อมูลที่ถูกลบได้สำรองไว้ในตาราง 'post_media_backup'")
	fmt.Println("  - กรุณา restart backend server เพื่อให้ trigger ทำงาน")
}
