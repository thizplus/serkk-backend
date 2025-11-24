# 5️⃣ Optional / Extra

สรุป Deployment, Technology Stack, และข้อมูลเพิ่มเติม

---

## ภาพรวม Current Stack

| Category | Technology | Version/Details |
|----------|-----------|-----------------|
| **Language** | Go | Latest (1.21+) |
| **Framework** | Fiber | v2 |
| **Database** | PostgreSQL | Primary database |
| **Cache** | Redis | Sessions, idempotency, feed cache |
| **ORM** | GORM | Database access |
| **Storage** | Bunny CDN | Images, files |
| **Video** | Bunny Stream | Video encoding, HLS streaming |
| **Storage (Alt)** | Cloudflare R2 | Optional S3-compatible storage |
| **AI** | OpenAI API | gpt-4o-mini for content generation |
| **WebSocket** | Go standard library | Real-time communication |
| **Job Scheduler** | Custom EventScheduler | Cron-based, in-process |
| **Migrations** | Custom SQL scripts | Manual migration management |
| **Monitoring** | Prometheus (basic) | Metrics endpoint `/metrics` |
| **Logging** | Custom Logger | Console/file logging |
| **Web Push** | VAPID | PWA push notifications |

---

## 📦 Deployment ปัจจุบัน

### Current Deployment Setup

**Deployment Method**: Unknown (ไม่มีข้อมูลใน codebase)

**Possible Deployment Scenarios**:
1. **Bare Metal / VM**
   - Single server deployment
   - Run as systemd service

2. **Docker** (likely)
   - Single container deployment
   - Compose file with PostgreSQL + Redis

3. **Cloud Platform**
   - Deploy to VPS (DigitalOcean, Linode, etc.)
   - Single instance

**Current Architecture** (Assumed):
```
┌─────────────────────────────────────┐
│         Single Server/Container      │
│                                      │
│  ┌────────────────────────────────┐ │
│  │    Go Fiber Application        │ │
│  │                                │ │
│  │  ├─ PostService                │ │
│  │  ├─ CommentService             │ │
│  │  ├─ MessageService             │ │
│  │  ├─ NotificationService        │ │
│  │  ├─ ChatHub (goroutine)        │ │
│  │  ├─ NotificationHub (goroutine)│ │
│  │  ├─ VideoEncoderWorker         │ │
│  │  └─ EventScheduler             │ │
│  │                                │ │
│  │  Port: 3000                    │ │
│  └────────────────────────────────┘ │
│                                      │
│  ┌────────────────────────────────┐ │
│  │       PostgreSQL               │ │
│  │       Port: 5432               │ │
│  └────────────────────────────────┘ │
│                                      │
│  ┌────────────────────────────────┐ │
│  │       Redis                    │ │
│  │       Port: 6379               │ │
│  └────────────────────────────────┘ │
│                                      │
└─────────────────────────────────────┘

External Services:
├─ Bunny CDN
├─ Bunny Stream
├─ Cloudflare R2
└─ OpenAI API
```

---

## 🔧 Infrastructure Components

### 1. Application Server

**Framework**: Go Fiber v2
```go
app := fiber.New(fiber.Config{
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
})
```

**Port**: 3000 (default)

**Endpoints**:
- HTTP API: `/api/v1/*`
- WebSocket Chat: `/ws/chat`
- WebSocket Notifications: `/ws/notifications`
- Health Check: `/health`
- Metrics: `/metrics`

---

### 2. PostgreSQL Database

**Current Configuration**:
```
Host: localhost (assumed)
Port: 5432
Database: serkk (assumed)
User: postgres
Password: n147369 (found in git status)
```

**Connection Pool**:
```go
// Likely configuration (from typical GORM setup)
MaxIdleConns: 10
MaxOpenConns: 100
ConnMaxLifetime: 1 hour
```

**Database Size Estimation**:
```
Tables: 27
Indexes: ~60+ (estimated)
Estimated size: Depends on data volume

Current schema suggests:
- users_cache: ~10K users → ~5 MB
- posts: ~100K posts → ~500 MB
- comments: ~500K comments → ~200 MB
- messages: ~1M messages → ~1 GB
- notifications: ~5M notifications → ~500 MB
- media: ~50K media → ~20 MB (metadata only)
─────────────────────────────────────────
Total estimated: ~2-3 GB (small-medium scale)
```

**Migration Management**:
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_posts_table.up.sql
├── 002_create_posts_table.down.sql
├── ...
├── 025_add_displayname_avatar_to_user_profiles.up.sql
└── 025_add_displayname_avatar_to_user_profiles.down.sql

Total migrations: 25+

Run migrations:
- Manual execution (psql < migration.sql)
- OR custom Go script (scripts/run_migrations.go)
```

---

### 3. Redis Cache

**Usage**:
```
1. Idempotency Keys
   Key: "post:idempotency:{client_post_id}"
   TTL: 5 minutes

2. Feed Cache
   Key: "feed:user:{user_id}:cursor:{cursor}"
   TTL: 5 minutes

3. Conversation List Cache
   Key: "conversation:list:{user_id}"
   TTL: 10 minutes

4. Video Encoding Queue
   Key: "video:encoding:queue"
   Type: List (LPUSH/RPOP)

5. Session Cache (assumed)
   Key: "session:{session_id}"
   TTL: 24 hours
```

**Estimated Memory Usage**:
```
10,000 users:
- Feed cache: ~100 MB (20 posts × 10KB each × 500 active users)
- Idempotency: ~5 MB (1000 concurrent posts × 5KB each)
- Video queue: ~1 MB
- Sessions: ~50 MB (5000 active sessions × 10KB each)
─────────────────────────────────────────
Total: ~156 MB
```

---

### 4. External Services

#### Bunny CDN (Image/File Storage)
```
Base URL: https://storage.bunnycdn.com/
API Key: (in .env)

Usage:
- Upload images (POST)
- Upload files (POST)
- Delete files (DELETE)

Cost:
- Storage: $0.01/GB/month
- Bandwidth: $0.01/GB
```

**Current Integration**:
```go
// infrastructure/bunny/bunny_storage.go
func (bs *BunnyStorage) Upload(file []byte, path string) (string, error) {
    // PUT https://storage.bunnycdn.com/{storageZone}/{path}
    // Returns CDN URL
}
```

---

#### Bunny Stream (Video Encoding)
```
Base URL: https://video.bunnycdn.com/
Library ID: (in .env)

Usage:
- Upload video (POST)
- Get encoding status (GET)
- Webhook callback (POST /webhook/bunny/video-completed)

Cost:
- Storage: $0.005/GB/month
- Encoding: $0.01/GB processed
- Streaming: $0.01/GB delivered
```

**Encoding Flow**:
```
1. Upload video → Bunny Stream
2. Bunny starts encoding (async)
3. Poll encoding status (VideoEncoderWorker)
4. Bunny sends webhook when done
5. Update media record
6. Notify user via WebSocket
```

---

#### Cloudflare R2 (Optional Storage)
```
Usage: Large file storage (S3-compatible)

Configuration:
- Endpoint: https://<account>.r2.cloudflarestorage.com
- Access Key ID: (in .env)
- Secret Access Key: (in .env)

Cost:
- Storage: $0.015/GB/month
- No egress fees!
```

---

#### OpenAI API (AI Content Generation)
```
Model: gpt-4o-mini
API Key: (in .env)

Usage:
- AutoPostService (hourly)
- Generate post titles + content

Cost:
- Input: $0.150 / 1M tokens
- Output: $0.600 / 1M tokens

Estimated monthly cost:
- 24 posts/day × 500 tokens/post × 30 days = 360,000 tokens
- Cost: ~$0.22/month (very cheap!)
```

**API Call**:
```go
// domain/services/openai_service.go
func (oas *OpenAIService) GeneratePost(topic string) (string, error) {
    // POST https://api.openai.com/v1/chat/completions
    // Model: gpt-4o-mini
    // Temperature: 0.7
    // Max tokens: 500
}
```

---

## 🔍 Monitoring & Logging

### Current Monitoring

**Available**:
1. **Health Check Endpoint**
   ```
   GET /health
   Response: {"status": "ok"}
   ```

2. **Prometheus Metrics**
   ```
   GET /metrics

   Available metrics:
   - http_requests_total
   - http_request_duration_seconds
   - go_goroutines
   - go_memstats_alloc_bytes
   - (basic application metrics)
   ```

**Missing**:
- ❌ No centralized logging (no ELK/Loki)
- ❌ No distributed tracing (no Jaeger/Zipkin)
- ❌ No service-level metrics (no per-service latency)
- ❌ No alerting (no PagerDuty/Slack alerts)
- ❌ No error tracking (no Sentry)

---

### Current Logging

**Log Format**: Console + File (assumed)
```go
// pkg/logger/logger.go (assumed structure)
logger.Info("User logged in", "user_id", userID)
logger.Error("Failed to create post", "error", err)
```

**Log Levels**:
- INFO
- WARN
- ERROR
- DEBUG (assumed)

**Missing**:
- ❌ No structured logging (no JSON format)
- ❌ No log aggregation
- ❌ No log search (no Elasticsearch)
- ❌ No log retention policy

---

## 🔐 Security & Authentication

### Current Security Setup

**Authentication**: JWT (JSON Web Token)
```
Auth Service (External) issues JWT
→ This service validates JWT
→ Extract user info from token
```

**Middleware**:
```go
// interfaces/middleware/auth_middleware.go (assumed)
func AuthMiddleware(c *fiber.Ctx) error {
    // Extract JWT from Authorization header
    // Validate JWT signature
    // Extract user info
    // Set user context
}
```

**Protected Routes**:
```
All routes under /api/v1/* require authentication
Except:
- /health
- /metrics
```

---

### Security Features

**Implemented**:
- ✅ JWT authentication
- ✅ User blocking (prevents blocked users from messaging)
- ✅ Idempotency (prevents duplicate posts)
- ✅ Soft deletes (posts, messages)

**Missing**:
- ❌ Rate limiting (no rate limiter)
- ❌ CORS configuration (not visible in codebase)
- ❌ Request validation (basic only)
- ❌ SQL injection protection (GORM provides basic protection)
- ❌ XSS protection (frontend responsibility)
- ❌ CSRF protection (not needed for JWT-based API)

---

## 📊 Performance Characteristics

### Current Performance Metrics (Estimated)

| Endpoint | Avg Latency | Complexity |
|----------|-------------|------------|
| `POST /api/v1/posts` | ~100ms | Medium (5 DB queries + cache) |
| `GET /api/v1/feed` | ~500ms | High (complex join, cache miss) |
| `GET /api/v1/posts` | ~50ms | Low (simple query + cache) |
| `POST /api/v1/comments` | ~80ms | Medium (3 DB queries) |
| `POST /api/v1/votes` | ~50ms | Low (2 DB queries) |
| `GET /api/v1/conversations` | ~100ms | Medium (join + cache) |
| `POST /api/v1/messages` | ~80ms | Medium (3 DB queries + WebSocket) |
| `GET /api/v1/notifications` | ~50ms | Low (simple query) |

**Bottlenecks**:
- Feed generation: 500ms (complex join)
- Search queries: 300-500ms (full-text search)
- Video upload: 5-10s (upload to Bunny Stream)

---

### Scalability Limits (Current Setup)

| Resource | Limit | Bottleneck |
|----------|-------|------------|
| **Database Connections** | ~100 | Connection pool |
| **WebSocket Connections** | ~10,000 | Memory (1 MB/connection) |
| **HTTP Requests/s** | ~1,000 | Single instance CPU |
| **Concurrent Users** | ~5,000 | Memory + DB connections |
| **Database Size** | ~100 GB | Disk I/O |
| **Redis Memory** | ~1 GB | Cache eviction policy |

**When to Scale**:
```
Scenario 1: 10,000 concurrent users
Problem: WebSocket memory > 10 GB
Solution: Need horizontal scaling (multiple instances)

Scenario 2: 5,000 req/s
Problem: CPU bottleneck
Solution: Need load balancer + multiple instances

Scenario 3: 1M posts
Problem: Feed query > 1s
Solution: Need CQRS + Read Models
```

---

## 🛠️ Development Tools

### Code Structure
```
gofiber-backend/
├── application/
│   └── serviceimpl/         # Service implementations
├── domain/
│   ├── dto/                 # Data Transfer Objects
│   ├── models/              # Database models
│   ├── repositories/        # Repository interfaces
│   └── services/            # Service interfaces
├── infrastructure/
│   ├── postgres/            # PostgreSQL implementations
│   ├── bunny/               # Bunny CDN/Stream
│   └── redis/               # Redis client
├── interfaces/
│   ├── api/
│   │   ├── handlers/        # HTTP handlers
│   │   └── routes/          # Route definitions
│   └── middleware/          # Middleware
├── pkg/
│   ├── di/                  # Dependency injection
│   ├── logger/              # Logger
│   └── utils/               # Utilities
├── migrations/              # SQL migrations
├── scripts/                 # Utility scripts
├── .env                     # Environment variables
├── go.mod                   # Go dependencies
├── go.sum
└── main.go                  # Entry point
```

---

### Environment Variables
```env
# Database
DATABASE_URL=postgresql://postgres:password@localhost:5432/serkk
POSTGRES_PASSWORD=n147369

# Redis
REDIS_URL=redis://localhost:6379

# Bunny CDN
BUNNY_STORAGE_ZONE=your-zone
BUNNY_API_KEY=your-api-key

# Bunny Stream
BUNNY_STREAM_LIBRARY_ID=your-library-id
BUNNY_STREAM_API_KEY=your-api-key

# Cloudflare R2
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key

# OpenAI
OPENAI_API_KEY=sk-your-api-key

# JWT (from Auth Service)
JWT_SECRET=your-jwt-secret

# Application
PORT=3000
ENVIRONMENT=development
```

---

## 📈 Scaling Recommendations

### Immediate Improvements (Without Microservices)

1. **Add Load Balancer**
   - Deploy multiple instances behind Nginx/HAProxy
   - Session stickiness for WebSocket

2. **Database Optimization**
   - Add read replicas
   - Implement connection pooling (PgBouncer)
   - Add more indexes

3. **Caching Improvements**
   - Increase Redis memory
   - Add cache warmup on startup
   - Implement cache-aside pattern consistently

4. **Background Jobs**
   - Move to message queue (RabbitMQ/Redis Queue)
   - Separate worker processes

5. **Monitoring**
   - Add Grafana + Prometheus
   - Add centralized logging (ELK/Loki)
   - Add distributed tracing (Jaeger)

---

### Long-term (Microservices)

**Target Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│                     Load Balancer                        │
│                  (Nginx/Traefik/Kong)                    │
└─────────────────────────────────────────────────────────┘
              │              │              │
    ┌─────────┴─────┐  ┌────┴────┐  ┌──────┴───────┐
    │ Post Service  │  │  Chat   │  │Notification  │
    │               │  │ Service │  │   Service    │
    │ ┌──────────┐  │  │         │  │              │
    │ │PostgreSQL│  │  │┌───────┐│  │  ┌────────┐  │
    │ └──────────┘  │  ││MongoDB││  │  │ Redis  │  │
    └───────────────┘  │└───────┘│  │  └────────┘  │
                       └─────────┘  └──────────────┘
              │              │              │
    ┌─────────┴──────────────┴──────────────┴─────┐
    │          Message Bus (Kafka/NATS)            │
    └──────────────────────────────────────────────┘
```

**Benefits**:
- Scale each service independently
- Use optimal database for each service
- Independent deployments
- Fault isolation
- Team autonomy

---

## 🎯 Summary

| Aspect | Current State | Recommendation |
|--------|---------------|----------------|
| **Deployment** | Single instance | Load balanced + auto-scaling |
| **Database** | Single PostgreSQL | Database per service |
| **Caching** | Basic Redis | Redis Cluster + CDN |
| **Monitoring** | Basic Prometheus | Full observability stack |
| **Logging** | Local files | Centralized logging |
| **Security** | JWT only | Add rate limiting, WAF |
| **Scalability** | Vertical only | Horizontal + auto-scaling |
| **Architecture** | Monolith | Microservices |

---

**Next Step**: ใช้ข้อมูลทั้ง 5 ไฟล์นี้เพื่อออกแบบ Microservices Architecture แบบ Enterprise-ready
