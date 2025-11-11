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

	// Read and execute 001_initial_schema.sql
	migrationSQL, err := os.ReadFile("migrations/001_initial_schema.sql")
	if err != nil {
		// If migration file doesn't exist, file might not be found in some deployments
		// In that case, we can fallback to embedded migrations or skip
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	// Execute the migration
	_, err = sqlDB.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %v", err)
	}

	return nil
}
