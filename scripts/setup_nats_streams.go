package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	log.Println("🚀 Setting up NATS JetStream streams...")

	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	log.Println("✅ Connected to NATS")

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Create USER_EVENTS stream (Auth Service V2)
	createUserEventsStream(js)

	// Create SOCIAL_EVENTS stream (for future use)
	createSocialEventsStream(js)

	// Create NOTIFICATION_EVENTS stream (for future use)
	createNotificationEventsStream(js)

	log.Println("🎉 All streams created successfully!")
}

func createUserEventsStream(js nats.JetStreamContext) {
	streamName := "USER_EVENTS"

	// Check if stream already exists
	stream, err := js.StreamInfo(streamName)
	if err == nil {
		log.Printf("⚠️  Stream %s already exists (subjects: %v)", streamName, stream.Config.Subjects)
		log.Printf("✅ Stream: %s ready", streamName)
		printStreamInfo(js, streamName)
		return
	}

	// Create new stream
	// ตาม INTEGRATION_GUIDE.md และ EVENT_SCHEMA_V2.md จาก Auth Service
	_, err = js.AddStream(&nats.StreamConfig{
		Name:        streamName,
		Subjects:    []string{"user.events.*"}, // user.events.created, user.events.updated, user.events.deleted
		Description: "Stream for user identity events from Auth Service (V2 - Minimal Identity Event)",
		Storage:     nats.FileStorage,       // Persist to disk
		Retention:   nats.WorkQueuePolicy,   // ลบหลัง ack (ตาม Auth Service guide)
		MaxAge:      30 * 24 * time.Hour,    // Keep messages for 30 days
		MaxBytes:    10 * 1024 * 1024 * 1024, // Max 10 GB
		Duplicates:  2 * time.Minute,        // Prevent duplicate messages
		NoAck:       false,                   // Require acknowledgment
		Discard:     nats.DiscardOld,        // Discard old messages when limits reached
	})

	if err != nil {
		log.Fatalf("❌ Failed to create stream %s: %v", streamName, err)
	}

	log.Printf("✅ Created stream: %s", streamName)
	printStreamInfo(js, streamName)
}

func createSocialEventsStream(js nats.JetStreamContext) {
	streamName := "SOCIAL_EVENTS"

	// Check if stream already exists
	_, err := js.StreamInfo(streamName)
	if err == nil {
		log.Printf("⚠️  Stream %s already exists", streamName)
		return
	}

	// Create new stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:        streamName,
		Subjects:    []string{"SOCIAL_EVENTS.>"}, // post.created, comment.created, etc.
		Description: "Stream for social media events (posts, comments, votes, follows)",
		Storage:     nats.FileStorage,
		Retention:   nats.LimitsPolicy,
		MaxAge:      30 * 24 * time.Hour,  // 30 days
		MaxBytes:    50 * 1024 * 1024 * 1024, // 50 GB (more data expected)
		Duplicates:  2 * time.Minute,
		NoAck:       false,
		Discard:     nats.DiscardOld,
	})

	if err != nil {
		log.Fatalf("❌ Failed to create stream %s: %v", streamName, err)
	}

	log.Printf("✅ Created stream: %s", streamName)
	printStreamInfo(js, streamName)
}

func createNotificationEventsStream(js nats.JetStreamContext) {
	streamName := "NOTIFICATION_EVENTS"

	// Check if stream already exists
	_, err := js.StreamInfo(streamName)
	if err == nil {
		log.Printf("⚠️  Stream %s already exists", streamName)
		return
	}

	// Create new stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:        streamName,
		Subjects:    []string{"NOTIFICATION_EVENTS.>"}, // notification.created, notification.sent, etc.
		Description: "Stream for notification events",
		Storage:     nats.FileStorage,
		Retention:   nats.LimitsPolicy,
		MaxAge:      7 * 24 * time.Hour,  // 7 days (notifications are time-sensitive)
		MaxBytes:    5 * 1024 * 1024 * 1024, // 5 GB
		Duplicates:  1 * time.Minute,
		NoAck:       false,
		Discard:     nats.DiscardOld,
	})

	if err != nil {
		log.Fatalf("❌ Failed to create stream %s: %v", streamName, err)
	}

	log.Printf("✅ Created stream: %s", streamName)
	printStreamInfo(js, streamName)
}

func printStreamInfo(js nats.JetStreamContext, streamName string) {
	info, err := js.StreamInfo(streamName)
	if err != nil {
		log.Printf("⚠️  Could not get stream info: %v", err)
		return
	}

	fmt.Println("---")
	fmt.Printf("Stream: %s\n", info.Config.Name)
	fmt.Printf("  Subjects: %v\n", info.Config.Subjects)
	fmt.Printf("  Storage: %s\n", info.Config.Storage)
	fmt.Printf("  Max Age: %s\n", info.Config.MaxAge)
	fmt.Printf("  Max Bytes: %d MB\n", info.Config.MaxBytes/(1024*1024))
	fmt.Printf("  Messages: %d\n", info.State.Msgs)
	fmt.Printf("  Bytes: %d\n", info.State.Bytes)
	fmt.Println("---")
}
