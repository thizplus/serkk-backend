package main

import (
	"fmt"
	"log"

	"gofiber-template/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("Connected to database")

	// Drop search_history table (might exist with different schema)
	db.Exec("DROP TABLE IF EXISTS search_history CASCADE")
	fmt.Println("✓ Dropped search_history table")

	// Also drop search_histories if exists (plural form)
	db.Exec("DROP TABLE IF EXISTS search_histories CASCADE")
	fmt.Println("✓ Dropped search_histories table")

	fmt.Println("\nDone! Run the application again.")
}
