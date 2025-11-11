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

	// Query all tables
	var tables []struct {
		TableName string `gorm:"column:tablename"`
	}

	err = db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public' ORDER BY tablename").Scan(&tables).Error
	if err != nil {
		log.Fatalf("Failed to query tables: %v", err)
	}

	fmt.Printf("Total tables: %d\n\n", len(tables))
	for i, table := range tables {
		fmt.Printf("%d. %s\n", i+1, table.TableName)
	}
}
