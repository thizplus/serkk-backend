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

	sqlDB, _ := db.DB()

	fmt.Println("🔍 post_media table schema:")
	fmt.Println("=" + string(make([]byte, 60)) + "=\n")

	// Check table structure
	rows, err := sqlDB.Query(`
		SELECT
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_name = 'post_media'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatalf("Failed to get schema: %v", err)
	}
	defer rows.Close()

	fmt.Println("Columns:")
	for rows.Next() {
		var colName, dataType, nullable string
		var colDefault *string
		rows.Scan(&colName, &dataType, &nullable, &colDefault)
		defaultVal := "NULL"
		if colDefault != nil {
			defaultVal = *colDefault
		}
		fmt.Printf("  - %-20s %s-15s (nullable: %s-3s, default: %s)\n",
			colName, dataType, nullable, defaultVal)
	}

	// Check constraints
	rows2, err := sqlDB.Query(`
		SELECT
			tc.constraint_name,
			tc.constraint_type,
			kcu.column_name
		FROM information_schema.table_constraints tc
		LEFT JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_name = 'post_media'
		ORDER BY tc.constraint_type, tc.constraint_name
	`)
	if err != nil {
		log.Fatalf("Failed to get constraints: %v", err)
	}
	defer rows2.Close()

	fmt.Println("\nConstraints:")
	for rows2.Next() {
		var constraintName, constraintType string
		var columnName *string
		rows2.Scan(&constraintName, &constraintType, &columnName)
		col := "N/A"
		if columnName != nil {
			col = *columnName
		}
		fmt.Printf("  - %-30s (%-15s on %s)\n", constraintName, constraintType, col)
	}

	// Check indexes
	rows3, err := sqlDB.Query(`
		SELECT
			indexname,
			indexdef
		FROM pg_indexes
		WHERE tablename = 'post_media'
		ORDER BY indexname
	`)
	if err != nil {
		log.Fatalf("Failed to get indexes: %v", err)
	}
	defer rows3.Close()

	fmt.Println("\nIndexes:")
	for rows3.Next() {
		var indexName, indexDef string
		rows3.Scan(&indexName, &indexDef)
		fmt.Printf("  - %s\n", indexName)
		fmt.Printf("    %s\n", indexDef)
	}

	// Check for a specific post
	postID := "6d678ddb-2822-441a-8b71-a6de8918dd19"

	type Result struct {
		PostID       string `gorm:"column:post_id"`
		MediaID      string `gorm:"column:media_id"`
		DisplayOrder int    `gorm:"column:display_order"`
	}

	var results []Result
	err = db.Raw(`
		SELECT
			post_id::text,
			media_id::text,
			display_order
		FROM post_media
		WHERE post_id = ?
		ORDER BY display_order, media_id
	`, postID).Scan(&results).Error

	if err != nil {
		log.Fatalf("Failed to check post: %v", err)
	}

	fmt.Printf("\n\nPost %s has %d records in post_media:\n", postID[:8], len(results))
	for i, r := range results {
		fmt.Printf("  %2d. Media: %s... | Order: %d\n", i+1, r.MediaID[:8], r.DisplayOrder)
	}
}
