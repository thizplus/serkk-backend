package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	log.Println("🗑️  Deleting NATS consumers...")

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

	// Delete all consumers on USER_EVENTS stream
	streamName := "USER_EVENTS"

	// List all consumers
	consumers := js.ConsumerNames(streamName)

	consumerList := make([]string, 0)
	for name := range consumers {
		consumerList = append(consumerList, name)
	}

	if len(consumerList) == 0 {
		log.Println("✅ No consumers found on stream:", streamName)
		return
	}

	log.Printf("📋 Found %d consumers on stream %s:", len(consumerList), streamName)
	for _, name := range consumerList {
		log.Printf("   - %s", name)
	}

	// Delete each consumer
	for _, name := range consumerList {
		err := js.DeleteConsumer(streamName, name)
		if err != nil {
			log.Printf("❌ Failed to delete consumer %s: %v", name, err)
		} else {
			log.Printf("✅ Deleted consumer: %s", name)
		}
	}

	log.Println("🎉 Consumer cleanup completed!")
}
