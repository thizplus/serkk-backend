package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	// MaxIdleConns sets the maximum number of connections in the idle connection pool
	MaxIdleConns int

	// MaxOpenConns sets the maximum number of open connections to the database
	MaxOpenConns int

	// ConnMaxLifetime sets the maximum amount of time a connection may be reused
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime sets the maximum amount of time a connection may be idle
	ConnMaxIdleTime time.Duration
}

// DefaultPoolConfig returns default connection pool configuration
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// ProductionPoolConfig returns production-optimized pool configuration
func ProductionPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:    25,
		MaxOpenConns:    100,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// DevelopmentPoolConfig returns development pool configuration
func DevelopmentPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:    5,
		MaxOpenConns:    25,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}
}

// ConfigureConnectionPool configures the database connection pool
func ConfigureConnectionPool(db *gorm.DB, config PoolConfig) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)

	// Set maximum number of open connections
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)

	// Set maximum lifetime for a connection
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	// Set maximum idle time for a connection
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	return nil
}

// GetPoolStats returns current connection pool statistics
func GetPoolStats(db *gorm.DB) (map[string]interface{}, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	stats := sqlDB.Stats()

	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                   stats.InUse,
		"idle":                     stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}, nil
}

// HealthCheck performs a database health check
func HealthCheck(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// ConnectionPoolMonitor provides methods to monitor connection pool health
type ConnectionPoolMonitor struct {
	db *gorm.DB
}

// NewConnectionPoolMonitor creates a new connection pool monitor
func NewConnectionPoolMonitor(db *gorm.DB) *ConnectionPoolMonitor {
	return &ConnectionPoolMonitor{db: db}
}

// GetStats returns current pool statistics
func (m *ConnectionPoolMonitor) GetStats() (map[string]interface{}, error) {
	return GetPoolStats(m.db)
}

// IsHealthy checks if the connection pool is healthy
func (m *ConnectionPoolMonitor) IsHealthy() bool {
	stats, err := GetPoolStats(m.db)
	if err != nil {
		return false
	}

	// Check if we have too many connections in use
	inUse := stats["in_use"].(int)
	maxOpen := stats["max_open_connections"].(int)

	// Warn if more than 80% of connections are in use
	if float64(inUse)/float64(maxOpen) > 0.8 {
		return false
	}

	return true
}

// GetUtilization returns the connection pool utilization percentage
func (m *ConnectionPoolMonitor) GetUtilization() (float64, error) {
	stats, err := GetPoolStats(m.db)
	if err != nil {
		return 0, err
	}

	inUse := stats["in_use"].(int)
	maxOpen := stats["max_open_connections"].(int)

	if maxOpen == 0 {
		return 0, nil
	}

	return (float64(inUse) / float64(maxOpen)) * 100, nil
}
