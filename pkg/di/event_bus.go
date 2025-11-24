package di

import (
	"context"
	"gofiber-template/application/eventhandlers"
	"gofiber-template/infrastructure/eventbus"
	"log"
)

// SetupEventBus initializes the event bus (publisher and subscriber)
func (c *Container) SetupEventBus() error {
	log.Println("🔄 Setting up Event Bus...")

	// Create event bus configuration
	eventBusConfig := eventbus.EventBusConfig{
		Type:    eventbus.EventBusTypeNATS,
		NATSURL: c.Config.NATSURL, // Add this to your config
	}

	// Create event publisher
	publisher, err := eventbus.NewEventPublisher(eventBusConfig)
	if err != nil {
		return err
	}
	c.EventPublisher = publisher
	log.Println("✓ Event Publisher initialized")

	// Create event subscriber
	subscriber, err := eventbus.NewEventSubscriber(eventBusConfig)
	if err != nil {
		return err
	}
	c.EventSubscriber = subscriber
	log.Println("✓ Event Subscriber initialized")

	return nil
}

// SetupEventHandlers initializes all event handlers
func (c *Container) SetupEventHandlers() error {
	log.Println("🔄 Setting up Event Handlers...")

	// Create Auth Service V2 event handler (production)
	c.AuthServiceEventHandler = eventhandlers.NewAuthServiceEventHandler(
		c.UsersIdentityService,
		c.UserProfileRepository,
	)
	log.Println("✓ Auth Service V2 Event Handler created")

	// Old handler (for testing only)
	// c.UserEventHandler = eventhandlers.NewUserEventHandler(c.UsersCacheService)

	return nil
}

// SetupEventSubscriptions subscribes to all events
func (c *Container) SetupEventSubscriptions() error {
	log.Println("🔄 Setting up Event Subscriptions...")
	ctx := context.Background()

	// Subscribe to Auth Service V2 events (wildcard subscription)
	// ตาม INTEGRATION_GUIDE.md และ EVENT_SCHEMA_V2.md
	err := c.EventSubscriber.Subscribe(
		ctx,
		"user.events.*",                           // ← Wildcard pattern (user.events.created, user.events.updated, user.events.deleted)
		"social-backend-consumer",                 // ← Consumer name ตาม Auth Service recommendation
		c.AuthServiceEventHandler.HandleUserEvent, // ← Handler เดียว (แยก action ภายใน)
	)
	if err != nil {
		return err
	}
	log.Println("✓ Subscribed to: user.events.* (Auth Service V2)")

	log.Println("✅ All event subscriptions setup completed!")
	return nil
}

// Cleanup closes all event bus connections
func (c *Container) CleanupEventBus() error {
	log.Println("🛑 Cleaning up Event Bus...")

	if c.EventSubscriber != nil {
		if err := c.EventSubscriber.Close(); err != nil {
			log.Printf("⚠️  Error closing event subscriber: %v", err)
		}
	}

	if c.EventPublisher != nil {
		if err := c.EventPublisher.Close(); err != nil {
			log.Printf("⚠️  Error closing event publisher: %v", err)
		}
	}

	log.Println("✓ Event Bus cleaned up")
	return nil
}
