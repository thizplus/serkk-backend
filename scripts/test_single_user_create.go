package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// AuthServiceUserEvent - V2 Schema (Minimal Identity Event)
type AuthServiceUserEvent struct {
	// Minimal Identity Data
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Action   string `json:"action"` // "created" | "updated" | "deleted"

	// Observability Metadata
	RequestID   string `json:"request_id"`
	Timestamp   string `json:"timestamp"`    // ISO 8601
	ServiceName string `json:"service_name"` // "gofiber-auth"
	Sequence    uint64 `json:"sequence,omitempty"`
}

func main() {
	log.Println("🧪 Testing single user creation with unique data...")

	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream: %v", err)
	}

	log.Println("✅ Connected to NATS")

	// Generate unique user data
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")

	event := AuthServiceUserEvent{
		ID:          uniqueID,
		Email:       "unique_" + timestamp + "@example.com",
		Username:    "unique_user_" + timestamp,
		Action:      "created",
		RequestID:   uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
		ServiceName: "gofiber-auth",
	}

	data, _ := json.Marshal(event)
	_, err = js.Publish("user.events.created", data)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	log.Printf("✅ Published unique user.created event:")
	log.Printf("   UserID: %s", event.ID)
	log.Printf("   Email: %s", event.Email)
	log.Printf("   Username: %s", event.Username)
	log.Printf("   RequestID: %s", event.RequestID)

	log.Println("\n⏳ Waiting 2 seconds for processing...")
	time.Sleep(2 * time.Second)

	log.Println("\n✅ Done! Check server logs and database:")
	log.Printf("   SELECT * FROM users_identity WHERE id = '%s';", event.ID)
	log.Printf("   SELECT * FROM user_profiles WHERE user_id = '%s';", event.ID)
}
