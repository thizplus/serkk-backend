# 🎯 Event-Driven Migration Plan

**From Webhook-based Sync → Event-Driven Architecture (NATS → Kafka)**

---

## 📊 สถานการณ์ปัจจุบัน (Current State)

### Architecture ปัจจุบัน

```
┌─────────────────────┐         ┌──────────────────────┐
│   Auth Service      │         │   Main Monolith      │
│   (External)        │         │   (Our Backend)      │
│                     │         │                      │
│  - Authentication   │         │  - Social Features   │
│  - User CRUD        │         │  - Chat              │
│  - JWT Token        │         │  - Notifications     │
│                     │         │  - Media             │
│                     │         │  - Auto-Post         │
│                     │         │                      │
│                     │  HTTP   │                      │
│                     │ Webhook │  ┌────────────────┐  │
│                     │────────►│  │ InternalHandler│  │
│                     │  POST   │  │ HandleUserSync │  │
│                     │         │  └────────────────┘  │
│                     │         │         │            │
│                     │         │         ▼            │
│                     │         │  ┌────────────────┐  │
│                     │         │  │UsersCacheService│ │
│                     │         │  └────────────────┘  │
│                     │         │         │            │
│                     │         │         ▼            │
│                     │         │  ┌────────────────┐  │
│                     │         │  │  users_cache   │  │
│                     │         │  │  (PostgreSQL)  │  │
│                     │         │  └────────────────┘  │
└─────────────────────┘         └──────────────────────┘
```

### ปัญหาของ Webhook-based Approach

| Problem | Impact |
|---------|--------|
| **Synchronous** | Auth Service ต้องรอจน Monolith ตอบกลับ |
| **Tight Coupling** | Auth Service ต้องรู้ Monolith endpoint |
| **No Retry** | ถ้า Monolith down → event หายไป |
| **Single Consumer** | ถ้ามี service อื่นต้องการข้อมูล → ต้อง add webhook ใหม่ |
| **No Event History** | ไม่สามารถ replay events ได้ |

---

## 🎯 เป้าหมาย (Target State)

### Architecture ใหม่ (Event-Driven)

```
┌─────────────────────┐                            ┌──────────────────────┐
│   Auth Service      │                            │   Main Monolith      │
│   (External)        │                            │   (Our Backend)      │
│                     │                            │                      │
│  - Authentication   │                            │  - Social Features   │
│  - User CRUD        │                            │  - Chat              │
│  - JWT Token        │                            │  - Notifications     │
│                     │                            │  - Media             │
│                     │                            │  - Auto-Post         │
│                     │                            │                      │
│  ┌───────────────┐  │                            │  ┌────────────────┐  │
│  │Event Publisher│  │                            │  │Event Subscriber│  │
│  │ (Adapter)     │  │                            │  │  (Consumer)    │  │
│  └───────────────┘  │                            │  └────────────────┘  │
│         │           │                            │         ▲            │
└─────────┼───────────┘                            └─────────┼────────────┘
          │                                                  │
          │ Publish Event                                   │ Subscribe
          │                                                  │
          ▼                                                  │
┌──────────────────────────────────────────────────────────┐│
│                   Event Bus (NATS JetStream)              ││
│                                                            ││
│  Streams:                                                  ││
│  ├─ AUTH_EVENTS                                           ││
│  │  ├─ user.created                                       ││
│  │  ├─ user.updated                                       ││
│  │  ├─ user.deleted                                       ││
│  │  └─ user.password_changed                              ││
│  │                                                         ││
│  ├─ SOCIAL_EVENTS (future)                                ││
│  │  ├─ post.created                                       ││
│  │  └─ comment.created                                    ││
│  │                                                         ││
│  └─ NOTIFICATION_EVENTS (future)                          ││
│     └─ notification.sent                                  ││
│                                                            ││
└────────────────────────────────────────────────────────────┘
```

### ข้อดีของ Event-Driven

| Benefit | Description |
|---------|-------------|
| **Async** | Auth Service ไม่ต้องรอ Monolith ตอบกลับ |
| **Loose Coupling** | Auth Service ไม่ต้องรู้ว่าใครเป็น consumer |
| **Automatic Retry** | NATS JetStream retry เมื่อ consumer fail |
| **Multiple Consumers** | หลาย service subscribe event เดียวกันได้ |
| **Event History** | เก็บ event history, replay ได้ |
| **Scalability** | Consumer scale แยกกันได้ |

---

## 🏗️ Architecture Design

### 1. Event Bus Abstraction Layer

**หลักการสำคัญ**: **ไม่ผูกติดกับ NATS** → ใช้ Interface/Port Pattern (Clean Architecture)

```
domain/ports/
├── event_publisher.go      → Interface สำหรับ publish events
└── event_subscriber.go     → Interface สำหรับ subscribe events

infrastructure/eventbus/
├── nats/
│   ├── nats_publisher.go   → NATS implementation
│   └── nats_subscriber.go
├── kafka/
│   ├── kafka_publisher.go  → Kafka implementation (future)
│   └── kafka_subscriber.go
└── factory.go              → Factory pattern เลือก implementation
```

**Pattern นี้ทำให้**:
- ✅ เปลี่ยนจาก NATS → Kafka ได้โดยไม่แก้ business logic
- ✅ Test ง่าย (mock EventPublisher/Subscriber)
- ✅ Clean Architecture compliant

---

### 2. Event Schema Design

#### Base Event Interface

```go
// domain/events/base_event.go
package events

import "time"

type BaseEvent struct {
    EventID       string                 `json:"eventId"`       // UUID
    EventType     string                 `json:"eventType"`     // e.g., "user.created"
    EventVersion  string                 `json:"eventVersion"`  // e.g., "v1"
    Timestamp     time.Time              `json:"timestamp"`
    ProducerService string               `json:"producerService"` // e.g., "auth-service"
    Metadata      map[string]interface{} `json:"metadata"`
}

type Event interface {
    GetEventID() string
    GetEventType() string
    GetEventVersion() string
    GetTimestamp() time.Time
}
```

#### User Events

```go
// domain/events/user_events.go
package events

type UserCreatedEvent struct {
    BaseEvent
    Data UserCreatedData `json:"data"`
}

type UserCreatedData struct {
    UserID      string `json:"userId"`
    Email       string `json:"email"`
    Username    string `json:"username"`
    DisplayName string `json:"displayName,omitempty"`
    Avatar      string `json:"avatar,omitempty"`
    Role        string `json:"role"`
    IsActive    bool   `json:"isActive"`
}

type UserUpdatedEvent struct {
    BaseEvent
    Data UserUpdatedData `json:"data"`
}

type UserUpdatedData struct {
    UserID      string  `json:"userId"`
    Email       *string `json:"email,omitempty"`       // nil = not changed
    Username    *string `json:"username,omitempty"`
    DisplayName *string `json:"displayName,omitempty"`
    Avatar      *string `json:"avatar,omitempty"`
    Role        *string `json:"role,omitempty"`
    IsActive    *bool   `json:"isActive,omitempty"`
}

type UserDeletedEvent struct {
    BaseEvent
    Data UserDeletedData `json:"data"`
}

type UserDeletedData struct {
    UserID string `json:"userId"`
}
```

---

### 3. Event Publisher Interface (Port)

```go
// domain/ports/event_publisher.go
package ports

import (
    "context"
    "your-project/domain/events"
)

// EventPublisher is the interface for publishing events
// This allows us to switch from NATS to Kafka without changing business logic
type EventPublisher interface {
    // Publish publishes an event to the event bus
    // subject: topic/stream name (e.g., "AUTH_EVENTS.user.created")
    // event: the event data
    Publish(ctx context.Context, subject string, event events.Event) error

    // PublishBatch publishes multiple events atomically (if supported)
    PublishBatch(ctx context.Context, events []PublishRequest) error

    // Close closes the publisher connection
    Close() error
}

type PublishRequest struct {
    Subject string
    Event   events.Event
}
```

---

### 4. Event Subscriber Interface (Port)

```go
// domain/ports/event_subscriber.go
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
    Subject   string            // The subject/topic this message was received on
    Data      []byte            // Raw event data (JSON)
    Headers   map[string]string // Message headers
    Metadata  MessageMetadata   // Message metadata

    // Ack acknowledges the message (processed successfully)
    Ack func() error

    // Nack negatively acknowledges the message (failed, will retry)
    Nack func() error

    // Term terminates the message (failed permanently, move to DLQ)
    Term func() error
}

type MessageMetadata struct {
    MessageID     string
    Timestamp     int64
    DeliveryCount int  // How many times this message has been delivered
    StreamSeq     uint64
}
```

---

### 5. NATS JetStream Implementation

#### NATS Publisher

```go
// infrastructure/eventbus/nats/nats_publisher.go
package nats

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    "your-project/domain/events"
    "your-project/domain/ports"
)

type NATSPublisher struct {
    js nats.JetStreamContext
    nc *nats.Conn
}

func NewNATSPublisher(natsURL string) (*NATSPublisher, error) {
    // Connect to NATS
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    // Create JetStream context
    js, err := nc.JetStream()
    if err != nil {
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    return &NATSPublisher{
        js: js,
        nc: nc,
    }, nil
}

func (p *NATSPublisher) Publish(ctx context.Context, subject string, event events.Event) error {
    // Ensure event has metadata
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
    _, err = p.js.Publish(subject, data)
    if err != nil {
        return fmt.Errorf("failed to publish event to NATS: %w", err)
    }

    return nil
}

func (p *NATSPublisher) PublishBatch(ctx context.Context, requests []ports.PublishRequest) error {
    // NATS doesn't have native batch publish, so we publish one by one
    // TODO: Could use async publish for better performance
    for _, req := range requests {
        if err := p.Publish(ctx, req.Subject, req.Event); err != nil {
            return err
        }
    }
    return nil
}

func (p *NATSPublisher) Close() error {
    if p.nc != nil {
        p.nc.Close()
    }
    return nil
}
```

#### NATS Subscriber

```go
// infrastructure/eventbus/nats/nats_subscriber.go
package nats

import (
    "context"
    "fmt"

    "github.com/nats-io/nats.go"
    "your-project/domain/ports"
)

type NATSSubscriber struct {
    js            nats.JetStreamContext
    nc            *nats.Conn
    subscriptions []*nats.Subscription
}

func NewNATSSubscriber(natsURL string) (*NATSSubscriber, error) {
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    js, err := nc.JetStream()
    if err != nil {
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    return &NATSSubscriber{
        js:            js,
        nc:            nc,
        subscriptions: []*nats.Subscription{},
    }, nil
}

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
                Headers: convertHeaders(msg.Header),
                Metadata: ports.MessageMetadata{
                    MessageID: msg.Header.Get("Nats-Msg-Id"),
                },
                Ack:  func() error { return msg.Ack() },
                Nack: func() error { return msg.Nak() },
                Term: func() error { return msg.Term() },
            }

            // Call the handler
            if err := handler(ctx, eventMsg); err != nil {
                // Handler failed, negative acknowledge (will retry)
                _ = msg.Nak()
            } else {
                // Handler succeeded, acknowledge
                _ = msg.Ack()
            }
        },
        nats.Durable(consumerGroup),
        nats.ManualAck(),
        nats.AckExplicit(),
        nats.MaxDeliver(3), // Retry max 3 times
    )

    if err != nil {
        return fmt.Errorf("failed to subscribe: %w", err)
    }

    s.subscriptions = append(s.subscriptions, sub)
    return nil
}

func (s *NATSSubscriber) Unsubscribe(subject string) error {
    // Find and unsubscribe from matching subscriptions
    for i, sub := range s.subscriptions {
        if sub.Subject == subject {
            if err := sub.Unsubscribe(); err != nil {
                return err
            }
            // Remove from slice
            s.subscriptions = append(s.subscriptions[:i], s.subscriptions[i+1:]...)
        }
    }
    return nil
}

func (s *NATSSubscriber) Close() error {
    for _, sub := range s.subscriptions {
        _ = sub.Unsubscribe()
    }
    if s.nc != nil {
        s.nc.Close()
    }
    return nil
}

func convertHeaders(natsHeaders nats.Header) map[string]string {
    headers := make(map[string]string)
    for key := range natsHeaders {
        headers[key] = natsHeaders.Get(key)
    }
    return headers
}
```

---

### 6. Kafka Implementation (Future-Ready)

```go
// infrastructure/eventbus/kafka/kafka_publisher.go
package kafka

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/segmentio/kafka-go"
    "your-project/domain/events"
    "your-project/domain/ports"
)

type KafkaPublisher struct {
    writer *kafka.Writer
}

func NewKafkaPublisher(brokers []string) (*KafkaPublisher, error) {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(brokers...),
        Balancer: &kafka.LeastBytes{},
    }

    return &KafkaPublisher{
        writer: writer,
    }, nil
}

func (p *KafkaPublisher) Publish(ctx context.Context, subject string, event events.Event) error {
    // Serialize event to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    // Publish to Kafka
    err = p.writer.WriteMessages(ctx, kafka.Message{
        Topic: subject, // Kafka uses "topic" instead of "subject"
        Value: data,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(event.GetEventType())},
            {Key: "event-version", Value: []byte(event.GetEventVersion())},
        },
    })

    if err != nil {
        return fmt.Errorf("failed to publish event to Kafka: %w", err)
    }

    return nil
}

func (p *KafkaPublisher) PublishBatch(ctx context.Context, requests []ports.PublishRequest) error {
    messages := make([]kafka.Message, len(requests))
    for i, req := range requests {
        data, _ := json.Marshal(req.Event)
        messages[i] = kafka.Message{
            Topic: req.Subject,
            Value: data,
        }
    }
    return p.writer.WriteMessages(ctx, messages...)
}

func (p *KafkaPublisher) Close() error {
    return p.writer.Close()
}
```

```go
// infrastructure/eventbus/kafka/kafka_subscriber.go
package kafka

import (
    "context"
    "fmt"

    "github.com/segmentio/kafka-go"
    "your-project/domain/ports"
)

type KafkaSubscriber struct {
    readers []*kafka.Reader
}

func NewKafkaSubscriber(brokers []string) (*KafkaSubscriber, error) {
    return &KafkaSubscriber{
        readers: []*kafka.Reader{},
    }, nil
}

func (s *KafkaSubscriber) Subscribe(
    ctx context.Context,
    subject string,
    consumerGroup string,
    handler ports.EventHandler,
) error {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  []string{"localhost:9092"},
        Topic:    subject,
        GroupID:  consumerGroup,
        MinBytes: 10e3, // 10KB
        MaxBytes: 10e6, // 10MB
    })

    s.readers = append(s.readers, reader)

    // Start consuming in background goroutine
    go func() {
        for {
            msg, err := reader.ReadMessage(ctx)
            if err != nil {
                break
            }

            eventMsg := &ports.EventMessage{
                Subject: msg.Topic,
                Data:    msg.Value,
                Headers: convertKafkaHeaders(msg.Headers),
                Metadata: ports.MessageMetadata{
                    MessageID: fmt.Sprintf("%d-%d", msg.Partition, msg.Offset),
                    Timestamp: msg.Time.Unix(),
                },
                Ack:  func() error { return reader.CommitMessages(ctx, msg) },
                Nack: func() error { return nil }, // Kafka doesn't have Nack, just don't commit
                Term: func() error { return reader.CommitMessages(ctx, msg) },
            }

            if err := handler(ctx, eventMsg); err != nil {
                // Don't commit (will be retried)
                continue
            }
            // Auto-commit handled by reader config
        }
    }()

    return nil
}

func (s *KafkaSubscriber) Unsubscribe(subject string) error {
    // Implementation...
    return nil
}

func (s *KafkaSubscriber) Close() error {
    for _, reader := range s.readers {
        _ = reader.Close()
    }
    return nil
}

func convertKafkaHeaders(headers []kafka.Header) map[string]string {
    result := make(map[string]string)
    for _, h := range headers {
        result[h.Key] = string(h.Value)
    }
    return result
}
```

---

### 7. Factory Pattern (Switch Between NATS and Kafka)

```go
// infrastructure/eventbus/factory.go
package eventbus

import (
    "fmt"

    "your-project/domain/ports"
    "your-project/infrastructure/eventbus/kafka"
    "your-project/infrastructure/eventbus/nats"
)

type EventBusType string

const (
    EventBusTypeNATS  EventBusType = "nats"
    EventBusTypeKafka EventBusType = "kafka"
)

type EventBusConfig struct {
    Type    EventBusType
    NATSURL string   // For NATS
    Brokers []string // For Kafka
}

func NewEventPublisher(config EventBusConfig) (ports.EventPublisher, error) {
    switch config.Type {
    case EventBusTypeNATS:
        return nats.NewNATSPublisher(config.NATSURL)
    case EventBusTypeKafka:
        return kafka.NewKafkaPublisher(config.Brokers)
    default:
        return nil, fmt.Errorf("unsupported event bus type: %s", config.Type)
    }
}

func NewEventSubscriber(config EventBusConfig) (ports.EventSubscriber, error) {
    switch config.Type {
    case EventBusTypeNATS:
        return nats.NewNATSSubscriber(config.NATSURL)
    case EventBusTypeKafka:
        return kafka.NewKafkaSubscriber(config.Brokers)
    default:
        return nil, fmt.Errorf("unsupported event bus type: %s", config.Type)
    }
}
```

---

### 8. Configuration (.env)

```env
# Event Bus Configuration
EVENT_BUS_TYPE=nats              # or "kafka"

# NATS Configuration (used when EVENT_BUS_TYPE=nats)
NATS_URL=nats://localhost:4222

# Kafka Configuration (used when EVENT_BUS_TYPE=kafka)
KAFKA_BROKERS=localhost:9092,localhost:9093
```

---

## 📝 Implementation Steps

### Step 1: Setup NATS JetStream (1-2 days)

#### Install NATS Server

```bash
# Using Docker Compose
# docker-compose.yml
version: '3.8'
services:
  nats:
    image: nats:latest
    ports:
      - "4222:4222"  # Client connections
      - "8222:8222"  # HTTP monitoring
    command: ["-js", "-m", "8222"]  # Enable JetStream and monitoring
    volumes:
      - nats-data:/data
    environment:
      - NATS_JETSTREAM_MAX_MEMORY=1G
      - NATS_JETSTREAM_MAX_STORAGE=10G

volumes:
  nats-data:
```

```bash
docker-compose up -d nats
```

#### Create NATS Streams

```go
// scripts/setup_nats_streams.go
package main

import (
    "fmt"
    "log"

    "github.com/nats-io/nats.go"
)

func main() {
    // Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    // Create AUTH_EVENTS stream
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "AUTH_EVENTS",
        Subjects: []string{"AUTH_EVENTS.*"}, // AUTH_EVENTS.user.created, etc.
        Storage:  nats.FileStorage,
        MaxAge:   30 * 24 * 60 * 60 * 1e9, // 30 days retention
        MaxBytes: 10 * 1024 * 1024 * 1024, // 10 GB max storage
    })
    if err != nil {
        log.Printf("Stream AUTH_EVENTS already exists or error: %v", err)
    } else {
        fmt.Println("✅ Created stream: AUTH_EVENTS")
    }

    // Create SOCIAL_EVENTS stream (for future)
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "SOCIAL_EVENTS",
        Subjects: []string{"SOCIAL_EVENTS.*"},
        Storage:  nats.FileStorage,
        MaxAge:   30 * 24 * 60 * 60 * 1e9,
        MaxBytes: 10 * 1024 * 1024 * 1024,
    })
    if err != nil {
        log.Printf("Stream SOCIAL_EVENTS already exists or error: %v", err)
    } else {
        fmt.Println("✅ Created stream: SOCIAL_EVENTS")
    }

    fmt.Println("✅ NATS streams setup completed")
}
```

---

### Step 2: Implement Event Bus Abstraction (2-3 days)

#### Create the interfaces and implementations following the structure above:

```
1. Create domain/ports/event_publisher.go
2. Create domain/ports/event_subscriber.go
3. Create domain/events/base_event.go
4. Create domain/events/user_events.go
5. Create infrastructure/eventbus/nats/nats_publisher.go
6. Create infrastructure/eventbus/nats/nats_subscriber.go
7. Create infrastructure/eventbus/factory.go
```

---

### Step 3: Update Auth Service to Publish Events (3-4 days)

**Auth Service Side** (if you have access to modify it):

```go
// auth-service/application/serviceimpl/user_service_impl.go
package serviceimpl

import (
    "context"
    "your-auth-project/domain/events"
    "your-auth-project/domain/ports"
)

type UserServiceImpl struct {
    userRepo       ports.UserRepository
    eventPublisher ports.EventPublisher // Add this
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req CreateUserRequest) error {
    // 1. Create user in database
    user, err := s.userRepo.Create(ctx, req)
    if err != nil {
        return err
    }

    // 2. Publish event (FIRE AND FORGET - don't wait!)
    event := events.UserCreatedEvent{
        BaseEvent: events.BaseEvent{
            EventID:         uuid.New().String(),
            EventType:       "user.created",
            EventVersion:    "v1",
            Timestamp:       time.Now(),
            ProducerService: "auth-service",
        },
        Data: events.UserCreatedData{
            UserID:      user.ID,
            Email:       user.Email,
            Username:    user.Username,
            DisplayName: user.DisplayName,
            Avatar:      user.Avatar,
            Role:        user.Role,
            IsActive:    user.IsActive,
        },
    }

    // Publish in background (don't block the API response)
    go func() {
        if err := s.eventPublisher.Publish(context.Background(), "AUTH_EVENTS.user.created", event); err != nil {
            // Log error but don't fail the user creation
            log.Printf("Failed to publish user.created event: %v", err)
        }
    }()

    return nil
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) error {
    // 1. Update user in database
    user, err := s.userRepo.Update(ctx, userID, req)
    if err != nil {
        return err
    }

    // 2. Publish event
    event := events.UserUpdatedEvent{
        BaseEvent: events.BaseEvent{
            EventID:         uuid.New().String(),
            EventType:       "user.updated",
            EventVersion:    "v1",
            Timestamp:       time.Now(),
            ProducerService: "auth-service",
        },
        Data: events.UserUpdatedData{
            UserID:      user.ID,
            Email:       &user.Email,
            Username:    &user.Username,
            DisplayName: &user.DisplayName,
            Avatar:      &user.Avatar,
        },
    }

    go func() {
        _ = s.eventPublisher.Publish(context.Background(), "AUTH_EVENTS.user.updated", event)
    }()

    return nil
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID string) error {
    // 1. Delete user
    err := s.userRepo.Delete(ctx, userID)
    if err != nil {
        return err
    }

    // 2. Publish event
    event := events.UserDeletedEvent{
        BaseEvent: events.BaseEvent{
            EventID:         uuid.New().String(),
            EventType:       "user.deleted",
            EventVersion:    "v1",
            Timestamp:       time.Now(),
            ProducerService: "auth-service",
        },
        Data: events.UserDeletedData{
            UserID: userID,
        },
    }

    go func() {
        _ = s.eventPublisher.Publish(context.Background(), "AUTH_EVENTS.user.deleted", event)
    }()

    return nil
}
```

---

### Step 4: Update Monolith to Subscribe to Events (3-4 days)

#### Create Event Handlers

```go
// application/eventhandlers/user_event_handler.go
package eventhandlers

import (
    "context"
    "encoding/json"
    "log"

    "your-project/domain/events"
    "your-project/domain/ports"
    "your-project/domain/services"
)

type UserEventHandler struct {
    usersCacheService services.UsersCacheService
}

func NewUserEventHandler(usersCacheService services.UsersCacheService) *UserEventHandler {
    return &UserEventHandler{
        usersCacheService: usersCacheService,
    }
}

// HandleUserCreated handles user.created event
func (h *UserEventHandler) HandleUserCreated(ctx context.Context, msg *ports.EventMessage) error {
    // Parse event
    var event events.UserCreatedEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        log.Printf("Failed to unmarshal user.created event: %v", err)
        return err
    }

    log.Printf("📥 Received event: user.created - UserID: %s, Email: %s", event.Data.UserID, event.Data.Email)

    // Create user in users_cache
    err := h.usersCacheService.CreateOrUpdateUser(ctx, services.UserCacheDTO{
        ID:          event.Data.UserID,
        Email:       event.Data.Email,
        Username:    event.Data.Username,
        DisplayName: event.Data.DisplayName,
        Avatar:      event.Data.Avatar,
        Role:        event.Data.Role,
        IsActive:    event.Data.IsActive,
    })

    if err != nil {
        log.Printf("❌ Failed to create user in cache: %v", err)
        return err // Will trigger retry
    }

    log.Printf("✅ Successfully synced user.created - UserID: %s", event.Data.UserID)
    return nil
}

// HandleUserUpdated handles user.updated event
func (h *UserEventHandler) HandleUserUpdated(ctx context.Context, msg *ports.EventMessage) error {
    var event events.UserUpdatedEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        return err
    }

    log.Printf("📥 Received event: user.updated - UserID: %s", event.Data.UserID)

    // Update user in users_cache
    err := h.usersCacheService.CreateOrUpdateUser(ctx, services.UserCacheDTO{
        ID:          event.Data.UserID,
        Email:       getStringValue(event.Data.Email),
        Username:    getStringValue(event.Data.Username),
        DisplayName: getStringValue(event.Data.DisplayName),
        Avatar:      getStringValue(event.Data.Avatar),
        Role:        getStringValue(event.Data.Role),
        IsActive:    getBoolValue(event.Data.IsActive),
    })

    if err != nil {
        log.Printf("❌ Failed to update user in cache: %v", err)
        return err
    }

    log.Printf("✅ Successfully synced user.updated - UserID: %s", event.Data.UserID)
    return nil
}

// HandleUserDeleted handles user.deleted event
func (h *UserEventHandler) HandleUserDeleted(ctx context.Context, msg *ports.EventMessage) error {
    var event events.UserDeletedEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        return err
    }

    log.Printf("📥 Received event: user.deleted - UserID: %s", event.Data.UserID)

    // Delete user from users_cache (soft delete or hard delete)
    err := h.usersCacheService.DeleteUser(ctx, event.Data.UserID)
    if err != nil {
        log.Printf("❌ Failed to delete user from cache: %v", err)
        return err
    }

    log.Printf("✅ Successfully synced user.deleted - UserID: %s", event.Data.UserID)
    return nil
}

func getStringValue(ptr *string) string {
    if ptr == nil {
        return ""
    }
    return *ptr
}

func getBoolValue(ptr *bool) bool {
    if ptr == nil {
        return false
    }
    return *ptr
}
```

---

#### Register Event Subscribers in DI Container

```go
// pkg/di/container.go
package di

import (
    "context"
    "log"

    "your-project/application/eventhandlers"
    "your-project/domain/ports"
    "your-project/infrastructure/eventbus"
)

func (c *Container) SetupEventSubscribers() error {
    ctx := context.Background()

    // Create event subscriber
    eventBusConfig := eventbus.EventBusConfig{
        Type:    eventbus.EventBusType(c.config.EventBusType), // "nats" or "kafka"
        NATSURL: c.config.NATSURL,
        Brokers: c.config.KafkaBrokers,
    }

    subscriber, err := eventbus.NewEventSubscriber(eventBusConfig)
    if err != nil {
        return err
    }

    c.eventSubscriber = subscriber

    // Create event handlers
    userEventHandler := eventhandlers.NewUserEventHandler(c.usersCacheService)

    // Subscribe to AUTH_EVENTS.user.created
    err = subscriber.Subscribe(
        ctx,
        "AUTH_EVENTS.user.created",
        "monolith-user-sync", // Consumer group name
        userEventHandler.HandleUserCreated,
    )
    if err != nil {
        return err
    }
    log.Println("✅ Subscribed to: AUTH_EVENTS.user.created")

    // Subscribe to AUTH_EVENTS.user.updated
    err = subscriber.Subscribe(
        ctx,
        "AUTH_EVENTS.user.updated",
        "monolith-user-sync",
        userEventHandler.HandleUserUpdated,
    )
    if err != nil {
        return err
    }
    log.Println("✅ Subscribed to: AUTH_EVENTS.user.updated")

    // Subscribe to AUTH_EVENTS.user.deleted
    err = subscriber.Subscribe(
        ctx,
        "AUTH_EVENTS.user.deleted",
        "monolith-user-sync",
        userEventHandler.HandleUserDeleted,
    )
    if err != nil {
        return err
    }
    log.Println("✅ Subscribed to: AUTH_EVENTS.user.deleted")

    log.Println("🎉 Event subscribers setup completed")
    return nil
}
```

---

#### Update main.go

```go
// main.go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "your-project/pkg/di"
)

func main() {
    // Initialize DI container
    container, err := di.NewContainer()
    if err != nil {
        log.Fatal(err)
    }

    // Setup event subscribers
    if err := container.SetupEventSubscribers(); err != nil {
        log.Fatalf("Failed to setup event subscribers: %v", err)
    }

    // Start HTTP server
    app := container.GetFiberApp()
    go func() {
        if err := app.Listen(":3000"); err != nil {
            log.Fatal(err)
        }
    }()

    log.Println("🚀 Server started on :3000")

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("🛑 Shutting down gracefully...")

    // Cleanup
    container.Cleanup()

    log.Println("👋 Server stopped")
}
```

---

### Step 5: Deprecate Webhook (Gradually)

#### Phase 1: Dual Mode (Both Webhook + Events)

```
Auth Service:
  ├─ Publish event (new)
  └─ Send webhook (old, for backward compatibility)

Monolith:
  ├─ Subscribe to events (new, primary)
  └─ Handle webhook (old, fallback)
```

**Monitor**: Check if events are working correctly

---

#### Phase 2: Events Only (Remove Webhook)

```
Auth Service:
  ├─ Publish event (only)
  └─ Remove webhook code

Monolith:
  ├─ Subscribe to events (only)
  └─ Remove webhook endpoint
```

---

### Step 6: Testing (2-3 days)

#### Unit Tests

```go
// application/eventhandlers/user_event_handler_test.go
package eventhandlers_test

import (
    "context"
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "your-project/application/eventhandlers"
    "your-project/domain/events"
    "your-project/domain/ports"
)

type MockUsersCacheService struct {
    mock.Mock
}

func (m *MockUsersCacheService) CreateOrUpdateUser(ctx context.Context, dto interface{}) error {
    args := m.Called(ctx, dto)
    return args.Error(0)
}

func TestHandleUserCreated(t *testing.T) {
    // Arrange
    mockService := new(MockUsersCacheService)
    handler := eventhandlers.NewUserEventHandler(mockService)

    event := events.UserCreatedEvent{
        Data: events.UserCreatedData{
            UserID:   "user-123",
            Email:    "test@example.com",
            Username: "testuser",
        },
    }
    eventData, _ := json.Marshal(event)

    msg := &ports.EventMessage{
        Data: eventData,
    }

    mockService.On("CreateOrUpdateUser", mock.Anything, mock.Anything).Return(nil)

    // Act
    err := handler.HandleUserCreated(context.Background(), msg)

    // Assert
    assert.NoError(t, err)
    mockService.AssertExpectations(t)
}
```

---

#### Integration Tests

```go
// tests/integration/event_integration_test.go
package integration_test

import (
    "context"
    "encoding/json"
    "testing"
    "time"

    "github.com/nats-io/nats.go"
    "github.com/stretchr/testify/assert"
    "your-project/domain/events"
)

func TestEventPublishAndSubscribe(t *testing.T) {
    // Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    assert.NoError(t, err)
    defer nc.Close()

    js, err := nc.JetStream()
    assert.NoError(t, err)

    // Publish event
    event := events.UserCreatedEvent{
        BaseEvent: events.BaseEvent{
            EventID:   "test-123",
            EventType: "user.created",
            Timestamp: time.Now(),
        },
        Data: events.UserCreatedData{
            UserID:   "user-123",
            Email:    "test@example.com",
            Username: "testuser",
        },
    }

    eventData, _ := json.Marshal(event)
    _, err = js.Publish("AUTH_EVENTS.user.created", eventData)
    assert.NoError(t, err)

    // Subscribe and verify
    received := make(chan bool)
    _, err = js.Subscribe("AUTH_EVENTS.user.created", func(msg *nats.Msg) {
        var receivedEvent events.UserCreatedEvent
        json.Unmarshal(msg.Data, &receivedEvent)

        assert.Equal(t, "user-123", receivedEvent.Data.UserID)
        received <- true
    })
    assert.NoError(t, err)

    // Wait for message
    select {
    case <-received:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("Timeout waiting for event")
    }
}
```

---

## 🔄 Migration Path: NATS → Kafka

### When to Migrate?

**Stay with NATS if**:
- ✅ Event throughput < 100,000 events/day
- ✅ Event retention < 30 days
- ✅ Simple use cases
- ✅ Team comfortable with NATS

**Migrate to Kafka if**:
- ⚠️ Event throughput > 1,000,000 events/day
- ⚠️ Need long-term event retention (> 30 days)
- ⚠️ Need complex event processing (Kafka Streams)
- ⚠️ Need multi-datacenter replication

---

### Migration Steps (When Ready)

#### Step 1: Deploy Kafka Cluster

```yaml
# docker-compose.yml (add Kafka)
version: '3.8'
services:
  kafka:
    image: confluentinc/cp-kafka:latest
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    ports:
      - "2181:2181"
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
```

---

#### Step 2: Update Configuration

```env
# .env (switch to Kafka)
EVENT_BUS_TYPE=kafka              # Changed from "nats"
KAFKA_BROKERS=localhost:9092
```

---

#### Step 3: Deploy and Test

```bash
# No code changes needed!
# Factory pattern automatically switches to Kafka implementation

# 1. Stop services
docker-compose down

# 2. Update .env
# EVENT_BUS_TYPE=kafka

# 3. Start services
docker-compose up -d

# 4. Verify
# Check logs: should see "Using Kafka event bus"
```

---

#### Step 4: Dual Mode (Optional Safety)

**Run both NATS and Kafka temporarily**:

```go
// Publish to both (for safety)
func (s *UserService) CreateUser(...) error {
    // ... create user ...

    // Publish to both NATS and Kafka
    go func() {
        _ = s.natsPublisher.Publish(ctx, subject, event)
        _ = s.kafkaPublisher.Publish(ctx, subject, event)
    }()
}
```

**Subscribe from both**:

```go
// Subscribe to both (one will be primary, one backup)
natsSubscriber.Subscribe(ctx, subject, group, handler)
kafkaSubscriber.Subscribe(ctx, subject, group, handler)
```

**After verifying Kafka works → remove NATS**

---

## 📊 Monitoring & Observability

### Metrics to Track

```prometheus
# Event publishing
events_published_total{service="auth-service", event_type="user.created", status="success"}
events_published_duration_seconds{service="auth-service", event_type="user.created"}

# Event consumption
events_consumed_total{service="monolith", event_type="user.created", status="success"}
events_consumed_duration_seconds{service="monolith", event_type="user.created"}
events_consumed_errors_total{service="monolith", event_type="user.created"}

# Queue metrics (NATS)
nats_stream_messages{stream="AUTH_EVENTS"}
nats_stream_bytes{stream="AUTH_EVENTS"}
nats_consumer_pending{stream="AUTH_EVENTS", consumer="monolith-user-sync"}
```

---

### Logging

```json
// Publisher log
{
  "timestamp": "2025-11-24T10:00:00Z",
  "level": "info",
  "service": "auth-service",
  "message": "Event published",
  "eventType": "user.created",
  "eventId": "uuid",
  "userId": "user-123",
  "subject": "AUTH_EVENTS.user.created"
}

// Consumer log
{
  "timestamp": "2025-11-24T10:00:01Z",
  "level": "info",
  "service": "monolith",
  "message": "Event consumed successfully",
  "eventType": "user.created",
  "eventId": "uuid",
  "userId": "user-123",
  "processingTimeMs": 50
}

// Error log
{
  "timestamp": "2025-11-24T10:00:02Z",
  "level": "error",
  "service": "monolith",
  "message": "Failed to process event",
  "eventType": "user.created",
  "eventId": "uuid",
  "error": "database connection timeout",
  "retryCount": 2
}
```

---

## 🎯 Summary & Timeline

### Implementation Timeline

| Phase | Duration | Tasks |
|-------|----------|-------|
| **Phase 1: Setup** | 1-2 days | Install NATS, create streams |
| **Phase 2: Abstraction** | 2-3 days | Implement Event Bus interfaces |
| **Phase 3: Auth Service** | 3-4 days | Add event publishing to Auth Service |
| **Phase 4: Monolith** | 3-4 days | Add event subscription to Monolith |
| **Phase 5: Testing** | 2-3 days | Unit + integration tests |
| **Phase 6: Deploy** | 1 day | Deploy to production (dual mode) |
| **Phase 7: Deprecate Webhook** | 1 week | Monitor, then remove webhook |
| **Total** | **2-3 weeks** | |

---

### Key Benefits

| Benefit | Description |
|---------|-------------|
| **Loose Coupling** | Services don't need to know about each other |
| **Async** | Non-blocking communication |
| **Scalability** | Easy to add new consumers |
| **Flexibility** | Switch NATS → Kafka without code changes |
| **Event History** | Can replay events |
| **Resilience** | Automatic retry on failure |

---

### Next Steps

1. ✅ **Review this plan** with your team
2. ✅ **Setup NATS JetStream** locally (docker-compose)
3. ✅ **Implement Event Bus abstraction** (interfaces + NATS impl)
4. ✅ **Update Auth Service** to publish events
5. ✅ **Update Monolith** to subscribe to events
6. ✅ **Test thoroughly** (unit + integration)
7. ✅ **Deploy** (dual mode: webhook + events)
8. ✅ **Monitor** for 1 week
9. ✅ **Remove webhook** if events working perfectly
10. ✅ **Celebrate** 🎉

---

**คำถามเพิ่มเติม?** พร้อมช่วยเหมือนเดิมครับ! 🚀
