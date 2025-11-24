package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"gofiber-template/domain/events"
	"gofiber-template/domain/ports"
)

// NATSPublisher implements the EventPublisher interface for NATS JetStream
type NATSPublisher struct {
	js nats.JetStreamContext
	nc *nats.Conn
}

// NewNATSPublisher creates a new NATS publisher
func NewNATSPublisher(natsURL string) (*NATSPublisher, error) {
	// Connect to NATS
	nc, err := nats.Connect(natsURL,
		nats.MaxReconnects(-1),           // Unlimited reconnects
		nats.ReconnectWait(2*time.Second), // Wait 2 seconds between reconnects
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &NATSPublisher{
		js: js,
		nc: nc,
	}, nil
}

// Publish publishes an event to NATS JetStream
func (p *NATSPublisher) Publish(ctx context.Context, subject string, event interface{}) error {
	// Ensure event has required fields
	if baseEvent, ok := event.(events.BaseEvent); ok {
		if baseEvent.EventID == "" {
			baseEvent.EventID = uuid.New().String()
		}
		if baseEvent.Timestamp.IsZero() {
			baseEvent.Timestamp = time.Now()
		}
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to NATS JetStream
	_, err = p.js.Publish(subject, data,
		nats.Context(ctx),
		nats.MsgId(p.getEventID(event)), // Deduplication based on EventID
	)
	if err != nil {
		return fmt.Errorf("failed to publish event to NATS: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple events in a batch
// Note: NATS doesn't have native batch publish, so we publish one by one
// TODO: Could use async publish for better performance
func (p *NATSPublisher) PublishBatch(ctx context.Context, requests []ports.PublishRequest) error {
	for _, req := range requests {
		if err := p.Publish(ctx, req.Subject, req.Event); err != nil {
			return fmt.Errorf("failed to publish event in batch: %w", err)
		}
	}
	return nil
}

// Close closes the NATS connection
func (p *NATSPublisher) Close() error {
	if p.nc != nil {
		p.nc.Close()
	}
	return nil
}

// getEventID extracts the event ID from the event (for deduplication)
func (p *NATSPublisher) getEventID(event interface{}) string {
	if baseEvent, ok := event.(events.BaseEvent); ok {
		return baseEvent.EventID
	}
	// Try to extract via interface if available
	type eventIDGetter interface {
		GetEventID() string
	}
	if getter, ok := event.(eventIDGetter); ok {
		return getter.GetEventID()
	}
	return uuid.New().String()
}
