# 3️⃣ Current Interactions

สรุป Interactions และ Dependencies ระหว่าง Modules ใน Monolith ปัจจุบัน

---

## ภาพรวม Interaction Patterns

ระบบปัจจุบันใช้ **3 รูปแบบหลัก** ในการสื่อสารระหว่าง Modules:

1. **Direct Service Calls** - Service เรียก Service อื่นโดยตรง
2. **WebSocket Broadcasting** - Real-time notification ผ่าน WebSocket Hubs
3. **Webhook Integration** - รับข้อมูลจาก External Services

---

## 📋 Service Dependency Map

### 1. PostService Dependencies

**PostService เป็น Core Service ที่มี Dependencies มากที่สุด**

```
PostService
    ├─► TagService (create/update tags)
    ├─► MediaRepository (link media)
    ├─► VoteRepository (get vote counts)
    ├─► SavedPostRepository (check saved status)
    ├─► NotificationHub (WebSocket broadcast)
    ├─► Redis (idempotency + feed cache)
    └─► FeedCacheService (invalidate cache)
```

**Interaction Flow: การสร้างโพสต์**

```
1. User → PostHandler.CreatePost()
2. PostHandler → PostService.CreatePost()
3. PostService:
   a. Check idempotency (Redis)
      - Key: "post:idempotency:{client_post_id}"
      - TTL: 5 minutes
   b. Create post record (PostRepository)
   c. Process tags (TagService.CreateTag() for each)
   d. Link media (MediaRepository.LinkToPost())
   e. Invalidate feed cache (FeedCacheService)
   f. Notify followers:
      - Get follower list (FollowRepository)
      - Broadcast via NotificationHub (WebSocket)
      - Create notifications (NotificationService)
4. Return PostDTO to user
```

**Called By**:
- `AutoPostService` (AI-generated posts)
- `SimpleAutoPostService` (queue-based posts)

---

### 2. CommentService Dependencies

```
CommentService
    ├─► PostRepository (update comment count)
    ├─► VoteRepository (get vote counts)
    └─► NotificationService (notify post author)
```

**Interaction Flow: การสร้างคอมเมนต์**

```
1. User → CommentHandler.CreateComment()
2. CommentHandler → CommentService.CreateComment()
3. CommentService:
   a. Validate post exists (PostRepository)
   b. Check depth limit (max 10)
   c. Create comment (CommentRepository)
   d. Update post.comment_count (PostRepository)
   e. Notify post author:
      - CommentService → NotificationService.CreateNotification()
      - NotificationService → NotificationHub.Broadcast() (WebSocket)
4. Return CommentDTO
```

---

### 3. VoteService Dependencies

```
VoteService
    ├─► PostRepository (update vote count)
    ├─► CommentRepository (update vote count)
    ├─► UserRepository (get voter info)
    └─► NotificationService (notify author)
```

**Interaction Flow: การโหวต**

```
1. User → VoteHandler.CreateVote()
2. VoteHandler → VoteService.CreateVote()
3. VoteService:
   a. Check duplicate vote (VoteRepository)
   b. Create vote record
   c. Update target's vote count:
      - If target_type == "post" → PostRepository.UpdateVotes()
      - If target_type == "comment" → CommentRepository.UpdateVotes()
   d. Notify target author:
      - VoteService → NotificationService.CreateNotification()
      - NotificationService → NotificationHub.Broadcast() (WebSocket)
4. Return VoteDTO
```

---

### 4. FollowService Dependencies

```
FollowService
    ├─► UserRepository (get user info)
    └─► NotificationService (notify followed user)
```

**Interaction Flow: การติดตาม**

```
1. User → FollowHandler.CreateFollow()
2. FollowHandler → FollowService.CreateFollow()
3. FollowService:
   a. Check duplicate follow (FollowRepository)
   b. Create follow record
   c. Update user_profiles.followers_count (UserProfileRepository)
   d. Update user_profiles.following_count (UserProfileRepository)
   e. Notify followed user:
      - FollowService → NotificationService.CreateNotification()
      - NotificationService → NotificationHub.Broadcast() (WebSocket)
4. Return FollowDTO
```

---

### 5. MessageService Dependencies (มี Circular Dependency)

```
MessageService
    ├─► ConversationRepository (update last_message, unread_count)
    ├─► BlockRepository (check blocked status)
    ├─► UserRepository (get user info)
    ├─► Redis (cache conversation list)
    ├─► ChatHub (WebSocket broadcast) ◄─┐
    └─► PushService (web push notification) │
                                             │
ChatHub ─────────────────────────────────────┘
    └─► MessageService (injected via setter)
```

**Interaction Flow: การส่งข้อความ**

```
1. User → MessageHandler.SendMessage()
2. MessageHandler → MessageService.SendMessage()
3. MessageService:
   a. Check blocked status (BlockRepository)
   b. Get/Create conversation (ConversationService)
   c. Create message (MessageRepository)
   d. Update conversation:
      - last_message_id = new message ID
      - receiver_unread_count += 1
      - last_message_at = now
   e. Broadcast to receiver:
      - MessageService → ChatHub.BroadcastMessage() (WebSocket)
   f. Send push notification:
      - MessageService → PushService.SendPushNotification()
4. Return MessageDTO
```

**Circular Dependency Resolution**:
```go
// In container.go
chatHub := infrastructure.NewChatHub()
messageService := serviceimpl.NewMessageService(...)
chatHub.SetMessageService(messageService) // Setter injection
```

---

### 6. NotificationService Dependencies (มี Circular Dependency)

```
NotificationService
    ├─► NotificationRepository (save notification)
    ├─► NotificationSettingsRepository (check user preferences)
    ├─► UserRepository (get user info)
    ├─► NotificationHub (WebSocket broadcast) ◄─┐
    └─► PushService (web push notification) ◄───┤
                                                 │
NotificationHub ──────────────────────────────────┤
    └─► NotificationService (injected via setter) │
                                                 │
PushService ──────────────────────────────────────┘
    └─► NotificationService (injected via setter)
```

**Interaction Flow: การสร้าง Notification**

```
1. Any Service → NotificationService.CreateNotification()
2. NotificationService:
   a. Check user preferences (NotificationSettingsRepository)
   b. If disabled → return early
   c. Create notification (NotificationRepository)
   d. Broadcast to user:
      - NotificationService → NotificationHub.Broadcast() (WebSocket)
   e. Send push notification:
      - NotificationService → PushService.SendPushNotification()
3. Return NotificationDTO
```

---

### 7. AutoPostService Dependencies

```
AutoPostService
    ├─► AutoPostSettingRepository (get settings)
    ├─► AutoPostLogRepository (save logs)
    ├─► PostService (create posts)
    └─► OpenAIService (generate content)
```

**Interaction Flow: การสร้างโพสต์อัตโนมัติ**

```
1. EventScheduler (Cron: every hour) → AutoPostService.ProcessAllEnabledSettings()
2. AutoPostService:
   a. Get enabled settings (AutoPostSettingRepository)
   b. For each setting:
      i. Pick random topic from settings.topics
      ii. Generate content:
          - AutoPostService → OpenAIService.GeneratePost()
          - OpenAI API (gpt-4o-mini)
      iii. Create post:
          - AutoPostService → PostService.CreatePost()
      iv. Save log (AutoPostLogRepository)
      v. Update total_posts_generated
3. Return summary
```

---

### 8. VideoEncoderWorker Dependencies

```
VideoEncoderWorker
    ├─► RedisService (poll encoding queue)
    ├─► BunnyStreamService (check encoding status)
    ├─► MediaRepository (update media record)
    ├─► NotificationService (notify user)
    ├─► PostService (if video used in post)
    └─► NotificationHub (WebSocket broadcast)
```

**Interaction Flow: การเข้ารหัสวิดีโอ (Async)**

```
1. User uploads video → MediaUploadService
2. MediaUploadService:
   a. Upload to Bunny Stream
   b. Queue encoding task:
      - Redis LPUSH "video:encoding:queue" {media_id}
3. VideoEncoderWorker (goroutine polls every 10s):
   a. Redis RPOP "video:encoding:queue"
   b. Check encoding status (BunnyStreamService API)
   c. If completed:
      i. Update media record (MediaRepository)
      ii. Notify user:
          - VideoEncoderWorker → NotificationService
          - NotificationHub.Broadcast() (WebSocket)
      iii. If used in post → PostService.UpdatePost()
4. If still encoding → re-queue to Redis
```

---

## 🌐 WebSocket Interaction Patterns

### ChatHub (Real-time Chat)

```
ChatHub (runs in goroutine)
    ├─► Manages WebSocket connections per user
    ├─► Broadcast message to receiver
    └─► Called by: MessageService
```

**Connection Flow**:
```
1. User connects: ws://localhost:3000/ws/chat?token={jwt}
2. ChatHub.Register(userID, conn)
3. MessageService → ChatHub.BroadcastMessage(receiverID, message)
4. ChatHub sends message to receiver's WebSocket
```

**Data Flow**:
```
MessageService.SendMessage()
    → ChatHub.BroadcastMessage()
    → WebSocket.WriteJSON(message)
    → Receiver's browser
```

---

### NotificationHub (Real-time Notifications)

```
NotificationHub (runs in goroutine)
    ├─► Manages WebSocket connections per user
    ├─► Broadcast notification to user
    └─► Called by: NotificationService, PostService, VideoEncoderWorker
```

**Connection Flow**:
```
1. User connects: ws://localhost:3000/ws/notifications?token={jwt}
2. NotificationHub.Register(userID, conn)
3. NotificationService → NotificationHub.Broadcast(userID, notification)
4. NotificationHub sends to user's WebSocket
```

**Notification Triggers**:
```
Event                      → Service                → NotificationHub
─────────────────────────────────────────────────────────────────────
New comment                → CommentService         → Broadcast to post author
New vote                   → VoteService            → Broadcast to target author
New follower               → FollowService          → Broadcast to followed user
New message                → MessageService         → Broadcast to receiver
Video encoding complete    → VideoEncoderWorker     → Broadcast to uploader
New post from followed     → PostService            → Broadcast to all followers
```

---

## 🔗 External Service Integrations

### 1. Auth Service Integration (Webhook)

**Purpose**: ซิงค์ข้อมูลผู้ใช้จาก Auth Service

```
Auth Service (External)
    → Webhook POST /internal/webhooks/user-sync
    → InternalHandler.HandleUserSync()
    → UsersCacheService.CreateOrUpdateUser()
    → users_cache table
```

**Webhook Payload**:
```json
{
  "event": "user.created" | "user.updated" | "user.deleted",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "display_name": "Display Name",
    "avatar": "https://cdn.example.com/avatar.jpg",
    "role": "user",
    "is_active": true
  }
}
```

**Flow**:
```
1. User registers in Auth Service
2. Auth Service triggers webhook
3. This service receives webhook
4. UsersCacheService updates users_cache
5. All services use users_cache as read model
```

---

### 2. Bunny Stream Integration (Webhook)

**Purpose**: รับการแจ้งเตือนเมื่อการเข้ารหัสวิดีโอเสร็จสิ้น

```
Bunny Stream (External)
    → Webhook POST /webhook/bunny/video-completed
    → WebhookHandler.HandleVideoCompleted()
    → MediaService.UpdateVideoStatus()
    → NotificationService (notify uploader)
```

**Webhook Payload**:
```json
{
  "VideoGuid": "bunny-stream-video-id",
  "Status": "completed",
  "EncodingProgress": 100
}
```

**Flow**:
```
1. User uploads video → Bunny Stream
2. Bunny encodes video (async)
3. Bunny triggers webhook when done
4. MediaService updates media record
5. NotificationHub notifies user via WebSocket
```

---

### 3. OpenAI API Integration (Request-Response)

**Purpose**: สร้างเนื้อหาโพสต์ด้วย AI

```
AutoPostService
    → OpenAIService.GeneratePost()
    → HTTP POST https://api.openai.com/v1/chat/completions
    → Parse response
    → Return generated content
```

**Request Flow**:
```
1. EventScheduler triggers AutoPostService (hourly)
2. AutoPostService picks random topic
3. OpenAIService calls OpenAI API
4. Get generated title + content
5. PostService creates post
6. AutoPostLogRepository saves log
```

---

### 4. Storage Services Integration

```
MediaUploadService
    ├─► Bunny CDN (images, files)
    │   └─► HTTP PUT https://storage.bunnycdn.com/{path}
    ├─► Bunny Stream (videos)
    │   └─► HTTP POST https://video.bunnycdn.com/library/{libraryId}/videos
    └─► Cloudflare R2 (optional, large files)
        └─► S3-compatible API
```

---

## 🔄 Data Flow Patterns

### Pattern 1: Feed Generation (Complex Query)

```
GET /api/v1/feed
    → PostHandler.GetFeedPosts()
    → PostService.GetFeedPosts()
    → Check cache (FeedCacheService)
    → If miss:
        a. Get followed users (FollowRepository)
        b. Get posts from followed users (PostRepository)
        c. Join with users_cache (author info)
        d. Join with post_media → media (images/videos)
        e. Join with post_tags → tags
        f. Get vote status (VoteRepository)
        g. Get saved status (SavedPostRepository)
        h. Cache result (FeedCacheService, 5 min TTL)
    → Return paginated feed
```

**Repositories Involved**:
- PostRepository
- FollowRepository
- UserRepository (users_cache)
- MediaRepository
- TagRepository
- VoteRepository
- SavedPostRepository

---

### Pattern 2: Cache Invalidation (Write-through)

```
PostService.UpdatePost()
    → Update post (PostRepository)
    → Invalidate caches:
        a. FeedCacheService.InvalidateFeedCache(author_id)
        b. FeedCacheService.InvalidateFollowerFeeds(author_id)
        c. Redis DEL "feed:user:{user_id}:*"
```

**Cache Keys**:
```
feed:user:{user_id}:cursor:{cursor}
post:idempotency:{client_post_id}
conversation:list:{user_id}
```

---

### Pattern 3: Denormalization Update (Consistency)

```
VoteService.CreateVote()
    → Transaction:
        a. Insert vote (VoteRepository)
        b. Update target's vote count:
           - posts.votes += 1 (if upvote)
           - comments.votes += 1
        c. Commit transaction
```

**Denormalized Fields** (updated in same transaction):
- `posts.votes`
- `posts.comment_count`
- `comments.votes`
- `user_profiles.followers_count`
- `user_profiles.following_count`
- `tags.post_count`
- `conversations.last_message_id`
- `conversations.user1_unread_count`
- `conversations.user2_unread_count`

---

## 📊 Service Call Frequency (Estimated)

| Service | Called By | Frequency |
|---------|-----------|-----------|
| `NotificationService` | PostService, CommentService, VoteService, FollowService, MessageService | Very High |
| `UserRepository` (users_cache) | All services | Very High |
| `PostService` | AutoPostService, SimpleAutoPostService, UI | High |
| `TagService` | PostService | High |
| `MediaService` | PostService, MessageService | Medium |
| `VoteService` | UI | Medium |
| `CommentService` | UI | Medium |
| `MessageService` | UI | Medium |
| `FollowService` | UI | Low |
| `AutoPostService` | EventScheduler (hourly) | Low |

---

## 🎯 Key Insights

### 1. Tight Coupling
- **PostService** มี dependencies มากที่สุด (7 dependencies)
- **NotificationService** ถูกเรียกโดยเกือบทุก service
- **users_cache** ถูกใช้โดยทุก service (shared database)

### 2. Circular Dependencies
- `MessageService` ↔ `ChatHub`
- `NotificationService` ↔ `NotificationHub`
- `NotificationService` ↔ `PushService`
- แก้ไขด้วย Setter Injection

### 3. Real-time Communication
- ใช้ **2 WebSocket Hubs**: ChatHub, NotificationHub
- ทำงานใน separate goroutines
- Broadcast แบบ one-to-one (user-specific)

### 4. External Dependencies
- **Auth Service**: Webhook-based sync (eventual consistency)
- **Bunny Stream**: Async video encoding (webhook callback)
- **OpenAI**: Synchronous API calls (timeout risk)

### 5. Caching Strategy
- **Redis** ใช้สำหรับ: Idempotency, Feed Cache, Session
- **Manual invalidation** เมื่อมีการเปลี่ยนแปลงข้อมูล
- **TTL**: 5 minutes for feed cache

### 6. Database Access Patterns
- **Heavy reads**: posts, users_cache, follows
- **Heavy writes**: notifications, messages, votes
- **Complex joins**: feed generation, search queries
- **Denormalized counts**: เพื่อลด joins

---

## 🚨 Identified Issues

1. **Shared Database Bottleneck**
   - ทุก service access PostgreSQL เดียวกัน
   - ไม่สามารถ scale แยกกันได้

2. **Synchronous External Calls**
   - OpenAI API calls block requests
   - ควรใช้ async/queue

3. **WebSocket Hub Coupling**
   - Hubs ผูกติดกับ Services แน่น
   - ควรแยกเป็น Notification Service

4. **Manual Cache Invalidation**
   - ง่ายต่อการพลาด invalidate
   - ควรใช้ Event-driven invalidation

5. **Complex Feed Query**
   - ต้อง join 7+ tables
   - ควรใช้ Read Model/CQRS
