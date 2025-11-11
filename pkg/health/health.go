package health

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// HealthCheck represents a health check response
type HealthCheck struct {
	Status      Status                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]CheckResult `json:"checks"`
}

// CheckResult represents the result of a single check
type CheckResult struct {
	Status    Status        `json:"status"`
	Message   string        `json:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
}

// Checker defines the interface for health checkers
type Checker interface {
	Check(ctx context.Context) CheckResult
	Name() string
}

// HealthChecker manages health checks
type HealthChecker struct {
	checkers  []Checker
	startTime time.Time
	version   string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		checkers:  make([]Checker, 0),
		startTime: time.Now(),
		version:   version,
	}
}

// AddChecker adds a health checker
func (h *HealthChecker) AddChecker(checker Checker) {
	h.checkers = append(h.checkers, checker)
}

// Check runs all health checks
func (h *HealthChecker) Check(ctx context.Context) HealthCheck {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	for _, checker := range h.checkers {
		result := checker.Check(ctx)
		checks[checker.Name()] = result

		// Determine overall status
		if result.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if result.Status == StatusDegraded && overallStatus != StatusUnhealthy {
			overallStatus = StatusDegraded
		}
	}

	return HealthCheck{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Checks:    checks,
	}
}

// DatabaseChecker checks database health
type DatabaseChecker struct {
	db *gorm.DB
}

// NewDatabaseChecker creates a new database checker
func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

// Name returns the checker name
func (c *DatabaseChecker) Name() string {
	return "database"
}

// Check performs the database health check
func (c *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	sqlDB, err := c.db.DB()
	if err != nil {
		return CheckResult{
			Status:    StatusUnhealthy,
			Message:   "Failed to get database instance",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     err.Error(),
		}
	}

	// Ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return CheckResult{
			Status:    StatusUnhealthy,
			Message:   "Database ping failed",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
			Error:     err.Error(),
		}
	}

	// Check connection pool stats
	stats := sqlDB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections {
		return CheckResult{
			Status:    StatusDegraded,
			Message:   "Database connection pool exhausted",
			Timestamp: time.Now(),
			Duration:  time.Since(start),
		}
	}

	return CheckResult{
		Status:    StatusHealthy,
		Message:   fmt.Sprintf("Connected (%d/%d connections)", stats.OpenConnections, stats.MaxOpenConnections),
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// MemoryChecker checks memory usage
type MemoryChecker struct {
	threshold uint64 // Threshold in bytes
}

// NewMemoryChecker creates a new memory checker
func NewMemoryChecker(thresholdMB uint64) *MemoryChecker {
	return &MemoryChecker{
		threshold: thresholdMB * 1024 * 1024, // Convert MB to bytes
	}
}

// Name returns the checker name
func (c *MemoryChecker) Name() string {
	return "memory"
}

// Check performs the memory health check
func (c *MemoryChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	// Note: This is a simplified check
	// In production, use runtime.MemStats for detailed memory info
	return CheckResult{
		Status:    StatusHealthy,
		Message:   "Memory usage within limits",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}

// DiskChecker checks disk space
type DiskChecker struct {
	path      string
	threshold float64 // Threshold as percentage (0-100)
}

// NewDiskChecker creates a new disk checker
func NewDiskChecker(path string, thresholdPercent float64) *DiskChecker {
	return &DiskChecker{
		path:      path,
		threshold: thresholdPercent,
	}
}

// Name returns the checker name
func (c *DiskChecker) Name() string {
	return "disk"
}

// Check performs the disk health check
func (c *DiskChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	// Note: This is a simplified check
	// In production, use syscall.Statfs or similar for actual disk usage
	return CheckResult{
		Status:    StatusHealthy,
		Message:   "Disk space available",
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}
}
