# 🚀 Event-Driven Advanced Topics

**เพิ่มเติมจากแผนหลัก 07_event_driven_migration_plan.md**

ตามคำแนะนำจากผู้เชี่ยวชาญ

---

## 📋 Table of Contents

1. [Event Ordering](#event-ordering)
2. [Event Replay](#event-replay)
3. [Schema Evolution](#schema-evolution)
4. [Monitoring & Observability](#monitoring--observability)
5. [Consumer Scaling](#consumer-scaling)
6. [Data Migration (Critical!)](#data-migration-critical)
7. [Structured Logging](#structured-logging)
8. [PoC Implementation Steps](#poc-implementation-steps)
9. [Bulk Migration Tool](#bulk-migration-tool)

---

## 1. Event Ordering

### ปัญหา: Event Order

**Scenario**:
```
Event 1: user.created (UserID: 123, Email: old@example.com)
Event 2: user.updated (UserID: 123, Email: new@example.com)

ถ้า Event 2 ถึงก่อน Event 1 → ข้อมูลผิด!
```

### Solution: Event Ordering Strategies

#### Strategy 1: NATS JetStream Ordering

**NATS รับประกัน ordering ภายใน stream เดียวกัน**

```go
// Publisher: Publish events sequentially
func (s *UserService) UpdateUserProfile(ctx context.Context, userID string, req UpdateRequest) error {
    // 1. Update in database
    user, err := s.repo.Update(ctx, userID, req)
    if err != nil {
        return err
    }

    // 2. Publish event (synchronous to preserve order)
    event := UserUpdatedEvent{
        BaseEvent: BaseEvent{
            EventID:      uuid.New().String(),
            EventType:    "user.updated",
            EventVersion: "v1",
            Timestamp:    time.Now(),
            SequenceNum:  s.getNextSequence(userID), // Important!
        },
        Data: UserUpdatedData{
            UserID: userID,
            Email:  user.Email,
        },
    }

    // Synchronous publish to preserve order
    err = s.eventPublisher.Publish(ctx, "AUTH_EVENTS.user.updated", event)
    if err != nil {
        log.Printf("Failed to publish event: %v", err)
        // Decision: Return error or just log?
    }

    return nil
}
```

**NATS Consumer Configuration**:
```go
// Ensure ordered delivery
_, err := js.Subscribe(
    "AUTH_EVENTS.user.*",
    func(msg *nats.Msg) {
        // Process message
    },
    nats.Durable("monolith-user-sync"),
    nats.ManualAck(),
    nats.AckExplicit(),
    nats.DeliverAll(),        // Deliver from beginning
    nats.AckWait(30*time.Second),
    nats.MaxAckPending(1),    // Process one message at a time (strict ordering)
)
```

**Trade-off**:
- ✅ Strict ordering guaranteed
- ❌ Slower throughput (sequential processing)

---

#### Strategy 2: Kafka Partition-based Ordering

**Kafka guarantees ordering per partition**

```go
// Kafka Publisher with key-based partitioning
func (p *KafkaPublisher) Publish(ctx context.Context, subject string, event Event) error {
    data, _ := json.Marshal(event)

    // Use UserID as partition key → same user always same partition
    err := p.writer.WriteMessages(ctx, kafka.Message{
        Topic: subject,
        Key:   []byte(event.GetUserID()), // Important! Same key = same partition
        Value: data,
    })

    return err
}
```

**Kafka Consumer Configuration**:
```go
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:  []string{"localhost:9092"},
    Topic:    "AUTH_EVENTS",
    GroupID:  "monolith-user-sync",
    MinBytes: 10e3,
    MaxBytes: 10e6,
    // Kafka automatically guarantees ordering per partition
})
```

---

#### Strategy 3: Idempotency + Sequence Numbers (Recommended!)

**Best Practice: Make consumers idempotent + use sequence numbers**

```go
// Add SequenceNum to events
type BaseEvent struct {
    EventID      string
    EventType    string
    EventVersion string
    Timestamp    time.Time
    SequenceNum  int64  // NEW: Sequence number per entity
}

// Consumer: Check sequence number before applying
func (h *UserEventHandler) HandleUserUpdated(ctx context.Context, msg *EventMessage) error {
    var event UserUpdatedEvent
    json.Unmarshal(msg.Data, &event)

    // 1. Get current sequence number from database
    currentSeq, err := h.repo.GetUserSequence(event.Data.UserID)
    if err != nil {
        return err
    }

    // 2. Check if event is newer
    if event.SequenceNum <= currentSeq {
        log.Printf("⚠️  Skipping old event (seq: %d, current: %d)", event.SequenceNum, currentSeq)
        return nil // Skip old event (idempotent)
    }

    // 3. Apply event
    err = h.usersCacheService.UpdateUser(ctx, event.Data)
    if err != nil {
        return err
    }

    // 4. Update sequence number
    err = h.repo.UpdateUserSequence(event.Data.UserID, event.SequenceNum)

    return err
}
```

**Database Schema**:
```sql
-- Add sequence tracking
ALTER TABLE users_cache ADD COLUMN sequence_num BIGINT DEFAULT 0;
CREATE INDEX idx_users_cache_sequence ON users_cache(id, sequence_num);
```

---

## 2. Event Replay

### Use Cases
1. **New Consumer** ต้องการข้อมูลย้อนหลัง
2. **Bug Fix** ต้อง replay events เพื่อแก้ข้อมูล
3. **Data Recovery** ต้องการ restore ข้อมูล

---

### NATS JetStream Event Replay

#### Method 1: DeliverAll (Replay from Beginning)

```go
// Create new consumer that replays all events
func ReplayAllEvents(js nats.JetStreamContext) error {
    // Subscribe with DeliverAll option
    _, err := js.Subscribe(
        "AUTH_EVENTS.user.*",
        func(msg *nats.Msg) {
            // Process message
            log.Printf("Replaying event: %s", msg.Subject)
            // ... handle event ...
            msg.Ack()
        },
        nats.Durable("replay-consumer-"+time.Now().Format("20060102150405")),
        nats.DeliverAll(), // Start from first message in stream
        nats.ManualAck(),
    )

    return err
}
```

---

#### Method 2: Replay from Specific Time

```go
func ReplayFromDate(js nats.JetStreamContext, startTime time.Time) error {
    _, err := js.Subscribe(
        "AUTH_EVENTS.user.*",
        func(msg *nats.Msg) {
            // Process message
        },
        nats.Durable("replay-time-"+startTime.Format("20060102")),
        nats.StartTime(startTime), // Replay from specific time
        nats.ManualAck(),
    )

    return err
}
```

---

#### Method 3: Replay Specific Sequence Range

```go
func ReplaySequenceRange(js nats.JetStreamContext, startSeq, endSeq uint64) error {
    // Fetch messages by sequence
    for seq := startSeq; seq <= endSeq; seq++ {
        msg, err := js.GetMsg("AUTH_EVENTS", seq)
        if err != nil {
            log.Printf("Failed to get message seq %d: %v", seq, err)
            continue
        }

        // Process message
        log.Printf("Replaying seq %d: %s", seq, string(msg.Data))

        // Handle event
        // ... your event handler logic ...
    }

    return nil
}
```

---

### Kafka Event Replay

```go
func ReplayKafkaFromBeginning(topic string) {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers:  []string{"localhost:9092"},
        Topic:    topic,
        GroupID:  "replay-consumer-" + uuid.New().String(), // New group = replay all
        MinBytes: 10e3,
        MaxBytes: 10e6,
    })

    // Seek to beginning
    reader.SetOffset(kafka.FirstOffset)

    for {
        msg, err := reader.ReadMessage(context.Background())
        if err != nil {
            break
        }

        // Process message
        log.Printf("Replaying: %s", string(msg.Value))
    }

    reader.Close()
}
```

---

### Replay Tool (Command Line)

```go
// cmd/replay/main.go
package main

import (
    "flag"
    "log"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    var (
        natsURL   = flag.String("nats", "nats://localhost:4222", "NATS URL")
        stream    = flag.String("stream", "AUTH_EVENTS", "Stream name")
        subject   = flag.String("subject", "AUTH_EVENTS.user.*", "Subject filter")
        startTime = flag.String("start-time", "", "Start time (RFC3339)")
        startSeq  = flag.Uint64("start-seq", 0, "Start sequence")
        endSeq    = flag.Uint64("end-seq", 0, "End sequence")
    )
    flag.Parse()

    // Connect to NATS
    nc, err := nats.Connect(*natsURL)
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    // Replay based on flags
    if *startSeq > 0 {
        log.Printf("Replaying from sequence %d to %d", *startSeq, *endSeq)
        replayBySequence(js, *stream, *startSeq, *endSeq)
    } else if *startTime != "" {
        t, _ := time.Parse(time.RFC3339, *startTime)
        log.Printf("Replaying from time %s", t)
        replayFromTime(js, *subject, t)
    } else {
        log.Println("Replaying all events")
        replayAll(js, *subject)
    }

    log.Println("✅ Replay completed")
}

func replayAll(js nats.JetStreamContext, subject string) {
    // Implementation...
}

func replayFromTime(js nats.JetStreamContext, subject string, startTime time.Time) {
    // Implementation...
}

func replayBySequence(js nats.JetStreamContext, stream string, start, end uint64) {
    // Implementation...
}
```

**Usage**:
```bash
# Replay all events
go run cmd/replay/main.go -nats nats://localhost:4222

# Replay from specific time
go run cmd/replay/main.go -start-time "2025-11-01T00:00:00Z"

# Replay specific sequence range
go run cmd/replay/main.go -start-seq 100 -end-seq 200
```

---

## 3. Schema Evolution

### Problem: Schema Changes Over Time

**Scenario**:
```
v1: user.created → { userId, email, username }
v2: user.created → { userId, email, username, phoneNumber } (NEW!)
```

**Challenge**: Old consumers ยังทำงานกับ v1 schema

---

### Solution: Backward Compatible Schema Changes

#### Rule 1: Always Add Fields (Never Remove)

```go
// ❌ BAD: Removed field (breaks old consumers)
type UserCreatedEventV2 struct {
    UserID   string
    Email    string
    // username removed! (breaking change!)
}

// ✅ GOOD: Added field with default value
type UserCreatedEventV2 struct {
    UserID      string  `json:"userId"`
    Email       string  `json:"email"`
    Username    string  `json:"username"`
    PhoneNumber *string `json:"phoneNumber,omitempty"` // NEW (optional)
}
```

---

#### Rule 2: Use Pointer for Optional Fields

```go
// Old event (v1)
type UserCreatedEventV1 struct {
    UserID   string
    Email    string
    Username string
}

// New event (v2) - backward compatible
type UserCreatedEventV2 struct {
    UserID      string
    Email       string
    Username    string
    PhoneNumber *string `json:"phoneNumber,omitempty"` // Pointer = optional
    Avatar      *string `json:"avatar,omitempty"`      // Pointer = optional
}
```

**Consumer handles both versions**:
```go
func HandleUserCreated(msg *EventMessage) error {
    var event UserCreatedEventV2 // Always use latest version
    json.Unmarshal(msg.Data, &event)

    // Check if field exists
    phoneNumber := ""
    if event.PhoneNumber != nil {
        phoneNumber = *event.PhoneNumber
    }

    // Save to database
    usersCacheService.CreateUser(event.UserID, event.Email, phoneNumber)
}
```

---

#### Rule 3: Version Dispatching

```go
func HandleUserCreated(msg *EventMessage) error {
    // Parse base event to check version
    var base BaseEvent
    json.Unmarshal(msg.Data, &base)

    // Dispatch based on version
    switch base.EventVersion {
    case "v1":
        return handleUserCreatedV1(msg)
    case "v2":
        return handleUserCreatedV2(msg)
    default:
        log.Printf("⚠️  Unknown event version: %s", base.EventVersion)
        return handleUserCreatedLatest(msg) // Default to latest
    }
}

func handleUserCreatedV1(msg *EventMessage) error {
    var event UserCreatedEventV1
    json.Unmarshal(msg.Data, &event)
    // ... handle v1 ...
}

func handleUserCreatedV2(msg *EventMessage) error {
    var event UserCreatedEventV2
    json.Unmarshal(msg.Data, &event)
    // ... handle v2 ...
}
```

---

#### Rule 4: Event Transformer Pattern

```go
// Transformer converts old events to new format
type EventTransformer interface {
    Transform(oldEvent interface{}) (newEvent interface{}, err error)
}

type UserCreatedV1ToV2Transformer struct{}

func (t *UserCreatedV1ToV2Transformer) Transform(oldEvent interface{}) (interface{}, error) {
    v1 := oldEvent.(UserCreatedEventV1)

    // Convert v1 to v2
    v2 := UserCreatedEventV2{
        UserID:      v1.UserID,
        Email:       v1.Email,
        Username:    v1.Username,
        PhoneNumber: nil, // Not available in v1
        Avatar:      nil,
    }

    return v2, nil
}

// Use transformer in consumer
func HandleUserCreated(msg *EventMessage) error {
    var base BaseEvent
    json.Unmarshal(msg.Data, &base)

    var event UserCreatedEventV2

    if base.EventVersion == "v1" {
        var v1 UserCreatedEventV1
        json.Unmarshal(msg.Data, &v1)

        // Transform v1 to v2
        transformer := UserCreatedV1ToV2Transformer{}
        v2, _ := transformer.Transform(v1)
        event = v2.(UserCreatedEventV2)
    } else {
        json.Unmarshal(msg.Data, &event)
    }

    // Now handle v2 uniformly
    return processUserCreatedV2(event)
}
```

---

## 4. Monitoring & Observability

### NATS JetStream Monitoring

#### Enable NATS Monitoring

```yaml
# docker-compose.yml
services:
  nats:
    image: nats:latest
    ports:
      - "4222:4222"  # Client connections
      - "8222:8222"  # HTTP monitoring (important!)
    command: ["-js", "-m", "8222"]
```

**Access Monitoring**:
```
http://localhost:8222

Endpoints:
- /varz      → Server info
- /connz     → Connection stats
- /routez    → Route info
- /subsz     → Subscription stats
- /jsz       → JetStream stats (important!)
```

---

#### Prometheus Metrics

```go
// infrastructure/eventbus/nats/nats_metrics.go
package nats

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    eventsPublishedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_published_total",
            Help: "Total number of events published",
        },
        []string{"event_type", "status"}, // labels
    )

    eventsPublishedDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "events_published_duration_seconds",
            Help:    "Duration of event publishing",
            Buckets: prometheus.DefBuckets,
        },
        []string{"event_type"},
    )

    eventsConsumedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_consumed_total",
            Help: "Total number of events consumed",
        },
        []string{"event_type", "status"},
    )

    eventsConsumedDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "events_consumed_duration_seconds",
            Help:    "Duration of event consumption",
            Buckets: prometheus.DefBuckets,
        },
        []string{"event_type"},
    )

    eventsConsumedErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_consumed_errors_total",
            Help: "Total number of event consumption errors",
        },
        []string{"event_type", "error_type"},
    )
)

// Instrument publisher
func (p *NATSPublisher) Publish(ctx context.Context, subject string, event Event) error {
    start := time.Now()

    err := p.publishInternal(ctx, subject, event)

    // Record metrics
    duration := time.Since(start).Seconds()
    status := "success"
    if err != nil {
        status = "error"
    }

    eventsPublishedTotal.WithLabelValues(event.GetEventType(), status).Inc()
    eventsPublishedDuration.WithLabelValues(event.GetEventType()).Observe(duration)

    return err
}

// Instrument consumer
func (h *UserEventHandler) HandleUserCreated(ctx context.Context, msg *EventMessage) error {
    start := time.Now()

    err := h.handleInternal(ctx, msg)

    duration := time.Since(start).Seconds()
    status := "success"
    if err != nil {
        status = "error"
        eventsConsumedErrors.WithLabelValues("user.created", err.Error()).Inc()
    }

    eventsConsumedTotal.WithLabelValues("user.created", status).Inc()
    eventsConsumedDuration.WithLabelValues("user.created").Observe(duration)

    return err
}
```

---

#### Grafana Dashboard

**Key Metrics to Monitor**:

```promql
# Event throughput (events/sec)
rate(events_published_total[5m])

# Event consumption lag
nats_stream_messages - nats_consumer_delivered

# Error rate
rate(events_consumed_errors_total[5m]) / rate(events_consumed_total[5m])

# P95 latency
histogram_quantile(0.95, rate(events_consumed_duration_seconds_bucket[5m]))

# Consumer pending messages
nats_consumer_pending{stream="AUTH_EVENTS", consumer="monolith-user-sync"}
```

---

### Alerting Rules

```yaml
# alerts.yml
groups:
  - name: event_bus
    rules:
      - alert: HighEventConsumptionErrors
        expr: rate(events_consumed_errors_total[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High event consumption error rate"
          description: "Error rate is {{ $value }} errors/sec"

      - alert: EventConsumerLag
        expr: nats_consumer_pending > 1000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Event consumer lagging behind"
          description: "Consumer has {{ $value }} pending messages"

      - alert: EventPublishingFailed
        expr: rate(events_published_total{status="error"}[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Event publishing failures"
```

---

## 5. Consumer Scaling

### NATS Queue Groups (Load Balancing)

**Multiple instances of consumer share the workload**

```go
// Instance 1, 2, 3 all subscribe with same queue group
func SetupConsumer(instanceID int) {
    subscriber.Subscribe(
        ctx,
        "AUTH_EVENTS.user.*",
        "monolith-user-sync", // Same queue group = load balancing
        eventHandler.HandleUserCreated,
    )

    log.Printf("Instance %d: Subscribed to AUTH_EVENTS", instanceID)
}
```

**NATS automatically distributes messages across instances**:
```
Event 1 → Instance 1
Event 2 → Instance 2
Event 3 → Instance 3
Event 4 → Instance 1 (round-robin)
```

---

### Kafka Consumer Groups (Partitioned Load Balancing)

```go
// Multiple consumers in same group
reader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:  []string{"localhost:9092"},
    Topic:    "AUTH_EVENTS",
    GroupID:  "monolith-user-sync", // Same group = share partitions
})
```

**Kafka distributes partitions across consumers**:
```
Topic: AUTH_EVENTS (3 partitions)

Consumer 1: Partition 0
Consumer 2: Partition 1
Consumer 3: Partition 2
```

---

### Horizontal Scaling Strategy

#### Step 1: Measure Current Load

```bash
# Check consumer pending messages
nats stream info AUTH_EVENTS

# Check processing rate
# If pending messages growing → need more consumers
```

---

#### Step 2: Add More Consumer Instances

```yaml
# docker-compose.yml
services:
  monolith-consumer-1:
    image: your-monolith:latest
    environment:
      - EVENT_CONSUMER_ENABLED=true
      - INSTANCE_ID=1

  monolith-consumer-2:
    image: your-monolith:latest
    environment:
      - EVENT_CONSUMER_ENABLED=true
      - INSTANCE_ID=2

  monolith-consumer-3:
    image: your-monolith:latest
    environment:
      - EVENT_CONSUMER_ENABLED=true
      - INSTANCE_ID=3
```

---

#### Step 3: Monitor Distribution

```promql
# Events per consumer instance
rate(events_consumed_total[5m]) by (instance_id)

# Should be roughly equal across instances
```

---

## 6. Data Migration (Critical!)

### Problem: Existing Users Not in Event Stream

**Scenario**:
```
Existing users in Auth Service: 10,000 users
Users in Monolith users_cache: 0 users (new setup)

Need to sync 10,000 users!
```

---

### Solution 1: Bulk Migration Script

```go
// scripts/migrate_existing_users.go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    _ "github.com/lib/pq"
)

func main() {
    // 1. Connect to Auth Service database
    authDB, err := sql.Open("postgres", "postgres://auth-db-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer authDB.Close()

    // 2. Connect to NATS
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    // 3. Query all users from Auth Service
    rows, err := authDB.Query(`
        SELECT id, email, username, display_name, avatar, role, is_active, created_at
        FROM users
        WHERE is_active = true
        ORDER BY created_at ASC
    `)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    publishedCount := 0
    batchSize := 100
    batch := []nats.PubAck{}

    // 4. Publish user.created events for all existing users
    for rows.Next() {
        var user User
        err := rows.Scan(
            &user.ID,
            &user.Email,
            &user.Username,
            &user.DisplayName,
            &user.Avatar,
            &user.Role,
            &user.IsActive,
            &user.CreatedAt,
        )
        if err != nil {
            log.Printf("Failed to scan user: %v", err)
            continue
        }

        // Create user.created event
        event := UserCreatedEvent{
            BaseEvent: BaseEvent{
                EventID:         uuid.New().String(),
                EventType:       "user.created",
                EventVersion:    "v1",
                Timestamp:       user.CreatedAt, // Use original created_at
                ProducerService: "migration-script",
                Metadata: map[string]interface{}{
                    "migration": true,
                    "batch":     true,
                },
            },
            Data: UserCreatedData{
                UserID:      user.ID,
                Email:       user.Email,
                Username:    user.Username,
                DisplayName: user.DisplayName,
                Avatar:      user.Avatar,
                Role:        user.Role,
                IsActive:    user.IsActive,
            },
        }

        // Publish event
        eventData, _ := json.Marshal(event)
        _, err = js.Publish("AUTH_EVENTS.user.created", eventData)
        if err != nil {
            log.Printf("Failed to publish event for user %s: %v", user.ID, err)
            continue
        }

        publishedCount++

        if publishedCount%batchSize == 0 {
            log.Printf("✅ Published %d users", publishedCount)
        }

        // Rate limiting (don't overwhelm NATS)
        time.Sleep(10 * time.Millisecond)
    }

    log.Printf("🎉 Migration completed! Published %d user.created events", publishedCount)
}

type User struct {
    ID          string
    Email       string
    Username    string
    DisplayName string
    Avatar      string
    Role        string
    IsActive    bool
    CreatedAt   time.Time
}
```

**Run migration**:
```bash
go run scripts/migrate_existing_users.go
```

---

### Solution 2: Direct Database Sync (Faster)

```go
// scripts/direct_sync_users.go
package main

import (
    "database/sql"
    "log"
)

func main() {
    // Connect to both databases
    authDB, _ := sql.Open("postgres", "auth-db-connection")
    monolithDB, _ := sql.Open("postgres", "monolith-db-connection")

    // Copy users directly (bypass events)
    _, err := monolithDB.Exec(`
        INSERT INTO users_cache (id, email, username, display_name, avatar, role, is_active, synced_at, created_at, updated_at)
        SELECT id, email, username, display_name, avatar, role, is_active, NOW(), created_at, updated_at
        FROM dblink('auth-db-connection', '
            SELECT id, email, username, display_name, avatar, role, is_active, created_at, updated_at
            FROM users
            WHERE is_active = true
        ') AS t(id uuid, email varchar, username varchar, display_name varchar, avatar varchar, role varchar, is_active boolean, created_at timestamp, updated_at timestamp)
        ON CONFLICT (id) DO UPDATE SET
            email = EXCLUDED.email,
            username = EXCLUDED.username,
            display_name = EXCLUDED.display_name,
            avatar = EXCLUDED.avatar,
            role = EXCLUDED.role,
            is_active = EXCLUDED.is_active,
            synced_at = NOW()
    `)

    if err != nil {
        log.Fatal(err)
    }

    log.Println("✅ Direct sync completed")
}
```

---

### Solution 3: Hybrid Approach (Recommended!)

**Best Practice: Direct sync first, then events**

```bash
# Step 1: Bulk sync existing users (fast)
go run scripts/direct_sync_users.go

# Step 2: Enable event-driven sync for new changes
# Auth Service starts publishing events
# Monolith subscribes to events

# Step 3: Verify sync
go run scripts/verify_sync.go
```

---

### Verification Script

```go
// scripts/verify_sync.go
package main

import (
    "database/sql"
    "log"
)

func main() {
    authDB, _ := sql.Open("postgres", "auth-db-connection")
    monolithDB, _ := sql.Open("postgres", "monolith-db-connection")

    // Count users in both databases
    var authCount, monolithCount int
    authDB.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&authCount)
    monolithDB.QueryRow("SELECT COUNT(*) FROM users_cache").Scan(&monolithCount)

    log.Printf("Auth DB users: %d", authCount)
    log.Printf("Monolith users_cache: %d", monolithCount)

    if authCount == monolithCount {
        log.Println("✅ Sync verified!")
    } else {
        log.Printf("⚠️  Mismatch! Difference: %d", authCount-monolithCount)

        // Find missing users
        rows, _ := authDB.Query(`
            SELECT id, email FROM users
            WHERE is_active = true
            AND id NOT IN (SELECT id FROM dblink('monolith-db-connection', 'SELECT id FROM users_cache'))
        `)

        for rows.Next() {
            var id, email string
            rows.Scan(&id, &email)
            log.Printf("Missing user: %s (%s)", id, email)
        }
    }
}
```

---

## 7. Structured Logging

### Logging Best Practices

```go
// pkg/logger/event_logger.go
package logger

import (
    "context"
    "encoding/json"
    "log"
    "time"
)

type EventLog struct {
    Timestamp     time.Time              `json:"timestamp"`
    Level         string                 `json:"level"`
    Service       string                 `json:"service"`
    Message       string                 `json:"message"`
    EventID       string                 `json:"eventId,omitempty"`
    EventType     string                 `json:"eventType,omitempty"`
    UserID        string                 `json:"userId,omitempty"`
    CorrelationID string                 `json:"correlationId,omitempty"`
    Duration      int64                  `json:"durationMs,omitempty"`
    Error         string                 `json:"error,omitempty"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

func LogEventPublished(eventID, eventType, userID string, duration time.Duration) {
    logEntry := EventLog{
        Timestamp: time.Now(),
        Level:     "info",
        Service:   "auth-service",
        Message:   "Event published",
        EventID:   eventID,
        EventType: eventType,
        UserID:    userID,
        Duration:  duration.Milliseconds(),
    }

    jsonLog, _ := json.Marshal(logEntry)
    log.Println(string(jsonLog))
}

func LogEventConsumed(eventID, eventType, userID string, duration time.Duration) {
    logEntry := EventLog{
        Timestamp: time.Now(),
        Level:     "info",
        Service:   "monolith",
        Message:   "Event consumed successfully",
        EventID:   eventID,
        EventType: eventType,
        UserID:    userID,
        Duration:  duration.Milliseconds(),
    }

    jsonLog, _ := json.Marshal(logEntry)
    log.Println(string(jsonLog))
}

func LogEventError(eventID, eventType, userID string, err error, retryCount int) {
    logEntry := EventLog{
        Timestamp: time.Now(),
        Level:     "error",
        Service:   "monolith",
        Message:   "Failed to process event",
        EventID:   eventID,
        EventType: eventType,
        UserID:    userID,
        Error:     err.Error(),
        Metadata: map[string]interface{}{
            "retryCount": retryCount,
        },
    }

    jsonLog, _ := json.Marshal(logEntry)
    log.Println(string(jsonLog))
}
```

---

### Use in Event Handlers

```go
func (h *UserEventHandler) HandleUserCreated(ctx context.Context, msg *EventMessage) error {
    start := time.Now()

    var event UserCreatedEvent
    json.Unmarshal(msg.Data, &event)

    // Log event received
    logger.LogEventReceived(event.EventID, event.EventType, event.Data.UserID)

    // Process event
    err := h.usersCacheService.CreateOrUpdateUser(ctx, event.Data)

    duration := time.Since(start)

    if err != nil {
        // Log error
        logger.LogEventError(
            event.EventID,
            event.EventType,
            event.Data.UserID,
            err,
            msg.Metadata.DeliveryCount,
        )
        return err
    }

    // Log success
    logger.LogEventConsumed(event.EventID, event.EventType, event.Data.UserID, duration)

    return nil
}
```

---

### Log Output Example

```json
{
  "timestamp": "2025-11-24T10:00:00Z",
  "level": "info",
  "service": "monolith",
  "message": "Event consumed successfully",
  "eventId": "a1b2c3d4-1234-5678-90ab-cdef12345678",
  "eventType": "user.created",
  "userId": "user-123",
  "durationMs": 50,
  "metadata": {
    "handler": "UserEventHandler",
    "consumerGroup": "monolith-user-sync"
  }
}
```

---

## 8. PoC Implementation Steps

### Phase 0: Preparation (1 day)

```bash
# 1. Setup local NATS
docker-compose up -d nats

# 2. Create streams
go run scripts/setup_nats_streams.go

# 3. Verify NATS is running
curl http://localhost:8222/varz
```

---

### Phase 1: PoC - Single Event (2-3 days)

**Goal**: Prove that user.created event works end-to-end

#### Step 1.1: Implement Event Publishing in Auth Service

```go
// auth-service: Publish user.created event
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }

    // Publish event
    event := UserCreatedEvent{...}
    go s.eventPublisher.Publish(ctx, "AUTH_EVENTS.user.created", event)

    return nil
}
```

---

#### Step 1.2: Implement Event Subscription in Monolith

```go
// monolith: Subscribe to user.created
subscriber.Subscribe(
    ctx,
    "AUTH_EVENTS.user.created",
    "monolith-user-sync",
    userEventHandler.HandleUserCreated,
)
```

---

#### Step 1.3: Test Manually

```bash
# 1. Create a test user in Auth Service
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"testuser","password":"password123"}'

# 2. Check NATS stream
nats stream info AUTH_EVENTS

# 3. Check Monolith database
psql -d monolith -c "SELECT * FROM users_cache WHERE email='test@example.com';"

# Expected: User should exist in users_cache
```

---

#### Step 1.4: Verify Metrics

```bash
# Check Prometheus metrics
curl http://localhost:3000/metrics | grep events_published_total
curl http://localhost:3000/metrics | grep events_consumed_total
```

---

### Phase 2: Dual-Mode Operation (1 week)

**Goal**: Run webhook + events in parallel, validate consistency

#### Step 2.1: Enable Dual Publishing in Auth Service

```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }

    // 1. Send webhook (existing)
    go s.webhookClient.SendUserCreated(user)

    // 2. Publish event (new)
    go s.eventPublisher.Publish(ctx, "AUTH_EVENTS.user.created", event)

    return nil
}
```

---

#### Step 2.2: Monolith Handles Both

```go
// Webhook handler (existing)
func (h *InternalHandler) HandleUserSync(c *fiber.Ctx) error {
    var payload WebhookPayload
    c.BodyParser(&payload)

    log.Println("📨 Webhook received")
    h.usersCacheService.CreateOrUpdateUser(ctx, payload.Data)

    return c.SendStatus(200)
}

// Event handler (new)
func (h *UserEventHandler) HandleUserCreated(ctx context.Context, msg *EventMessage) error {
    var event UserCreatedEvent
    json.Unmarshal(msg.Data, &event)

    log.Println("📥 Event received")
    h.usersCacheService.CreateOrUpdateUser(ctx, event.Data)

    return nil
}
```

---

#### Step 2.3: Validation Script

```go
// scripts/validate_dual_mode.go
func main() {
    // Compare webhook vs event delivery
    webhookCount := getWebhookReceivedCount()
    eventCount := getEventConsumedCount()

    log.Printf("Webhook received: %d", webhookCount)
    log.Printf("Events consumed: %d", eventCount)

    if webhookCount == eventCount {
        log.Println("✅ Dual mode working correctly")
    } else {
        log.Println("⚠️  Mismatch detected!")
    }
}
```

---

### Phase 3: Events Only (1 week monitoring)

**Goal**: Remove webhook, rely on events only

#### Step 3.1: Disable Webhook in Auth Service

```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) error {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return err
    }

    // Only publish event (webhook removed)
    go s.eventPublisher.Publish(ctx, "AUTH_EVENTS.user.created", event)

    return nil
}
```

---

#### Step 3.2: Remove Webhook Handler in Monolith

```go
// Comment out or remove webhook route
// app.Post("/internal/webhooks/user-sync", internalHandler.HandleUserSync)
```

---

#### Step 3.3: Monitor for 1 Week

```bash
# Check for errors
kubectl logs -f monolith-pod | grep ERROR

# Check consumer lag
nats consumer info AUTH_EVENTS monolith-user-sync

# Check user sync is still working
go run scripts/verify_sync.go
```

---

## 9. Bulk Migration Tool

### Complete Migration Tool

```go
// cmd/migrate/main.go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "flag"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/nats-io/nats.go"
    _ "github.com/lib/pq"
)

type MigrationConfig struct {
    AuthDBURL     string
    MonolithDBURL string
    NATSURL       string
    BatchSize     int
    RateLimit     int // events per second
    DryRun        bool
}

func main() {
    config := MigrationConfig{}

    flag.StringVar(&config.AuthDBURL, "auth-db", "", "Auth Service DB connection string")
    flag.StringVar(&config.MonolithDBURL, "monolith-db", "", "Monolith DB connection string")
    flag.StringVar(&config.NATSURL, "nats", "nats://localhost:4222", "NATS URL")
    flag.IntVar(&config.BatchSize, "batch-size", 100, "Batch size")
    flag.IntVar(&config.RateLimit, "rate-limit", 100, "Max events per second")
    flag.BoolVar(&config.DryRun, "dry-run", false, "Dry run (don't publish events)")
    flag.Parse()

    if config.AuthDBURL == "" || config.MonolithDBURL == "" {
        log.Fatal("Missing required flags: -auth-db, -monolith-db")
    }

    migrator := NewMigrator(config)
    migrator.Run()
}

type Migrator struct {
    config      MigrationConfig
    authDB      *sql.DB
    monolithDB  *sql.DB
    js          nats.JetStreamContext
    nc          *nats.Conn
    publishedCount int
    errorCount     int
}

func NewMigrator(config MigrationConfig) *Migrator {
    // Connect to Auth DB
    authDB, err := sql.Open("postgres", config.AuthDBURL)
    if err != nil {
        log.Fatal(err)
    }

    // Connect to Monolith DB
    monolithDB, err := sql.Open("postgres", config.MonolithDBURL)
    if err != nil {
        log.Fatal(err)
    }

    // Connect to NATS
    nc, err := nats.Connect(config.NATSURL)
    if err != nil {
        log.Fatal(err)
    }

    js, err := nc.JetStream()
    if err != nil {
        log.Fatal(err)
    }

    return &Migrator{
        config:     config,
        authDB:     authDB,
        monolithDB: monolithDB,
        js:         js,
        nc:         nc,
    }
}

func (m *Migrator) Run() {
    log.Println("🚀 Starting user migration...")

    // Step 1: Get total count
    var totalUsers int
    m.authDB.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&totalUsers)
    log.Printf("Total users to migrate: %d", totalUsers)

    if m.config.DryRun {
        log.Println("⚠️  DRY RUN MODE - No events will be published")
    }

    // Step 2: Migrate in batches
    offset := 0
    for {
        users := m.fetchUserBatch(offset, m.config.BatchSize)
        if len(users) == 0 {
            break
        }

        for _, user := range users {
            if err := m.migrateUser(user); err != nil {
                log.Printf("❌ Failed to migrate user %s: %v", user.ID, err)
                m.errorCount++
            } else {
                m.publishedCount++
            }

            // Rate limiting
            time.Sleep(time.Second / time.Duration(m.config.RateLimit))
        }

        offset += m.config.BatchSize
        progress := float64(offset) / float64(totalUsers) * 100
        log.Printf("Progress: %.2f%% (%d/%d)", progress, offset, totalUsers)
    }

    log.Println("✅ Migration completed!")
    log.Printf("Published: %d events", m.publishedCount)
    log.Printf("Errors: %d", m.errorCount)
}

func (m *Migrator) fetchUserBatch(offset, limit int) []User {
    rows, err := m.authDB.Query(`
        SELECT id, email, username, display_name, avatar, role, is_active, created_at
        FROM users
        WHERE is_active = true
        ORDER BY created_at ASC
        LIMIT $1 OFFSET $2
    `, limit, offset)
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()

    users := []User{}
    for rows.Next() {
        var user User
        rows.Scan(
            &user.ID,
            &user.Email,
            &user.Username,
            &user.DisplayName,
            &user.Avatar,
            &user.Role,
            &user.IsActive,
            &user.CreatedAt,
        )
        users = append(users, user)
    }

    return users
}

func (m *Migrator) migrateUser(user User) error {
    // Create event
    event := UserCreatedEvent{
        BaseEvent: BaseEvent{
            EventID:         uuid.New().String(),
            EventType:       "user.created",
            EventVersion:    "v1",
            Timestamp:       user.CreatedAt,
            ProducerService: "migration-tool",
            Metadata: map[string]interface{}{
                "migration": true,
            },
        },
        Data: UserCreatedData{
            UserID:      user.ID,
            Email:       user.Email,
            Username:    user.Username,
            DisplayName: user.DisplayName,
            Avatar:      user.Avatar,
            Role:        user.Role,
            IsActive:    user.IsActive,
        },
    }

    if m.config.DryRun {
        log.Printf("DRY RUN: Would publish event for user %s", user.ID)
        return nil
    }

    // Publish event
    eventData, _ := json.Marshal(event)
    _, err := m.js.Publish("AUTH_EVENTS.user.created", eventData)

    return err
}

type User struct {
    ID          string
    Email       string
    Username    string
    DisplayName string
    Avatar      string
    Role        string
    IsActive    bool
    CreatedAt   time.Time
}
```

---

### Usage

```bash
# Dry run (test without publishing)
go run cmd/migrate/main.go \
  -auth-db "postgres://user:pass@localhost:5432/auth_db" \
  -monolith-db "postgres://user:pass@localhost:5432/monolith_db" \
  -nats "nats://localhost:4222" \
  -dry-run

# Real migration (slow rate for safety)
go run cmd/migrate/main.go \
  -auth-db "postgres://user:pass@localhost:5432/auth_db" \
  -monolith-db "postgres://user:pass@localhost:5432/monolith_db" \
  -nats "nats://localhost:4222" \
  -rate-limit 50 \
  -batch-size 100

# Fast migration (production)
go run cmd/migrate/main.go \
  -auth-db "postgres://user:pass@localhost:5432/auth_db" \
  -monolith-db "postgres://user:pass@localhost:5432/monolith_db" \
  -nats "nats://localhost:4222" \
  -rate-limit 500 \
  -batch-size 500
```

---

## 🎯 Summary Checklist

### Before Going to Production

- [ ] **Event Ordering**: Implemented sequence numbers or partition keys
- [ ] **Event Replay**: Tested replay from specific time/sequence
- [ ] **Schema Evolution**: All events have version field
- [ ] **Monitoring**: NATS monitoring enabled (port 8222)
- [ ] **Prometheus Metrics**: Event publish/consume metrics instrumented
- [ ] **Grafana Dashboard**: Created dashboard for event bus metrics
- [ ] **Alerting**: Setup alerts for high error rate, consumer lag
- [ ] **Consumer Scaling**: Tested with multiple consumer instances
- [ ] **Data Migration**: Bulk migration script tested and executed
- [ ] **Structured Logging**: All events logged with EventID
- [ ] **PoC Completed**: Single event type working end-to-end
- [ ] **Dual Mode Tested**: Webhook + Events running in parallel
- [ ] **Verification Script**: Automated sync verification working
- [ ] **Load Testing**: Tested with production-level traffic
- [ ] **Rollback Plan**: Documented how to rollback to webhook

---

**Next**: ใช้เอกสารนี้ร่วมกับ `07_event_driven_migration_plan.md` เพื่อ implementation ที่สมบูรณ์ครับ! 🚀

ข้อแนะนำเพิ่มเติมสำหรับ PoC / rollout

Sequence number ทุก event สำคัญ

ใช้ป้องกัน race condition และให้แน่ใจว่า consumer process events ตามลำดับ

เริ่มด้วย PoC event “user.created” ก่อน

ทดสอบ flow, logging, และ metrics

ขยายไปยัง event อื่นหลังจากมั่นใจ

ตรวจสอบ metrics & logs ทุกขั้นตอน

ยืนยันว่า publisher → NATS → consumer → DB ทำงานถูกต้อง

ลดความเสี่ยงก่อน full rollout

Bulk migration

ควรทำ dry-run ก่อน

ตรวจสอบจำนวน records และ consistency ก่อน migrate จริง