package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"gofiber-template/domain/events"
	"gofiber-template/domain/ports"
	"gofiber-template/infrastructure/eventbus"
)

func main() {
	log.Println("🧪 Testing User Event Publishing...")

	// Create event bus configuration
	config := eventbus.EventBusConfig{
		Type:    eventbus.EventBusTypeNATS,
		NATSURL: "nats://localhost:4222",
	}

	// Create publisher
	publisher, err := eventbus.NewEventPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	log.Println("✅ Publisher connected to NATS")

	// Test 1: Publish user.created event
	log.Println("\n--- Test 1: Publishing user.created event ---")
	publishUserCreated(publisher)

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Test 2: Publish user.updated event
	log.Println("\n--- Test 2: Publishing user.updated event ---")
	publishUserUpdated(publisher)

	// Wait a bit
	time.Sleep(2 * time.Second)

	// Test 3: Publish user.deleted event
	log.Println("\n--- Test 3: Publishing user.deleted event ---")
	publishUserDeleted(publisher)

	log.Println("\n🎉 All test events published successfully!")
	log.Println("📝 Check your monolith logs to see if events were consumed")
}

func publishUserCreated(publisher ports.EventPublisher) {
	event := events.UserCreatedEvent{
		BaseEvent: events.BaseEvent{
			EventID:         uuid.New().String(),
			EventType:       "user.created",
			EventVersion:    "v1",
			Timestamp:       time.Now(),
			ProducerService: "test-script",
			SequenceNum:     1,
		},
		Data: events.UserCreatedData{
			UserID:      uuid.New().String(),
			Email:       "testuser@example.com",
			Username:    "testuser",
			DisplayName: "Test User",
			Avatar:      "https://example.com/avatar.jpg",
			Role:        "user",
			IsActive:    true,
		},
	}

	err := publisher.Publish(context.Background(), "AUTH_EVENTS.user.created", event)
	if err != nil {
		log.Fatalf("Failed to publish user.created: %v", err)
	}

	log.Printf("✅ Published user.created event (UserID: %s, Email: %s, EventID: %s)",
		event.Data.UserID, event.Data.Email, event.EventID)
}

func publishUserUpdated(publisher ports.EventPublisher) {
	newEmail := "updated@example.com"
	newDisplayName := "Updated User"

	event := events.UserUpdatedEvent{
		BaseEvent: events.BaseEvent{
			EventID:         uuid.New().String(),
			EventType:       "user.updated",
			EventVersion:    "v1",
			Timestamp:       time.Now(),
			ProducerService: "test-script",
			SequenceNum:     2,
		},
		Data: events.UserUpdatedData{
			UserID:      uuid.New().String(), // Generate valid UUID
			Email:       &newEmail,
			DisplayName: &newDisplayName,
		},
	}

	err := publisher.Publish(context.Background(), "AUTH_EVENTS.user.updated", event)
	if err != nil {
		log.Fatalf("Failed to publish user.updated: %v", err)
	}

	log.Printf("✅ Published user.updated event (UserID: %s, EventID: %s)",
		event.Data.UserID, event.EventID)
}

func publishUserDeleted(publisher ports.EventPublisher) {
	event := events.UserDeletedEvent{
		BaseEvent: events.BaseEvent{
			EventID:         uuid.New().String(),
			EventType:       "user.deleted",
			EventVersion:    "v1",
			Timestamp:       time.Now(),
			ProducerService: "test-script",
			SequenceNum:     3,
		},
		Data: events.UserDeletedData{
			UserID: uuid.New().String(), // Generate valid UUID
		},
	}

	err := publisher.Publish(context.Background(), "AUTH_EVENTS.user.deleted", event)
	if err != nil {
		log.Fatalf("Failed to publish user.deleted: %v", err)
	}

	log.Printf("✅ Published user.deleted event (UserID: %s, EventID: %s)",
		event.Data.UserID, event.EventID)
}
