package nats

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"gofiber-template/domain/ports"
)

// NATSSubscriber implements the EventSubscriber interface for NATS JetStream
type NATSSubscriber struct {
	js            nats.JetStreamContext
	nc            *nats.Conn
	subscriptions []*nats.Subscription
}

// NewNATSSubscriber creates a new NATS subscriber
func NewNATSSubscriber(natsURL string) (*NATSSubscriber, error) {
	// Connect to NATS
	nc, err := nats.Connect(natsURL,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
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

	return &NATSSubscriber{
		js:            js,
		nc:            nc,
		subscriptions: []*nats.Subscription{},
	}, nil
}

// Subscribe subscribes to a subject with a durable consumer
func (s *NATSSubscriber) Subscribe(
	ctx context.Context,
	subject string,
	consumerGroup string,
	handler ports.EventHandler,
) error {
	// Subscribe with durable consumer (consumer group)
	sub, err := s.js.QueueSubscribe(
		subject,
		consumerGroup,
		func(msg *nats.Msg) {
			// Convert NATS message to our EventMessage
			eventMsg := &ports.EventMessage{
				Subject: msg.Subject,
				Data:    msg.Data,
				Headers: convertNATSHeaders(msg.Header),
				Metadata: ports.MessageMetadata{
					MessageID:     getMessageID(msg),
					Timestamp:     time.Now().Unix(),
					DeliveryCount: getDeliveryCount(msg),
					StreamSeq:     getStreamSeq(msg),
				},
				Ack:  func() error { return msg.Ack() },
				Nack: func() error { return msg.Nak() },
				Term: func() error { return msg.Term() },
			}

			// Call the handler
			if err := handler(ctx, eventMsg); err != nil {
				// Handler failed, negative acknowledge (will retry)
				log.Printf("⚠️  Handler failed for subject %s: %v", msg.Subject, err)
				_ = msg.Nak()
			} else {
				// Handler succeeded, acknowledge
				_ = msg.Ack()
			}
		},
		nats.Durable(consumerGroup),        // Durable consumer (survives restarts)
		nats.ManualAck(),                   // Manual acknowledgment
		nats.AckExplicit(),                 // Explicit ack required
		nats.MaxDeliver(3),                 // Retry max 3 times
		nats.AckWait(30*time.Second),       // Wait 30 seconds for ack
		nats.DeliverAll(),                  // Deliver all messages (from beginning)
		nats.ReplayInstant(),               // Replay messages instantly
	)

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.subscriptions = append(s.subscriptions, sub)
	log.Printf("✅ Subscribed to subject: %s (consumer group: %s)", subject, consumerGroup)

	return nil
}

// Unsubscribe stops listening to a subject
func (s *NATSSubscriber) Unsubscribe(subject string) error {
	// Find and unsubscribe from matching subscriptions
	for i, sub := range s.subscriptions {
		if sub.Subject == subject {
			if err := sub.Unsubscribe(); err != nil {
				return err
			}
			// Remove from slice
			s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
			log.Printf("✅ Unsubscribed from subject: %s", subject)
		}
	}
	return nil
}

// Close closes all subscriptions
func (s *NATSSubscriber) Close() error {
	for _, sub := range s.subscriptions {
		_ = sub.Unsubscribe()
	}
	if s.nc != nil {
		s.nc.Close()
	}
	return nil
}

// Helper functions

func convertNATSHeaders(natsHeaders nats.Header) map[string]string {
	headers := make(map[string]string)
	for key := range natsHeaders {
		headers[key] = natsHeaders.Get(key)
	}
	return headers
}

func getMessageID(msg *nats.Msg) string {
	if msg.Header != nil {
		return msg.Header.Get("Nats-Msg-Id")
	}
	return ""
}

func getDeliveryCount(msg *nats.Msg) int {
	// NATS JetStream tracks delivery count internally
	// This is a simplified version
	meta, err := msg.Metadata()
	if err != nil {
		return 0
	}
	return int(meta.NumDelivered)
}

func getStreamSeq(msg *nats.Msg) uint64 {
	meta, err := msg.Metadata()
	if err != nil {
		return 0
	}
	return meta.Sequence.Stream
}
