package ports

import (
	"context"
)

// EventPublisher is the interface for publishing events
// This abstraction allows us to switch from NATS to Kafka without changing business logic
type EventPublisher interface {
	// Publish publishes an event to the event bus
	// subject: topic/stream name (e.g., "AUTH_EVENTS.user.created")
	// event: the event data (must implement Event interface)
	Publish(ctx context.Context, subject string, event interface{}) error

	// PublishBatch publishes multiple events atomically (if supported by underlying implementation)
	PublishBatch(ctx context.Context, events []PublishRequest) error

	// Close closes the publisher connection
	Close() error
}

// PublishRequest represents a single event to be published in a batch
type PublishRequest struct {
	Subject string
	Event   interface{}
}
