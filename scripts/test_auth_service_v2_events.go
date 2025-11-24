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
	log.Println("🧪 Testing Auth Service V2 Events (Minimal Identity)...")

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

	// Test 1: user.events.created
	log.Println("\n--- Test 1: user.events.created (V2 - Minimal Identity) ---")
	publishCreatedEvent(js)
	time.Sleep(2 * time.Second)

	// Test 2: user.events.updated
	log.Println("\n--- Test 2: user.events.updated (V2 - Minimal Identity) ---")
	publishUpdatedEvent(js)
	time.Sleep(2 * time.Second)

	// Test 3: user.events.deleted
	log.Println("\n--- Test 3: user.events.deleted (V2 - Minimal Identity) ---")
	publishDeletedEvent(js)

	log.Println("\n🎉 All V2 test events published successfully!")
	log.Println("📝 Check your monolith logs to see if events were consumed")
	log.Println("📊 Check database:")
	log.Println("   SELECT * FROM users_identity;")
	log.Println("   SELECT * FROM user_profiles;")
}

func publishCreatedEvent(js nats.JetStreamContext) {
	event := AuthServiceUserEvent{
		ID:          uuid.New().String(),
		Email:       "testv2@example.com",
		Username:    "testuser_v2",
		Action:      "created",
		RequestID:   uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
		ServiceName: "gofiber-auth",
	}

	data, _ := json.Marshal(event)
	_, err := js.Publish("user.events.created", data)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	log.Printf("✅ Published V2 event:")
	log.Printf("   Subject: user.events.created")
	log.Printf("   UserID: %s", event.ID)
	log.Printf("   Email: %s", event.Email)
	log.Printf("   Username: %s", event.Username)
	log.Printf("   RequestID: %s", event.RequestID)
	log.Printf("   ⚠️  NOTE: No displayName, avatar, role, isActive (V2 Minimal Identity)")
}

func publishUpdatedEvent(js nats.JetStreamContext) {
	event := AuthServiceUserEvent{
		ID:          uuid.New().String(),
		Email:       "updated_v2@example.com",
		Username:    "updated_user_v2",
		Action:      "updated",
		RequestID:   uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
		ServiceName: "gofiber-auth",
	}

	data, _ := json.Marshal(event)
	_, err := js.Publish("user.events.updated", data)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	log.Printf("✅ Published V2 updated event:")
	log.Printf("   Subject: user.events.updated")
	log.Printf("   UserID: %s", event.ID)
	log.Printf("   Email: %s", event.Email)
	log.Printf("   Username: %s", event.Username)
	log.Printf("   RequestID: %s", event.RequestID)
}

func publishDeletedEvent(js nats.JetStreamContext) {
	event := AuthServiceUserEvent{
		ID:          uuid.New().String(),
		Email:       "deleted_v2@example.com",
		Username:    "deleted_user_v2",
		Action:      "deleted",
		RequestID:   uuid.New().String(),
		Timestamp:   time.Now().Format(time.RFC3339),
		ServiceName: "gofiber-auth",
	}

	data, _ := json.Marshal(event)
	_, err := js.Publish("user.events.deleted", data)
	if err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}

	log.Printf("✅ Published V2 deleted event:")
	log.Printf("   Subject: user.events.deleted")
	log.Printf("   UserID: %s", event.ID)
	log.Printf("   Username: %s", event.Username)
	log.Printf("   RequestID: %s", event.RequestID)
}
