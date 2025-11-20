package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gofiber-template/infrastructure/postgres"
	"gorm.io/gorm"
)

type QueueItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	BotUserID  uuid.UUID `gorm:"type:uuid;not null"`
	Topic      string    `gorm:"type:text;not null"`
	Tone       string    `gorm:"type:varchar(50);default:'neutral'"`
	Status     string    `gorm:"type:varchar(20);default:'pending'"`
}

func (QueueItem) TableName() string {
	return "auto_post_queue"
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  Warning: .env file not found")
	}

	// Get CSV filename
	csvFile := "suekk_720_posts.csv"
	if len(os.Args) > 1 {
		csvFile = os.Args[1]
	}

	log.Printf("📁 Reading CSV: %s", csvFile)

	// Read CSV
	topics, err := readCSV(csvFile)
	if err != nil {
		log.Fatalf("❌ Failed to read CSV: %v", err)
	}

	log.Printf("✅ Found %d topics", len(topics))

	// Connect to database
	dbConfig := postgres.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}

	db, err := postgres.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	log.Println("✅ Connected to database")

	// Get Bot User ID
	botUserIDStr := os.Getenv("AUTO_POST_BOT_USER_ID")
	if botUserIDStr == "" {
		log.Fatal("❌ AUTO_POST_BOT_USER_ID not set in .env")
	}

	botUserID, err := uuid.Parse(botUserIDStr)
	if err != nil {
		log.Fatalf("❌ Invalid BOT_USER_ID: %v", err)
	}

	log.Printf("🤖 Bot User ID: %s", botUserID)

	// Import topics
	ctx := context.Background()
	successCount := 0
	failCount := 0

	for i, topic := range topics {
		item := &QueueItem{
			BotUserID: botUserID,
			Topic:     topic.Topic,
			Tone:      topic.Tone,
			Status:    "pending",
		}

		if err := db.WithContext(ctx).Create(item).Error; err != nil {
			log.Printf("❌ [%d/%d] Failed: %s", i+1, len(topics), topic.Topic)
			failCount++
		} else {
			if (i+1)%50 == 0 {
				log.Printf("✅ [%d/%d] Imported...", i+1, len(topics))
			}
			successCount++
		}
	}

	log.Println("")
	log.Println("=" + repeatStr("=", 60))
	log.Printf("📊 Import Summary:")
	log.Printf("  ✅ Success: %d topics", successCount)
	log.Printf("  ❌ Failed: %d topics", failCount)
	log.Printf("  📝 Total: %d topics", len(topics))
	log.Println("=" + repeatStr("=", 60))
	log.Println("")
	log.Println("🎯 Next Steps:")
	log.Println("  1. Start server: ./bin/api")
	log.Println("  2. Scheduler will auto-post 1 topic per hour")
	log.Println("  3. Monitor: SELECT * FROM auto_post_queue ORDER BY created_at;")
}

type CSVTopic struct {
	Topic string
	Tone  string
}

func readCSV(filename string) ([]CSVTopic, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty CSV file")
	}

	var topics []CSVTopic
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}

		if len(record) < 2 {
			log.Printf("⚠️  Skipping row %d: invalid format", i+1)
			continue
		}

		// CSV format: category,topic,tone
		// หรือ: topic,tone
		var topic, tone string

		if len(record) >= 3 {
			// Format: category,topic,tone
			topic = record[1]
			tone = record[2]
		} else {
			// Format: topic,tone
			topic = record[0]
			tone = record[1]
		}

		// Skip empty topics
		if topic == "" {
			continue
		}

		// Default tone
		if tone == "" {
			tone = "neutral"
		}

		topics = append(topics, CSVTopic{
			Topic: topic,
			Tone:  tone,
		})
	}

	return topics, nil
}

func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
