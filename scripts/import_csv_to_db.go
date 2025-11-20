package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gofiber-template/domain/models"
	"gofiber-template/infrastructure/postgres"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Get CSV filename
	csvFile := "suekk_720_posts.csv"
	if len(os.Args) > 1 {
		csvFile = os.Args[1]
	}

	log.Printf("📁 Reading CSV file: %s", csvFile)

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

	// Get Bot User ID from env
	botUserIDStr := os.Getenv("AUTO_POST_BOT_USER_ID")
	if botUserIDStr == "" {
		log.Fatal("❌ AUTO_POST_BOT_USER_ID not set in .env")
	}

	botUserID, err := uuid.Parse(botUserIDStr)
	if err != nil {
		log.Fatalf("❌ Invalid BOT_USER_ID: %v", err)
	}

	log.Printf("🤖 Bot User ID: %s", botUserID)

	// Group topics by category and tone
	grouped := groupTopics(topics)

	log.Printf("📊 Grouped into %d settings", len(grouped))

	// Create settings
	ctx := context.Background()
	successCount := 0
	failCount := 0

	for key, topicsList := range grouped {
		category := key.category
		tone := key.tone

		// Create chunks of 50 topics (max per setting)
		chunks := chunkTopics(topicsList, 50)

		for i, chunk := range chunks {
			settingName := fmt.Sprintf("%s_%s_%d", category, tone, i+1)

			// Convert topics to JSON
			topicsJSON, err := json.Marshal(chunk)
			if err != nil {
				log.Printf("❌ Failed to marshal topics for %s: %v", settingName, err)
				failCount++
				continue
			}

			// Create setting
			setting := &models.AutoPostSetting{
				BotUserID:        botUserID,
				IsEnabled:        false, // เริ่มต้นปิดไว้
				CronSchedule:     "0 * * * *",
				Model:            "gpt-4o-mini",
				Tone:             tone,
				EnableVariations: true,
				Topics:           topicsJSON,
				MaxTokens:        1500,
				Temperature:      "0.8",
			}

			if err := db.WithContext(ctx).Create(setting).Error; err != nil {
				log.Printf("❌ Failed to create setting %s: %v", settingName, err)
				failCount++
			} else {
				log.Printf("✅ Created: %s (%d topics)", settingName, len(chunk))
				successCount++
			}
		}
	}

	log.Println("=" * 60)
	log.Printf("📊 Import Summary:")
	log.Printf("  ✅ Success: %d settings", successCount)
	log.Printf("  ❌ Failed: %d settings", failCount)
	log.Printf("  📝 Total topics: %d", len(topics))
	log.Println("=" * 60)
	log.Println("")
	log.Println("🎯 Next Steps:")
	log.Println("  1. Review settings: SELECT * FROM auto_post_settings;")
	log.Println("  2. Test one setting: UPDATE auto_post_settings SET is_enabled = true WHERE id = '...';")
	log.Println("  3. Restart server to activate scheduler")
}

type CSVTopic struct {
	Category string
	Topic    string
	Tone     string
}

type groupKey struct {
	category string
	tone     string
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

	// Skip header row
	var topics []CSVTopic
	for i, record := range records {
		if i == 0 {
			// Skip header
			continue
		}

		if len(record) < 3 {
			log.Printf("⚠️  Skipping row %d: invalid format", i+1)
			continue
		}

		topic := CSVTopic{
			Category: record[0],
			Topic:    record[1],
			Tone:     record[2],
		}

		// Skip empty topics
		if topic.Topic == "" {
			continue
		}

		// Default tone if empty
		if topic.Tone == "" {
			topic.Tone = "neutral"
		}

		// Default category if empty
		if topic.Category == "" {
			topic.Category = "general"
		}

		topics = append(topics, topic)
	}

	return topics, nil
}

func groupTopics(topics []CSVTopic) map[groupKey][]string {
	grouped := make(map[groupKey][]string)

	for _, topic := range topics {
		key := groupKey{
			category: topic.Category,
			tone:     topic.Tone,
		}

		grouped[key] = append(grouped[key], topic.Topic)
	}

	return grouped
}

func chunkTopics(topics []string, size int) [][]string {
	var chunks [][]string

	for i := 0; i < len(topics); i += size {
		end := i + size
		if end > len(topics) {
			end = len(topics)
		}

		chunks = append(chunks, topics[i:end])
	}

	return chunks
}
