package ports

import (
	"context"
)

// EventSubscriber is the interface for subscribing to events
type EventSubscriber interface {
	// Subscribe subscribes to a subject/topic with a handler
	// subject: topic/stream name (e.g., "AUTH_EVENTS.user.*")
	// consumerGroup: consumer group name for load balancing
	// handler: callback function to handle the event
	Subscribe(ctx context.Context, subject string, consumerGroup string, handler EventHandler) error

	// Unsubscribe stops listening to a subject
	Unsubscribe(subject string) error

	// Close closes all subscriptions
	Close() error
}

// EventHandler is the callback function for handling events
type EventHandler func(ctx context.Context, msg *EventMessage) error

// EventMessage represents a received event message
type EventMessage struct {
	Subject  string            // The subject/topic this message was received on
	Data     []byte            // Raw event data (JSON)
	Headers  map[string]string // Message headers
	Metadata MessageMetadata   // Message metadata

	// Ack acknowledges the message (processed successfully)
	Ack func() error

	// Nack negatively acknowledges the message (failed, will retry)
	Nack func() error

	// Term terminates the message (failed permanently, move to DLQ)
	Term func() error
}

// MessageMetadata contains metadata about the message
type MessageMetadata struct {
	MessageID     string
	Timestamp     int64
	DeliveryCount int    // How many times this message has been delivered
	StreamSeq     uint64 // Stream sequence number
}
