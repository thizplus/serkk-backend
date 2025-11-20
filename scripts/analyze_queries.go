package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gofiber-template/domain/repositories"
	"gofiber-template/infrastructure/postgres"
	"gofiber-template/pkg/config"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type QueryLogger struct {
	queries []string
	count   int
}

func (ql *QueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	return ql
}

func (ql *QueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
}

func (ql *QueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
}

func (ql *QueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
}

func (ql *QueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc()
	ql.count++
	ql.queries = append(ql.queries, sql)
	fmt.Printf("\n📊 Query #%d:\n%s\n", ql.count, sql)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
	)

	queryLogger := &QueryLogger{
		queries: make([]string, 0),
		count:   0,
	}

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		Logger: queryLogger,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := postgres.NewPostRepository(db)
	ctx := context.Background()

	fmt.Println("🔍 Analyzing queries for List() method (20 posts, sorted by hot)")
	fmt.Println("=" + string(make([]byte, 80)))

	start := time.Now()
	posts, err := repo.List(ctx, 0, 20, repositories.SortByHot)
	duration := time.Since(start)

	if err != nil {
		log.Fatalf("List failed: %v", err)
	}

	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Printf("\n📈 Summary:\n")
	fmt.Printf("   Posts returned: %d\n", len(posts))
	fmt.Printf("   Total queries: %d\n", queryLogger.count)
	fmt.Printf("   Total time: %v\n", duration)
	fmt.Printf("   Avg time per query: %v\n", duration/time.Duration(queryLogger.count))

	if len(posts) > 0 {
		fmt.Printf("\n   Sample post:\n")
		fmt.Printf("   - Title: %s\n", posts[0].Title)
		fmt.Printf("   - Author: %s\n", posts[0].Author.Username)
		fmt.Printf("   - Media count: %d\n", len(posts[0].Media))
		fmt.Printf("   - Tags count: %d\n", len(posts[0].Tags))
	}

	fmt.Println("\n" + string(make([]byte, 80)))
	fmt.Printf("\n⚠️  Performance Issue:\n")
	fmt.Printf("   - %d queries for %d posts = %.1f queries per post\n", queryLogger.count, len(posts), float64(queryLogger.count)/float64(len(posts)))
	fmt.Printf("   - For 1000 req/sec = %d queries/sec to database\n", queryLogger.count*1000)
	fmt.Println("\n💡 Recommendation: Reduce to 4 queries using batch loading")
}
