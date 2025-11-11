package config

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetTestDB returns test database connection
func GetTestDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		"localhost",
		"postgres",
		"postgres",
		"gofiber_test", // Separate test database
		"5432",
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Silent in tests
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

// CleanTestDB cleans all tables
func CleanTestDB(db *gorm.DB) error {
	// Disable foreign key checks temporarily
	db.Exec("SET CONSTRAINTS ALL DEFERRED")

	// Get all tables
	tables := []string{
		"messages",
		"conversations",
		"conversation_participants",
		"notifications",
		"push_subscriptions",
		"comments",
		"votes",
		"saved_posts",
		"post_tags",
		"tags",
		"post_media",
		"media",
		"posts",
		"follows",
		"blocks",
		"search_histories",
		"users",
		"jobs",
		"tasks",
	}

	// Truncate all tables
	for _, table := range tables {
		db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}

	return nil
}
