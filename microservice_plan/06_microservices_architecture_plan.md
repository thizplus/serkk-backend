# 🏗️ Microservices Architecture Plan

**วางแผนแบบ Practical: ไม่แยกมากเกินไป แต่แยกพอดี**

---

## 🎯 หลักการออกแบบ

### ❌ สิ่งที่เราจะไม่ทำ
- ❌ แยก 21 services ทั้งหมด (มากเกินไป!)
- ❌ แยก PostService, CommentService, VoteService เป็นคนละ service (coupling สูง)
- ❌ ทำ Microservices ทั้งหมดพร้อมกัน (risky!)
- ❌ แยก database ทุก table ทันที (ซับซ้อนเกินไป)

### ✅ สิ่งที่เราจะทำ
- ✅ แยกตาม **Business Domain** (DDD - Domain-Driven Design)
- ✅ รวม services ที่เกี่ยวข้องกันแน่นไว้ด้วยกัน
- ✅ ใช้ **Event Bus** สำหรับ async communication
- ✅ Migration แบบ **Step-by-Step** (ทีละ service)
- ✅ เก็บ Monolith ไว้ก่อน แยกเฉพาะส่วนที่จำเป็น

---

## 📊 การแบ่ง Services (6 Core Services)

จาก 21 services เดิม → รวมเป็น **6 Core Microservices**

```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway                               │
│              (Routing, Auth, Rate Limiting)                  │
└─────────────────────────────────────────────────────────────┘
    │         │         │         │         │         │
    ▼         ▼         ▼         ▼         ▼         ▼
┌─────────┐ ┌────────┐ ┌────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐
│  User   │ │ Social │ │  Chat  │ │Notifica-│ │  Media  │ │ Search   │
│ Service │ │Service │ │Service │ │  tion   │ │ Service │ │ Service  │
│         │ │        │ │        │ │ Service │ │         │ │          │
│ ┌─────┐ │ │ ┌────┐ │ │ ┌────┐ │ │ ┌─────┐ │ │ ┌─────┐ │ │ ┌──────┐ │
│ │ PG  │ │ │ │ PG │ │ │ │Mongo│ │ │ │Redis│ │ │ │ S3  │ │ │ │Elastic││
│ └─────┘ │ │ └────┘ │ │ └────┘ │ │ └─────┘ │ │ └─────┘ │ │ └──────┘ │
└─────────┘ └────────┘ └────────┘ └─────────┘ └─────────┘ └──────────┘
    │         │         │         │         │         │
    └─────────┴─────────┴─────────┴─────────┴─────────┘
                        │
                ┌───────────────┐
                │  Event Bus    │
                │ (Kafka/NATS)  │
                └───────────────┘
```

---

## 🔹 Service 1: User Service

### 📌 Responsibility
จัดการข้อมูลผู้ใช้ทั้งหมด (Profile, Authentication sync)

### 🗂️ Owned Data
- `users_cache` (synced from Auth Service)
- `user_profiles` (social profile data)

### 🔧 API Endpoints
```
GET    /users/{userId}              - Get user info
GET    /users/{userId}/profile      - Get user profile
PUT    /users/{userId}/profile      - Update profile
GET    /users/{userId}/stats        - Get user stats
POST   /internal/users/sync         - Sync from Auth Service (webhook)
```

### 📤 Events Published
```
user.profile.updated
  → { userId, displayName, avatar, bio }

user.stats.updated
  → { userId, karma, followersCount, followingCount }
```

### 📥 Events Subscribed
```
post.created         → Update karma
comment.created      → Update karma
vote.created         → Update karma
follow.created       → Update followersCount/followingCount
follow.deleted       → Update followersCount/followingCount
```

### 💾 Database
**PostgreSQL** (separate database)
- Tables: `users_cache`, `user_profiles`

### 🔗 Dependencies
- Auth Service (external webhook)
- Event Bus

---

## 🔹 Service 2: Social Service

### 📌 Responsibility
จัดการฟีเจอร์ Social Network (Posts, Comments, Votes, Tags, Follows, Saved Posts)

**ทำไมไม่แยก?**
- Post, Comment, Vote มี coupling สูงมาก (ใช้ transaction ร่วมกัน)
- แยกแล้วจะต้องใช้ distributed transaction (ซับซ้อนมาก)
- Query ร่วมกันบ่อย (feed generation)

### 🗂️ Owned Data
- `posts` (posts, reposts)
- `comments` (nested comments)
- `votes` (votes on posts/comments)
- `tags` (tags, post_tags)
- `follows` (user relationships)
- `saved_posts` (bookmarks)
- `search_history` (search tracking)

### 🔧 API Endpoints
```
# Posts
POST   /posts                       - Create post
GET    /posts                       - List posts
GET    /posts/{postId}              - Get post
PUT    /posts/{postId}              - Update post
DELETE /posts/{postId}              - Delete post
GET    /feed                        - Get user feed

# Comments
POST   /posts/{postId}/comments     - Create comment
GET    /posts/{postId}/comments     - Get comments
PUT    /comments/{commentId}        - Update comment
DELETE /comments/{commentId}        - Delete comment

# Votes
POST   /votes                       - Create vote
DELETE /votes/{voteId}              - Delete vote

# Follows
POST   /users/{userId}/follow       - Follow user
DELETE /users/{userId}/follow       - Unfollow user
GET    /users/{userId}/followers    - Get followers
GET    /users/{userId}/following    - Get following

# Saved Posts
POST   /saved-posts                 - Save post
GET    /saved-posts                 - Get saved posts
DELETE /saved-posts/{postId}        - Remove saved post

# Tags
GET    /tags                        - List tags
GET    /tags/{tagId}                - Get tag
GET    /tags/{tagId}/posts          - Get posts by tag
```

### 📤 Events Published
```
post.created
  → { postId, authorId, title, content, type, tags[], mediaIds[] }

post.updated
  → { postId, title, content }

post.deleted
  → { postId, authorId }

comment.created
  → { commentId, postId, authorId, content, parentId }

vote.created
  → { voteId, userId, targetType, targetId, voteType }

follow.created
  → { followerId, followingId }

follow.deleted
  → { followerId, followingId }
```

### 📥 Events Subscribed
```
user.profile.updated  → Update cached user data in posts/comments
media.uploaded        → Link media to post
media.deleted         → Unlink media from post
```

### 💾 Database
**PostgreSQL** (separate database)
- Tables: `posts`, `comments`, `votes`, `tags`, `post_tags`, `post_media`, `follows`, `saved_posts`, `search_history`
- **Denormalized**: Author name/avatar cached in posts/comments (for performance)

### 📦 Redis Cache
```
feed:user:{userId}:cursor:{cursor}   → Feed cache (5 min TTL)
post:idempotency:{clientPostId}      → Idempotency (5 min TTL)
trending:tags                        → Trending tags (1 hour TTL)
```

### 🔗 Dependencies
- User Service (get user info via cache)
- Media Service (link media)
- Notification Service (via Event Bus)
- Event Bus

---

## 🔹 Service 3: Chat Service

### 📌 Responsibility
จัดการ Real-time Chat, Messages, Blocks, WebSocket connections

**ทำไมแยก?**
- Chat เป็น real-time (ต้องการ scalability แยก)
- WebSocket connections กิน memory สูง
- Database ต่างจาก Social (ใช้ MongoDB ได้)

### 🗂️ Owned Data
- `conversations` (chat threads)
- `messages` (chat messages)
- `blocks` (blocked users)

### 🔧 API Endpoints
```
# Conversations
POST   /conversations               - Create conversation
GET    /conversations               - List conversations
GET    /conversations/{convId}      - Get conversation
PUT    /conversations/{convId}/read - Mark as read

# Messages
POST   /conversations/{convId}/messages  - Send message
GET    /conversations/{convId}/messages  - Get messages
PUT    /messages/{messageId}             - Update message
DELETE /messages/{messageId}              - Delete message

# Blocks
POST   /blocks                      - Block user
GET    /blocks                      - List blocked users
DELETE /blocks/{blockedUserId}      - Unblock user

# WebSocket
WS     /ws/chat                     - Real-time chat
```

### 📤 Events Published
```
message.sent
  → { messageId, conversationId, senderId, receiverId, content, type, media }

message.read
  → { messageId, conversationId, readBy, readAt }

user.blocked
  → { blockerId, blockedId }
```

### 📥 Events Subscribed
```
user.profile.updated  → Update cached user data
media.uploaded        → Handle media in messages
```

### 💾 Database
**MongoDB** (document database)
- Collections: `conversations`, `messages`, `blocks`

**ทำไมใช้ MongoDB?**
- Messages เป็น append-only (เหมาะกับ document DB)
- Media field เป็น JSONB (MongoDB รองรับดี)
- ไม่มี complex joins
- Horizontal scaling ง่ายกว่า PostgreSQL

### 📦 Redis Cache
```
conversation:list:{userId}           → Conversation list (10 min TTL)
conversation:unread:{userId}         → Unread count
chat:ws:connections                  → WebSocket connection tracking
```

### 🔗 Dependencies
- User Service (get user info)
- Notification Service (via Event Bus)
- Media Service (handle media uploads)
- Event Bus

---

## 🔹 Service 4: Notification Service

### 📌 Responsibility
จัดการ Notifications, Web Push, WebSocket notifications

**ทำไมแยก?**
- Notification ถูกเรียกโดยทุก service (cross-cutting concern)
- Web Push ต้องการ background processing
- Scalability แยก (notification rate สูง)

### 🗂️ Owned Data
- `notifications` (activity notifications)
- `notification_settings` (user preferences)
- `push_subscriptions` (web push endpoints)

### 🔧 API Endpoints
```
GET    /notifications               - Get notifications
PUT    /notifications/mark-read     - Mark as read
GET    /notifications/settings      - Get settings
PUT    /notifications/settings      - Update settings
POST   /push/subscribe              - Subscribe web push
DELETE /push/unsubscribe            - Unsubscribe web push

# WebSocket
WS     /ws/notifications            - Real-time notifications
```

### 📤 Events Published
```
notification.created
  → { notificationId, userId, type, message }

notification.sent
  → { notificationId, userId, channels[] }
```

### 📥 Events Subscribed
```
post.created         → Notify followers
comment.created      → Notify post author
vote.created         → Notify target author
follow.created       → Notify followed user
message.sent         → Notify receiver
media.encoded        → Notify uploader (video ready)
```

### 💾 Database
**Redis** (fast, ephemeral)
- Keys: `notification:{userId}:*`
- Push subscriptions: Hash
- WebSocket connections: Set

**PostgreSQL** (archive, history)
- Tables: `notifications_archive`, `notification_settings`

**Strategy**: Recent notifications (7 days) in Redis, older in PostgreSQL

### 📦 Redis
```
notification:user:{userId}:recent    → Recent notifications (7 days)
notification:unread:{userId}         → Unread count
ws:notification:connections          → WebSocket connections
push:subscription:{userId}           → Push endpoints
```

### 🔗 Dependencies
- User Service (get user info)
- Web Push API (external)
- Event Bus

---

## 🔹 Service 5: Media Service

### 📌 Responsibility
จัดการ Upload, Storage, Video Encoding

**ทำไมแยก?**
- Media processing เป็น heavy operation (CPU/memory intensive)
- Video encoding ใช้เวลานาน (async)
- Scalability แยก (upload rate สูง)

### 🗂️ Owned Data
- `media` (media metadata)
- `files` (legacy files)

### 🔧 API Endpoints
```
POST   /media                       - Upload image/video
GET    /media/{mediaId}             - Get media info
DELETE /media/{mediaId}              - Delete media
POST   /upload/presigned            - Get presigned URL (R2)
POST   /webhook/bunny/video-completed - Video encoding callback
```

### 📤 Events Published
```
media.uploaded
  → { mediaId, userId, type, url, thumbnail, size }

media.encoded
  → { mediaId, videoId, hlsUrl, status }

media.deleted
  → { mediaId, url }
```

### 📥 Events Subscribed
```
post.deleted        → Delete unused media
message.deleted     → Delete unused media
```

### 💾 Database
**PostgreSQL** (metadata only)
- Tables: `media`, `files`

**Storage**:
- Images: Bunny CDN
- Videos: Bunny Stream (with HLS)
- Large files: Cloudflare R2

### 📦 Background Workers
```
VideoEncoderWorker
  → Poll Redis queue
  → Check Bunny Stream encoding status
  → Publish media.encoded event
```

### 🔗 Dependencies
- Bunny CDN (external)
- Bunny Stream (external)
- Cloudflare R2 (external)
- Event Bus

---

## 🔹 Service 6: Search Service

### 📌 Responsibility
Full-text search, Trending, Discovery

**ทำไมแยก?**
- Search queries ช้า (full-text search)
- ควรใช้ specialized database (Elasticsearch)
- Scalability แยก

### 🗂️ Owned Data
- Elasticsearch indexes (read models)

### 🔧 API Endpoints
```
GET    /search/posts                - Search posts
GET    /search/users                - Search users
GET    /search/tags                 - Search tags
GET    /search/trending             - Get trending content
GET    /search/history              - Get search history
```

### 📤 Events Published
```
search.performed
  → { userId, query, type, resultsCount }
```

### 📥 Events Subscribed
```
post.created         → Index post
post.updated         → Update index
post.deleted         → Remove from index
user.profile.updated → Update user index
tag.created          → Index tag
```

### 💾 Database
**Elasticsearch** (full-text search)
- Indexes: `posts`, `users`, `tags`

**PostgreSQL** (search history)
- Tables: `search_history`

### 🔗 Dependencies
- Social Service (via Event Bus)
- User Service (via Event Bus)
- Event Bus

---

## 🔹 (Optional) Service 7: Auto-Post Service

### 📌 Responsibility
AI-powered auto-posting (optional, low priority)

**ทำไมแยกหรือไม่แยก?**
- ✅ แยก: ถ้า AI generation มีปริมาณมาก
- ❌ ไม่แยก: ถ้าใช้น้อย, เก็บใน Social Service ก็ได้

### 🗂️ Owned Data
- `auto_post_settings`
- `auto_post_logs`

### 🔧 API Endpoints
```
GET    /auto-post/settings          - Get settings
PUT    /auto-post/settings          - Update settings
POST   /auto-post/generate          - Generate post manually
GET    /auto-post/logs              - Get generation logs
```

### 📤 Events Published
```
auto_post.generated
  → { logId, settingId, postId, topic, tokensUsed }
```

### 📥 Events Subscribed
```
(none)
```

### 💾 Database
**PostgreSQL**
- Tables: `auto_post_settings`, `auto_post_logs`

### 🔗 Dependencies
- OpenAI API (external)
- Social Service (create posts)
- Event Bus

**Recommendation**: เก็บใน Monolith ก่อน, แยกทีหลังถ้าจำเป็น

---

## 🚀 Event Bus Architecture

### เลือก Event Bus

**ตัวเลือก**:

| Technology | Pros | Cons | Recommendation |
|------------|------|------|----------------|
| **Kafka** | - Scalable, durable<br>- High throughput<br>- Event replay<br>- Industry standard | - Complex setup<br>- Heavy (requires ZooKeeper/KRaft)<br>- Overkill for small scale | ✅ Best for production |
| **NATS** | - Lightweight<br>- Fast<br>- Easy setup<br>- Good for microservices | - Less durable (JetStream needed)<br>- Smaller community | ✅ Good for small-medium scale |
| **RabbitMQ** | - Easy to use<br>- Good documentation<br>- Reliable | - Lower throughput than Kafka<br>- Single point of failure | ⚠️ OK but not ideal |
| **Redis Streams** | - Already using Redis<br>- Simple<br>- Fast | - Not as feature-rich<br>- Not designed for event bus | ⚠️ OK for simple use cases |

**แนะนำ**: **NATS with JetStream** (เริ่มต้น) → Migrate to **Kafka** (เมื่อ scale ใหญ่)

---

### Event Categories

#### 1. **Domain Events** (Business events)
```
post.created
post.updated
post.deleted
comment.created
vote.created
follow.created
message.sent
user.profile.updated
media.uploaded
notification.created
```

#### 2. **Integration Events** (Cross-service)
```
user.sync.requested       → User Service
cache.invalidate.feed     → Social Service
search.index.update       → Search Service
```

#### 3. **System Events** (Infrastructure)
```
service.health.check
service.metrics.report
```

---

### Event Schema Example

**Event: `post.created`**
```json
{
  "eventId": "uuid",
  "eventType": "post.created",
  "eventVersion": "v1",
  "timestamp": "2025-11-24T10:00:00Z",
  "producerService": "social-service",
  "data": {
    "postId": "uuid",
    "authorId": "uuid",
    "title": "Post Title",
    "content": "Post content...",
    "type": "text",
    "tags": ["tag1", "tag2"],
    "mediaIds": ["media-uuid-1"],
    "clientPostId": "idempotency-key"
  },
  "metadata": {
    "correlationId": "trace-id",
    "userId": "uuid",
    "requestId": "uuid"
  }
}
```

---

### Event Flow Example

**Scenario**: User creates a post

```
1. User → API Gateway → Social Service
2. Social Service:
   a. Validate request
   b. Create post in database
   c. Publish event: post.created
3. Event Bus broadcasts to subscribers:
   ├─► User Service: Update user.karma
   ├─► Notification Service: Notify followers
   ├─► Search Service: Index post
   └─► Media Service: Mark media as used
4. Social Service returns response to user (don't wait for events!)
```

**Key Principle**: **Fire and Forget** - Don't wait for event consumers

---

### Event Handling Patterns

#### Pattern 1: At-Least-Once Delivery
```
Event published → Kafka/NATS
  → Consumer receives event
  → Process event (idempotent!)
  → Acknowledge (ACK)
  → If fail → retry (with exponential backoff)
```

**Important**: Consumers must be **idempotent** (same event processed multiple times = same result)

#### Pattern 2: Dead Letter Queue (DLQ)
```
Event failed after max retries
  → Send to DLQ
  → Alert admin
  → Manual intervention or replay
```

#### Pattern 3: Event Sourcing (Optional)
```
Don't delete events, keep them forever
  → Replay events to rebuild state
  → Audit log
```

---

### Event Versioning

**Problem**: Events evolve over time

**Solution**: Version events

```json
// v1
{
  "eventType": "post.created",
  "eventVersion": "v1",
  "data": {
    "postId": "uuid",
    "title": "Post Title"
  }
}

// v2 (added new field)
{
  "eventType": "post.created",
  "eventVersion": "v2",
  "data": {
    "postId": "uuid",
    "title": "Post Title",
    "visibility": "public"  // NEW
  }
}
```

**Strategy**: Consumers support multiple versions (backward compatible)

---

## 🗺️ Migration Strategy (Step-by-Step)

### ❌ Don't Do This (Big Bang)
```
Week 1: Build all 6 microservices
Week 2: Migrate all data
Week 3: Switch to microservices
Week 4: Fix all the bugs (too many!)
```
**Risk**: Very high! Everything breaks at once.

---

### ✅ Do This (Strangler Fig Pattern)

**Principle**: Keep Monolith running, extract services one by one

```
Phase 0: Monolith (Current)
    │
    ├─► Phase 1: Extract Media Service (3-4 weeks)
    │       - Low risk, no tight coupling
    │       - Setup Event Bus
    │       - Learn microservices patterns
    │
    ├─► Phase 2: Extract Notification Service (3-4 weeks)
    │       - Connect to Event Bus
    │       - Migrate notifications
    │       - Keep WebSocket in Monolith (for now)
    │
    ├─► Phase 3: Extract Chat Service (4-6 weeks)
    │       - Migrate to MongoDB
    │       - Migrate WebSocket connections
    │       - High risk (real-time)
    │
    ├─► Phase 4: Extract User Service (3-4 weeks)
    │       - Separate user data
    │       - All services now use User Service API
    │
    ├─► Phase 5: Extract Search Service (2-3 weeks)
    │       - Setup Elasticsearch
    │       - Index existing data
    │
    └─► Phase 6: Social Service (Keep as Monolith or extract)
            - Option A: Keep as improved Monolith
            - Option B: Extract (4-6 weeks)
```

**Total Time**: ~6-9 months (depending on team size)

---

### Phase 1: Extract Media Service (Start Here!)

**Why start with Media Service?**
- ✅ Loosely coupled (mostly standalone)
- ✅ Clear boundaries (upload, storage)
- ✅ Learn microservices patterns (Event Bus, API Gateway)
- ✅ Low risk (if fails, just rollback)

**Steps**:

```
Week 1: Setup Infrastructure
  ├─ Setup NATS with JetStream
  ├─ Setup API Gateway (Kong/Traefik)
  ├─ Setup Docker Compose for local dev
  └─ Setup CI/CD pipeline

Week 2: Build Media Service
  ├─ Create new Go service
  ├─ Copy media handlers/services
  ├─ Connect to Event Bus
  ├─ Publish events: media.uploaded, media.deleted
  └─ Write tests

Week 3: Integration
  ├─ Monolith subscribes to media.uploaded events
  ├─ API Gateway routes /media/* to Media Service
  ├─ Deploy Media Service (separate container)
  └─ Monitor

Week 4: Validation & Rollback Plan
  ├─ Test all media upload scenarios
  ├─ Monitor performance
  ├─ If OK → keep, If not → rollback
  └─ Document learnings
```

**Rollback Plan**:
```
If Media Service fails:
  → API Gateway routes back to Monolith
  → Disable Media Service
  → Fix issues
  → Retry
```

---

### Phase 2-6: Continue Extraction

(Similar step-by-step plans for each service)

---

## 🏛️ Final Architecture

```
                    ┌──────────────────────────┐
                    │     Load Balancer        │
                    │      (Nginx/ALB)         │
                    └──────────────────────────┘
                               │
                    ┌──────────────────────────┐
                    │     API Gateway          │
                    │  (Kong/Traefik/KrakenD)  │
                    │                          │
                    │  - Routing               │
                    │  - Auth                  │
                    │  - Rate Limiting         │
                    │  - Request Validation    │
                    └──────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
    ┌──────────┐        ┌──────────┐        ┌──────────┐
    │  User    │        │ Social   │        │  Chat    │
    │ Service  │        │ Service  │        │ Service  │
    │          │        │          │        │          │
    │ ┌──────┐ │        │ ┌──────┐ │        │ ┌──────┐ │
    │ │ PG   │ │        │ │ PG   │ │        │ │Mongo │ │
    │ └──────┘ │        │ └──────┘ │        │ └──────┘ │
    └──────────┘        └──────────┘        └──────────┘
          │                    │                    │
          │                    │                    │
    ┌──────────┐        ┌──────────┐        ┌──────────┐
    │Notifica- │        │  Media   │        │ Search   │
    │   tion   │        │ Service  │        │ Service  │
    │ Service  │        │          │        │          │
    │ ┌──────┐ │        │ ┌──────┐ │        │ ┌──────┐ │
    │ │Redis │ │        │ │ S3   │ │        │ │Elastic│ │
    │ └──────┘ │        │ └──────┘ │        │ └──────┘ │
    └──────────┘        └──────────┘        └──────────┘
          │                    │                    │
          └────────────────────┼────────────────────┘
                               │
                    ┌──────────────────────────┐
                    │      Event Bus           │
                    │    (NATS/Kafka)          │
                    │                          │
                    │  Topics:                 │
                    │  - post.created          │
                    │  - user.profile.updated  │
                    │  - message.sent          │
                    │  - media.uploaded        │
                    │  - ...                   │
                    └──────────────────────────┘
```

---

## 📊 Database Strategy

### ❌ Before (Monolith)
```
Single PostgreSQL Database
  ├─ users_cache
  ├─ user_profiles
  ├─ posts
  ├─ comments
  ├─ conversations
  ├─ messages
  ├─ notifications
  └─ ... (27 tables)
```

### ✅ After (Microservices)
```
User Service DB (PostgreSQL)
  ├─ users_cache
  └─ user_profiles

Social Service DB (PostgreSQL)
  ├─ posts
  ├─ comments
  ├─ votes
  ├─ tags
  ├─ follows
  └─ saved_posts

Chat Service DB (MongoDB)
  ├─ conversations
  ├─ messages
  └─ blocks

Notification Service (Redis + PostgreSQL archive)
  ├─ notifications (Redis, 7 days)
  └─ notifications_archive (PostgreSQL)

Media Service DB (PostgreSQL)
  └─ media (metadata only, files in S3)

Search Service (Elasticsearch)
  ├─ posts_index
  ├─ users_index
  └─ tags_index
```

**Key Principle**: **Database per Service** (no shared database!)

---

## 🔒 Cross-Service Communication

### Rule 1: No Direct Database Access
```
❌ BAD:
Social Service → directly queries User Service database

✅ GOOD:
Social Service → calls User Service API or subscribes to events
```

---

### Rule 2: Sync vs Async

**Use Sync (REST API) when**:
- Need immediate response
- Read operations
- User is waiting

**Examples**:
```
Social Service → User Service API
  GET /users/{userId}  (sync, need user info now)
```

**Use Async (Events) when**:
- Fire and forget
- Write operations
- Side effects

**Examples**:
```
Social Service → publish post.created event
  → Notification Service subscribes (async, user doesn't wait)
  → Search Service subscribes (async, index later)
```

---

### Rule 3: Data Consistency

**Problem**: User profile changed in User Service, but Social Service has stale data

**Solution**: **Eventual Consistency** via Events

```
1. User updates profile in User Service
2. User Service publishes: user.profile.updated
3. Social Service subscribes:
   - Update cached user data in posts/comments table
4. Eventually consistent (delay: < 1 second)
```

**Important**: Users must accept eventual consistency (not immediate)

---

## 🛠️ Technology Stack (Updated)

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go | All microservices |
| **Framework** | Fiber v2 | HTTP APIs |
| **API Gateway** | Kong / Traefik | Routing, auth, rate limiting |
| **Event Bus** | NATS JetStream | Async communication |
| **Service Discovery** | Consul / NATS | Service registry |
| **Databases** | PostgreSQL, MongoDB, Redis | Per-service databases |
| **Search** | Elasticsearch | Full-text search |
| **Cache** | Redis | Distributed cache |
| **Storage** | S3 (R2/Bunny) | Object storage |
| **Monitoring** | Prometheus + Grafana | Metrics & dashboards |
| **Logging** | Loki + Promtail | Centralized logging |
| **Tracing** | Jaeger | Distributed tracing |
| **Container** | Docker | Containerization |
| **Orchestration** | Docker Compose (local) / Kubernetes (prod) | Container orchestration |
| **CI/CD** | GitHub Actions | Automated deployment |

---

## 📈 Observability (Critical!)

### 1. Metrics (Prometheus)

**Per-Service Metrics**:
```
http_requests_total{service="social-service", endpoint="/posts", method="POST"}
http_request_duration_seconds{service="social-service", endpoint="/posts"}
database_query_duration_seconds{service="social-service", query="insert_post"}
event_published_total{service="social-service", event="post.created"}
event_consumed_total{service="notification-service", event="post.created"}
```

---

### 2. Logging (Loki)

**Structured Logging**:
```json
{
  "timestamp": "2025-11-24T10:00:00Z",
  "service": "social-service",
  "level": "info",
  "message": "Post created",
  "postId": "uuid",
  "userId": "uuid",
  "correlationId": "trace-id"
}
```

---

### 3. Tracing (Jaeger)

**Distributed Trace Example**:
```
Request: POST /posts
├─ social-service: CreatePost (50ms)
│  ├─ PostgreSQL: INSERT post (10ms)
│  ├─ NATS: Publish post.created (5ms)
│  └─ Redis: Cache invalidation (5ms)
├─ notification-service: (async, triggered by event)
│  ├─ Get followers (20ms)
│  └─ Send notifications (30ms)
└─ search-service: (async)
   └─ Index post (100ms)

Total: 50ms (user sees response)
Async tasks: 150ms (happens in background)
```

---

## 💰 Cost Considerations

### Monolith vs Microservices Cost

**Monolith** (1 server):
```
1x Server (8 vCPU, 16 GB RAM): $80/month
1x PostgreSQL: included
1x Redis: included
───────────────────────────────────────
Total: ~$80/month
```

**Microservices** (6 services):
```
API Gateway: $20/month (2 vCPU, 4 GB)
User Service: $20/month (2 vCPU, 4 GB)
Social Service: $40/month (4 vCPU, 8 GB)
Chat Service: $40/month (4 vCPU, 8 GB)
Notification Service: $20/month (2 vCPU, 4 GB)
Media Service: $20/month (2 vCPU, 4 GB)
Search Service: $40/month (4 vCPU, 8 GB)
NATS: $10/month
Monitoring stack: $20/month
───────────────────────────────────────
Total: ~$230/month
```

**Cost Increase**: ~3x

**But**:
- ✅ Much better scalability
- ✅ Independent scaling (save money on low-traffic services)
- ✅ Better reliability (fault isolation)
- ✅ Faster development (parallel teams)

---

## ✅ Success Criteria

How do you know if migration is successful?

### Metrics to Track

| Metric | Before (Monolith) | Target (Microservices) |
|--------|-------------------|------------------------|
| **Deployment Time** | 20 min (entire app) | 3-5 min (per service) |
| **Deployment Frequency** | 1-2 times/week | Multiple times/day |
| **Recovery Time** | 10-30 min | 2-5 min (single service) |
| **P95 Latency** | 500ms | < 300ms |
| **Scalability** | Vertical only | Horizontal |
| **Development Speed** | Slow (coupling) | Fast (parallel teams) |
| **Incident Blast Radius** | Entire app | Single service |

---

## 🎯 Summary & Recommendations

### ✅ Recommended Approach

1. **Start Small**: Extract Media Service first (low risk)
2. **Use Event Bus**: NATS JetStream (easy) → Kafka (when needed)
3. **Database per Service**: Separate databases, no shared DB
4. **Eventual Consistency**: Accept slight delays (< 1 second)
5. **Observability First**: Metrics, logging, tracing from day 1
6. **Step-by-Step Migration**: 6-9 months, don't rush

---

### 🚦 Go/No-Go Decision

**Go for Microservices if**:
- ✅ Team size > 5 developers
- ✅ Need independent scaling
- ✅ Need independent deployments
- ✅ High traffic (> 1000 req/s)
- ✅ Long-term product (> 2 years)

**Stay with Monolith if**:
- ❌ Team size < 3 developers
- ❌ Low traffic (< 100 req/s)
- ❌ MVP / early stage
- ❌ Limited resources
- ❌ Short-term project

---

### 📌 Next Steps

1. **Review this plan** with your team
2. **Decide**: Microservices or Improved Monolith?
3. **If Microservices**: Start with Phase 1 (Media Service)
4. **Setup**: NATS, API Gateway, Docker Compose
5. **Execute**: Follow step-by-step migration plan
6. **Monitor**: Track metrics, learn, improve

---

**Good luck!** 🚀

มีคำถามเพิ่มเติมหรือต้องการความช่วยเหลือในการเริ่มต้น Phase 1 ไหมครับ?
