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
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Check messages table columns
	var columns []struct {
		ColumnName string
		DataType   string
	}

	err = db.Raw(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = 'messages'
		ORDER BY ordinal_position
	`).Scan(&columns).Error

	if err != nil {
		log.Fatalf("Failed to query columns: %v", err)
	}

	fmt.Println("✓ Messages table schema:")
	fmt.Println("=" + string(make([]byte, 50)))
	for _, col := range columns {
		fmt.Printf("  %-20s %s\n", col.ColumnName, col.DataType)
	}
}
