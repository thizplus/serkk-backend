package main

import (
	"fmt"
	"log"

	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type IndexInfo struct {
	TableName string `gorm:"column:tablename"`
	IndexName string `gorm:"column:indexname"`
	IndexDef  string `gorm:"column:indexdef"`
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔍 Checking Performance Indexes")
	fmt.Println(string(make([]byte, 80)))

	// Check critical indexes for feed queries
	criticalIndexes := []string{
		"idx_posts_hot_score",
		"idx_posts_created_status",
		"idx_posts_votes_status",
		"idx_post_media_post_order",
		"idx_post_tags_post",
		"idx_posts_published",
	}

	var results []IndexInfo
	query := `
		SELECT
			tablename,
			indexname,
			indexdef
		FROM pg_indexes
		WHERE schemaname = 'public'
		AND tablename IN ('posts', 'post_media', 'post_tags')
		ORDER BY tablename, indexname
	`

	err = db.Raw(query).Scan(&results).Error
	if err != nil {
		log.Fatalf("Failed to query indexes: %v", err)
	}

	// Create map of existing indexes
	existingIndexes := make(map[string]bool)
	for _, result := range results {
		existingIndexes[result.IndexName] = true
	}

	fmt.Println("\n📊 Critical Indexes Status:\n")

	allExist := true
	for _, indexName := range criticalIndexes {
		if existingIndexes[indexName] {
			fmt.Printf("   ✅ %s\n", indexName)
		} else {
			fmt.Printf("   ❌ %s (MISSING)\n", indexName)
			allExist = false
		}
	}

	fmt.Println("\n" + string(make([]byte, 80)))

	if allExist {
		fmt.Println("\n✅ All critical indexes exist!")
		fmt.Println("\n📈 Next step: Run performance test to measure improvement")
	} else {
		fmt.Println("\n⚠️  Some indexes are missing!")
		fmt.Println("\n💡 Action: Run migration 014_add_performance_indexes.sql")
	}

	// Show all post-related indexes
	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Println("\n📋 All Post-related Indexes:\n")

	for _, result := range results {
		if result.TableName == "posts" || result.TableName == "post_media" || result.TableName == "post_tags" {
			fmt.Printf("Table: %s\n", result.TableName)
			fmt.Printf("Index: %s\n", result.IndexName)
			fmt.Printf("Definition: %s\n\n", result.IndexDef)
		}
	}
}
