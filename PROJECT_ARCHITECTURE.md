# SUEKK Backend - Project Architecture

> สถาปัตยกรรมของ Social Media Platform Backend (Monolith → Microservices Ready)

---

## 🏗️ High-Level Architecture Overview

```mermaid
graph TB
    Client[Next.js Frontend<br/>Port 3000] --> API[Go Fiber Backend<br/>Port 8080]
    AuthClient[Auth Frontend] --> AuthService[Auth Service V2<br/>Port 8088]

    API --> Redis[Redis Cache<br/>Port 6379]
    API --> PostgresMain[(PostgreSQL<br/>gofiber_social)]
    API --> NATS[NATS Event Bus<br/>Port 4222]
    API --> R2[Cloudflare R2<br/>Media Storage]
    API --> Bunny[Bunny CDN<br/>Video Streaming]

    AuthService --> PostgresAuth[(PostgreSQL<br/>gofiber_auth)]
    AuthService --> NATS

    NATS --> API

    style API fill:#4A90E2
    style AuthService fill:#E27B4A
    style PostgresMain fill:#336791
    style PostgresAuth fill:#336791
```

---

## 📁 Project Structure (Clean Architecture)

```
gofiber-backend/
│
├─ cmd/                          # Application Entry Points
│  └─ api/
│     └─ main.go                 # Main application (Port 8080)
│
├─ domain/                       # 🎯 Business Logic Layer (Core Domain)
│  ├─ dto/                       # Data Transfer Objects
│  │  ├─ auth.go
│  │  ├─ user.go
│  │  ├─ post.go
│  │  └─ mappers.go              # Entity ↔ DTO conversions
│  │
│  ├─ events/                    # Domain Events (Event-Driven)
│  │  ├─ auth_events.go
│  │  └─ user_events.go
│  │
│  ├─ models/                    # Domain Entities (Database Models)
│  │  ├─ users_identity.go       # Auth identity (id, email, username)
│  │  ├─ user_profile.go         # Social profile (displayName, karma, bio)
│  │  ├─ post.go
│  │  ├─ comment.go
│  │  ├─ conversation.go
│  │  ├─ message.go
│  │  └─ ...
│  │
│  ├─ repositories/              # Repository Interfaces (Ports)
│  │  ├─ user_repository.go
│  │  ├─ post_repository.go
│  │  └─ ...
│  │
│  └─ services/                  # Service Interfaces (Ports)
│     ├─ user_profile_service.go
│     ├─ post_service.go
│     └─ ...
│
├─ application/                  # 🔧 Application Layer (Use Cases)
│  ├─ serviceimpl/               # Service Implementations
│  │  ├─ user_profile_service_impl.go
│  │  ├─ post_service_impl.go
│  │  ├─ conversation_service_impl.go
│  │  └─ ...
│  │
│  └─ eventhandlers/             # Event Handlers
│     └─ auth_service_v2_event_handler.go  # Listens to Auth Service events
│
├─ infrastructure/               # 🛠️ Infrastructure Layer (External Services)
│  ├─ postgres/                  # PostgreSQL Implementations
│  │  ├─ user_repository_impl.go
│  │  ├─ user_profile_repository_impl.go
│  │  ├─ post_repository_impl.go
│  │  ├─ conversation_repository_impl.go
│  │  └─ ...
│  │
│  ├─ redis/                     # Redis Cache
│  │  └─ redis_service.go
│  │
│  ├─ eventbus/                  # Event Bus (NATS)
│  │  └─ nats/
│  │     ├─ publisher.go
│  │     └─ subscriber.go
│  │
│  ├─ storage/                   # File Storage
│  │  ├─ r2_storage.go           # Cloudflare R2
│  │  ├─ bunny_storage.go        # Bunny CDN
│  │  └─ bunny_stream.go         # Video Streaming
│  │
│  ├─ websocket/                 # WebSocket Hubs
│  │  ├─ chat_hub.go
│  │  └─ notification_hub.go
│  │
│  └─ workers/                   # Background Workers
│     └─ video_encoder_worker.go
│
├─ interfaces/                   # 🌐 Presentation Layer (HTTP/WebSocket)
│  └─ api/
│     ├─ handlers/               # HTTP Handlers (Controllers)
│     │  ├─ user_profile_handler.go
│     │  ├─ post_handler.go
│     │  ├─ conversation_handler.go
│     │  ├─ internal_handler.go  # Internal APIs (webhooks)
│     │  └─ ...
│     │
│     ├─ routes/                 # Route Definitions
│     │  ├─ routes.go            # Main router
│     │  ├─ user_profile_routes.go
│     │  ├─ post_routes.go
│     │  ├─ internal_routes.go
│     │  └─ ...
│     │
│     └─ middleware/             # HTTP Middleware
│        ├─ auth.go              # JWT validation
│        ├─ cors.go
│        └─ ...
│
├─ pkg/                          # 📦 Shared Packages (Utilities)
│  ├─ config/                    # Configuration
│  ├─ database/                  # Database connection
│  ├─ di/                        # Dependency Injection
│  ├─ logger/                    # Logging
│  ├─ cache/                     # Cache utilities
│  ├─ auth_code_store/           # OAuth helpers
│  ├─ scheduler/                 # Cron jobs
│  ├─ ai/                        # AI integrations
│  └─ utils/                     # Common utilities
│
├─ migrations/                   # 🗄️ Database Migrations
│  ├─ 001_initial_schema.sql
│  ├─ 023_create_users_cache_table.sql
│  ├─ 024_drop_users_fk_constraints.sql
│  ├─ 025_add_displayname_avatar_to_user_profiles.sql
│  └─ ...
│
└─ scripts/                      # 🔧 Utility Scripts
   ├─ fix_duplicate_conversations.go
   ├─ sync_users_initial.go
   └─ ...
```

---

## 🎯 Architecture Layers (Clean Architecture)

```mermaid
graph LR
    A[Presentation Layer<br/>interfaces/api] --> B[Application Layer<br/>application/serviceimpl]
    B --> C[Domain Layer<br/>domain/models,services]
    B --> D[Infrastructure Layer<br/>infrastructure/postgres,redis]
    D --> C

    style A fill:#FF6B6B
    style B fill:#4ECDC4
    style C fill:#FFE66D
    style D fill:#95E1D3
```

### Layer Responsibilities:

1. **Presentation Layer** (`interfaces/`)
   - HTTP Handlers (Controllers)
   - Route definitions
   - Request/Response DTOs
   - Middleware (auth, CORS, etc.)

2. **Application Layer** (`application/`)
   - Service implementations (business logic)
   - Event handlers
   - Use case orchestration

3. **Domain Layer** (`domain/`)
   - Core business entities (models)
   - Business rules
   - Repository/Service interfaces (ports)

4. **Infrastructure Layer** (`infrastructure/`)
   - Database implementations
   - External service integrations
   - Event bus, cache, storage

---

## 🗄️ Database Architecture

### Current Setup: Dual Database (V2 Architecture)

```mermaid
graph TB
    subgraph "Auth Service Database (gofiber_auth)"
        U1[(users table<br/>Full user data)]
    end

    subgraph "Social Backend Database (gofiber_social)"
        UI[(users_identity<br/>id, email, username)]
        UP[(user_profiles<br/>displayName, avatar,<br/>bio, karma, stats)]
        P[(posts)]
        C[(comments)]
        CV[(conversations)]
        M[(messages)]
    end

    U1 -.NATS Events.-> UI
    UI --> UP
    P --> UI
    C --> UI
    CV --> UI
    M --> UI

    style U1 fill:#E27B4A
    style UI fill:#4A90E2
    style UP fill:#4A90E2
```

### Table Relationships:

**users_identity** (Identity data synced from Auth Service)
- ✅ Contains: `id`, `email`, `username`, `created_at`, `synced_at`
- ❌ Does NOT contain: `password`, `role`, `displayName`, `avatar`

**user_profiles** (Social-specific data)
- ✅ Contains: `displayName`, `avatar`, `bio`, `karma`, `followers_count`, `following_count`
- 🔗 Foreign Key: `user_id` → `users_identity.id`

**Other tables** reference `users_identity.id` for user relationships

---

## 🔄 Event-Driven Architecture (NATS)

```mermaid
sequenceDiagram
    participant Auth as Auth Service
    participant NATS as NATS Event Bus
    participant Backend as Social Backend
    participant DB as PostgreSQL (social)

    Auth->>NATS: Publish user.events.created
    NATS->>Backend: Subscribe (consumer group)
    Backend->>DB: INSERT users_identity
    Backend->>DB: INSERT user_profiles (default)
    Backend-->>NATS: Ack

    Note over Auth,DB: User data stays in sync via events
```

### Event Flow:

1. **Auth Service** publishes events:
   - `user.events.created` - New user registered
   - `user.events.updated` - User info updated
   - `user.events.deleted` - User deleted

2. **Social Backend** listens via NATS:
   - Consumer Group: `social-backend-consumer`
   - Subject: `user.events.*`
   - Handler: `auth_service_v2_event_handler.go`

3. **Data Sync**:
   - Creates/updates `users_identity` record
   - Creates default `user_profiles` if new user

---

## 🚀 Current Services & Features

### Core Features:
- ✅ **User Management** (via Auth Service V2 events)
- ✅ **User Profiles** (displayName, avatar, bio, karma)
- ✅ **Posts** (with media, tags, votes)
- ✅ **Comments** (nested, with votes)
- ✅ **Real-time Chat** (WebSocket, conversations, messages)
- ✅ **Notifications** (WebSocket, real-time)
- ✅ **Follow System** (followers, following)
- ✅ **Search** (posts, users, tags)
- ✅ **Media Upload** (Cloudflare R2, Bunny CDN)
- ✅ **Video Streaming** (Bunny Stream)
- ✅ **Auto-post System** (AI-generated posts, scheduled)

### External Integrations:
- **Redis** - Caching, feed cache
- **Cloudflare R2** - Image/file storage
- **Bunny CDN** - Video streaming
- **NATS** - Event bus (Auth Service ↔ Backend)
- **OpenAI** - AI post generation

---

## 🔮 Microservice Migration Plan

### Phase 1: Current State (Monolith with Separate Auth)
```mermaid
graph LR
    Frontend --> Backend[Social Backend Monolith<br/>Port 8080]
    Frontend --> Auth[Auth Service<br/>Port 8088]
    Backend --> DB1[(gofiber_social)]
    Auth --> DB2[(gofiber_auth)]
    Backend <-.NATS.-> Auth
```

### Phase 2: Extract Chat Service (Next Step)
```mermaid
graph LR
    Frontend --> API[API Gateway]
    API --> Social[Social Service<br/>posts, comments, profiles]
    API --> Chat[Chat Service<br/>conversations, messages]
    API --> Auth[Auth Service]

    Social --> DB1[(social_db)]
    Chat --> DB2[(chat_db)]
    Auth --> DB3[(auth_db)]

    Social <-.NATS.-> Chat
    Social <-.NATS.-> Auth
    Chat <-.NATS.-> Auth
```

### Phase 3: Full Microservices
```mermaid
graph TB
    Frontend --> Gateway[API Gateway / BFF]

    Gateway --> Auth[Auth Service]
    Gateway --> Social[Social Service<br/>posts, comments, votes]
    Gateway --> Profile[Profile Service<br/>user profiles, karma]
    Gateway --> Chat[Chat Service<br/>conversations, messages]
    Gateway --> Notification[Notification Service<br/>real-time notifications]
    Gateway --> Media[Media Service<br/>upload, processing]
    Gateway --> Search[Search Service<br/>posts, users, tags]

    subgraph "Event Bus (NATS)"
        NATS[NATS Messaging]
    end

    Auth <--> NATS
    Social <--> NATS
    Profile <--> NATS
    Chat <--> NATS
    Notification <--> NATS
    Media <--> NATS
    Search <--> NATS
```

---

## 📊 Service Boundaries (Suggested)

### 1. **Auth Service** (Already Separated ✅)
- User registration, login
- JWT token management
- Password management
- **DB**: `gofiber_auth` (users, sessions)

### 2. **Social Service** (Core Social Features)
- Posts, Comments, Votes
- Tags, Saved Posts
- **DB**: Posts, Comments, Votes, Tags

### 3. **Profile Service**
- User profiles (displayName, avatar, bio)
- User stats (karma, followers count)
- Follow relationships
- **DB**: users_identity, user_profiles, follows

### 4. **Chat Service**
- 1-on-1 conversations
- Real-time messaging (WebSocket)
- **DB**: conversations, messages

### 5. **Notification Service**
- Real-time notifications (WebSocket)
- Notification preferences
- **DB**: notifications, notification_settings

### 6. **Media Service**
- File uploads (images, videos)
- Media processing (resize, encode)
- CDN integration
- **DB**: media, files

### 7. **Search Service** (Future)
- Full-text search (Elasticsearch/Meilisearch)
- Search indexing
- Search analytics

---

## 🎯 Migration Strategy

### Prerequisites for Microservice Split:

1. ✅ **Event-Driven Communication** (NATS) - Already implemented
2. ✅ **Separate Databases** - Auth already separated
3. ⚠️ **API Gateway** - Not yet (currently direct calls)
4. ⚠️ **Service Discovery** - Not yet
5. ⚠️ **Distributed Tracing** - Not yet

### Recommended Next Steps:

1. **Extract Chat Service** (Easiest first)
   - Already has clear boundaries (conversations, messages)
   - Minimal coupling with other features
   - Can communicate via NATS events

2. **Add API Gateway** (Kong, Traefik, or custom)
   - Centralized routing
   - Authentication/Authorization
   - Rate limiting

3. **Split Profile Service**
   - Separate user profile management
   - Follow system
   - Karma/stats calculation

4. **Extract Notification Service**
   - Real-time notifications via WebSocket
   - Push notifications

5. **Media Service**
   - Upload handling
   - Media processing
   - CDN integration

---

## 🔐 Current Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Backend
    participant Auth
    participant Redis

    Client->>Auth: POST /auth/login
    Auth->>Auth: Validate credentials
    Auth-->>Client: Return JWT token

    Client->>Backend: GET /api/v1/posts (with JWT)
    Backend->>Redis: Check token cache
    alt Token in cache
        Redis-->>Backend: Return user info
    else Token not in cache
        Backend->>Auth: Validate token (internal API)
        Auth-->>Backend: User info
        Backend->>Redis: Cache user info
    end
    Backend-->>Client: Return posts
```

---

## 📈 Performance Optimizations

### Current Optimizations:
- ✅ **Redis Caching** - Token validation, feed cache
- ✅ **Database Indexes** - Optimized queries
- ✅ **Cursor Pagination** - Efficient data loading
- ✅ **CDN** - Static file serving (R2, Bunny)
- ✅ **Connection Pooling** - Database connections
- ✅ **GORM Preloading** - Reduce N+1 queries

### Recommended Improvements:
- ⚠️ Add **Read Replicas** for database
- ⚠️ Implement **CQRS** pattern (read/write separation)
- ⚠️ Add **Message Queue** for async tasks
- ⚠️ Implement **Circuit Breaker** pattern

---

## 🏁 Summary

### Current State:
- ✅ **Clean Architecture** (Domain → Application → Infrastructure)
- ✅ **Event-Driven** (NATS for Auth Service communication)
- ✅ **Separated Auth** (Microservice ready)
- ✅ **Dual Database** (Auth + Social)
- ⚠️ **Monolith** (Social features in one service)

### Strengths:
- Clear separation of concerns
- Easy to test (dependency injection)
- Event-driven foundation for microservices
- Scalable data model

### Next Steps for Microservices:
1. Extract Chat Service
2. Add API Gateway
3. Implement service discovery
4. Add distributed tracing
5. Split remaining services gradually

---

**Last Updated**: 2025-11-24
**Author**: Generated by Claude Code
**Project**: SUEKK Social Media Platform
