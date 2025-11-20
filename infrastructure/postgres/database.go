package postgres

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewDatabase(config DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	// Run SQL migrations from files
	return RunSQLMigrations(db)
}

// RunSQLMigrations runs SQL migration files in order
func RunSQLMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %v", err)
	}

	// List of migration files to run in order
	migrationFiles := []string{
		"migrations/001_initial_schema.sql",
		"migrations/008_create_chat_tables.sql",
		"migrations/009_add_media_file_fields.sql",
		"migrations/010_add_video_streaming_fields.sql",
		"migrations/011_make_message_content_nullable.sql",
		"migrations/012_add_post_status_column.sql",
		"migrations/013_add_post_media_display_order.sql",
		"migrations/014_add_performance_indexes.sql",
		"migrations/015_add_post_type_column.sql",
		"migrations/016_add_client_post_id.sql",
		"migrations/017_add_post_media_limit_trigger.sql",
		"migrations/018_create_auto_post_tables.sql",
		"migrations/019_update_auto_post_tables_v2.sql",
		"migrations/020_create_simple_auto_post_queue.sql",
		"migrations/add_push_subscriptions_unique_constraint.sql",
	}

	// Execute each migration file
	for _, migrationFile := range migrationFiles {
		migrationSQL, err := os.ReadFile(migrationFile)
		if err != nil {
			// Skip if migration file doesn't exist (optional migrations)
			fmt.Printf("Warning: Could not read %s: %v\n", migrationFile, err)
			continue
		}

		// Execute the migration
		_, err = sqlDB.Exec(string(migrationSQL))
		if err != nil {
			// Log but continue (in case migration already applied)
			fmt.Printf("Warning: Failed to execute %s: %v\n", migrationFile, err)
			continue
		}

		fmt.Printf("✓ Applied migration: %s\n", migrationFile)
	}

	return nil
}
