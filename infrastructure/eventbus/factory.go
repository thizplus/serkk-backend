package eventbus

import (
	"fmt"

	"gofiber-template/domain/ports"
	"gofiber-template/infrastructure/eventbus/nats"
)

// EventBusType represents the type of event bus
type EventBusType string

const (
	EventBusTypeNATS  EventBusType = "nats"
	EventBusTypeKafka EventBusType = "kafka"
)

// EventBusConfig contains configuration for event bus
type EventBusConfig struct {
	Type    EventBusType
	NATSURL string   // For NATS
	Brokers []string // For Kafka (future)
}

// NewEventPublisher creates a new event publisher based on configuration
func NewEventPublisher(config EventBusConfig) (ports.EventPublisher, error) {
	switch config.Type {
	case EventBusTypeNATS:
		return nats.NewNATSPublisher(config.NATSURL)
	case EventBusTypeKafka:
		// TODO: Implement Kafka publisher
		return nil, fmt.Errorf("kafka publisher not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported event bus type: %s", config.Type)
	}
}

// NewEventSubscriber creates a new event subscriber based on configuration
func NewEventSubscriber(config EventBusConfig) (ports.EventSubscriber, error) {
	switch config.Type {
	case EventBusTypeNATS:
		return nats.NewNATSSubscriber(config.NATSURL)
	case EventBusTypeKafka:
		// TODO: Implement Kafka subscriber
		return nil, fmt.Errorf("kafka subscriber not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported event bus type: %s", config.Type)
	}
}
