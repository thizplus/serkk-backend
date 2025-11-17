package postgres

import (
	"context"
	"testing"

	"gofiber-template/domain/models"
	"gofiber-template/domain/repositories"
	"gofiber-template/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates a database connection for benchmarking
func setupBenchmarkDB(b *testing.B) *gorm.DB {
	cfg, err := config.LoadConfig()
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	dsn := config.GetPostgresDSN(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable logging for clean benchmarks
	})
	if err != nil {
		b.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

// BenchmarkList_Current benchmarks the current implementation with Preload
func BenchmarkList_Current(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := NewPostRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.List(ctx, 0, 20, repositories.SortByHot)
		if err != nil {
			b.Fatalf("List failed: %v", err)
		}
	}
}

// BenchmarkListByAuthor_Current benchmarks the current ListByAuthor implementation
func BenchmarkListByAuthor_Current(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := NewPostRepository(db)
	ctx := context.Background()

	// Get a real user ID from database
	var user models.User
	db.First(&user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ListByAuthor(ctx, user.ID, 0, 20)
		if err != nil {
			b.Fatalf("ListByAuthor failed: %v", err)
		}
	}
}

// BenchmarkListByTag_Current benchmarks the current ListByTag implementation
func BenchmarkListByTag_Current(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := NewPostRepository(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.ListByTag(ctx, "golang", 0, 20, repositories.SortByHot)
		if err != nil && err != gorm.ErrRecordNotFound {
			b.Fatalf("ListByTag failed: %v", err)
		}
	}
}

// Benchmark with query counting
func BenchmarkList_WithQueryCount(b *testing.B) {
	db := setupBenchmarkDB(b)

	// Enable query logging to count queries
	var queryCount int
	db.Logger = logger.New(
		&queryCounter{count: &queryCount},
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	repo := NewPostRepository(db)
	ctx := context.Background()

	queryCount = 0
	_, err := repo.List(ctx, 0, 20, repositories.SortByHot)
	if err != nil {
		b.Fatalf("List failed: %v", err)
	}

	b.Logf("Total queries executed: %d", queryCount)
}

// queryCounter implements logger.Writer to count queries
type queryCounter struct {
	count *int
}

func (qc *queryCounter) Printf(format string, args ...interface{}) {
	*qc.count++
}
