# 4️⃣ Pain Points / Limitations

สรุปปัญหาและข้อจำกัดของ Monolith ปัจจุบัน

---

## ภาพรวมปัญหา

ระบบปัจจุบันพบ **6 กลุ่มปัญหาหลัก**:

1. **Scalability Issues** - ปัญหาในการ Scale
2. **Performance Bottlenecks** - คอขวดด้านประสิทธิภาพ
3. **High Coupling** - การพึ่งพาระหว่าง Modules สูง
4. **Deployment Challenges** - ปัญหาในการ Deploy
5. **Development Complexity** - ความซับซ้อนในการพัฒนา
6. **Operational Risks** - ความเสี่ยงในการดำเนินงาน

---

## 1️⃣ Scalability Issues

### 🚨 Problem 1.1: Cannot Scale Services Independently

**ปัญหา**:
- ทุก service รันในกระบวนการเดียวกัน
- ไม่สามารถ scale เฉพาะ service ที่มี load สูงได้
- ต้อง scale ทั้ง application แม้ว่าจะต้องการแค่ส่วนใดส่วนหนึ่ง

**ผลกระทบ**:
```
Scenario: Chat service มี users 10,000 concurrent connections
         แต่ Auto-Post service ใช้ทรัพยากรแค่ 1 request/hour

Problem: ต้อง scale ทั้ง monolith → waste resources
         Chat Hub กิน RAM สูง → ส่งผลกับ Post Service ด้วย
```

**ตัวอย่างการใช้ทรัพยากร**:
| Service | CPU | Memory | Load |
|---------|-----|--------|------|
| Chat (WebSocket) | 5% | 2 GB | High |
| Post Service | 30% | 500 MB | High |
| Auto-Post Service | 1% | 50 MB | Low |
| **Total** | **36%** | **2.5 GB** | **Must scale all** |

**ถ้าเป็น Microservices**:
- Scale เฉพาะ Chat Service (horizontal)
- Scale เฉพาะ Post Service (horizontal)
- Auto-Post Service ใช้ 1 instance เท่านั้น

---

### 🚨 Problem 1.2: Shared Database Bottleneck

**ปัญหา**:
- ทุก service ใช้ PostgreSQL database เดียวกัน
- Write operations แข่งขัน connection pool
- Read operations ทำให้ queries ช้า

**ผลกระทบ**:
```
Database Connections:
- Max connections: 100
- PostService: 30 connections
- MessageService: 20 connections
- CommentService: 15 connections
- VoteService: 10 connections
- NotificationService: 10 connections
- Others: 15 connections
───────────────────────────────────
Total: 100 connections (maxed out!)
```

**Connection Pool Exhaustion**:
```
Scenario: มี spike traffic ใน PostService (สร้างโพสต์เยอะ)
Result: Connection pool เต็ม → MessageService timeout
Impact: Chat ใช้งานไม่ได้แม้ว่า Chat load ปกติ
```

**Slow Queries**:
```sql
-- Feed generation query (ช้า!)
SELECT p.*, u.*, m.*, t.*
FROM posts p
JOIN users_cache u ON p.author_id = u.id
LEFT JOIN post_media pm ON p.id = pm.post_id
LEFT JOIN media m ON pm.media_id = m.id
LEFT JOIN post_tags pt ON p.id = pt.post_id
LEFT JOIN tags t ON pt.tag_id = t.id
WHERE p.author_id IN (SELECT following_id FROM follows WHERE follower_id = $1)
ORDER BY p.created_at DESC
LIMIT 20;

-- ต้อง join 7 tables!
-- Query time: ~500ms (slow!)
```

---

### 🚨 Problem 1.3: Resource Contention

**ปัญหา**:
- Services แข่งขันใช้ CPU, Memory, Disk I/O
- WebSocket Hubs กิน memory สูง (ถ้ามี users เยอะ)
- Background workers (VideoEncoderWorker, EventScheduler) แย่งทรัพยากร

**ตัวอย่าง**:
```
Normal time:
- PostService: 100 req/s → CPU 20%
- ChatHub: 5,000 connections → Memory 1.5 GB

Peak time:
- PostService: 500 req/s → CPU 60%
- ChatHub: 20,000 connections → Memory 6 GB
- VideoEncoderWorker: 100 videos encoding → CPU 20%
───────────────────────────────────────────────────
Total: CPU 80%, Memory 7.5 GB → app crashes (OOM)
```

---

## 2️⃣ Performance Bottlenecks

### 🚨 Problem 2.1: Complex Database Queries

**ปัญหา**:
- Feed generation ต้อง join หลาย tables
- N+1 query problem ใน comments
- Denormalized counts ต้อง update หลาย tables

**Feed Query Performance**:
```
GET /api/v1/feed
├─ Get followed users (1 query)
├─ Get posts (1 query with 7 joins!)
│  ├─ posts
│  ├─ users_cache (author)
│  ├─ post_media → media
│  ├─ post_tags → tags
│  ├─ votes (aggregate)
│  ├─ saved_posts (check saved)
│  └─ comments (aggregate)
├─ Total time: ~500ms per request
└─ Cache miss rate: ~30% (feed invalidated often)
```

**N+1 Query Problem**:
```go
// ❌ Bad: N+1 queries
comments := GetCommentsByPostID(postID) // 1 query
for _, comment := range comments {
    author := GetUserByID(comment.AuthorID) // N queries!
    votes := GetVotesByCommentID(comment.ID) // N queries!
}

// Total queries: 1 + (N * 2) = 201 queries for 100 comments!
```

---

### 🚨 Problem 2.2: Cache Invalidation Complexity

**ปัญหา**:
- Manual cache invalidation → easy to miss
- Feed cache invalidation ต้องทำหลายจุด
- Stale data ถ้าลืม invalidate

**Invalidation Points**:
```go
// เมื่อสร้างโพสต์ใหม่ → ต้อง invalidate:
1. Author's feed cache
2. All followers' feed caches (could be 10,000+!)
3. Tag page caches
4. Home page cache
5. Trending posts cache

// ถ้าลืม invalidate → users เห็น old data
```

**Problem Scenario**:
```
1. User A posts → invalidate A's feed ✓
2. User A posts → invalidate followers' feeds ✗ (forgot!)
3. User B (follower) sees old feed (missing new post)
4. User B refreshes → still old (cached for 5 min)
5. After 5 min → new post appears (cache expired)
```

---

### 🚨 Problem 2.3: Synchronous External API Calls

**ปัญหา**:
- OpenAI API calls block requests
- Bunny Stream upload blocks until complete
- No retry mechanism

**OpenAI Latency**:
```
AutoPostService.GeneratePost()
    → OpenAI API call (synchronous)
    → Average: 3-5 seconds
    → Max: 30 seconds (timeout)
    → Blocks EventScheduler thread
    → If fails → no retry → log error only
```

**Impact**:
```
Scenario: OpenAI API down (outage)
Result: AutoPostService fails → EventScheduler stops
Impact: All scheduled tasks blocked (other jobs can't run)
```

---

## 3️⃣ High Coupling

### 🚨 Problem 3.1: Service Dependencies

**ปัญหา**:
- PostService มี 7 dependencies
- NotificationService ถูกเรียกโดย 5 services
- Circular dependencies ระหว่าง MessageService ↔ ChatHub

**Dependency Graph**:
```
PostService
    ├─► TagService
    ├─► MediaRepository
    ├─► VoteRepository
    ├─► SavedPostRepository
    ├─► NotificationHub
    ├─► Redis
    └─► FeedCacheService

NotificationService ◄─── Called by:
    ├─ PostService
    ├─ CommentService
    ├─ VoteService
    ├─ FollowService
    └─ MessageService
```

**Impact**:
- เปลี่ยน NotificationService → ต้อง test 5 services
- เปลี่ยน PostService → ต้อง update 7 dependencies
- Circular dependency → ยาก maintain

---

### 🚨 Problem 3.2: Shared Database Schema

**ปัญหา**:
- ทุก service access ทุก table
- ไม่มีการแบ่ง database boundary
- Schema change ส่งผลกับทุก service

**Coupling Example**:
```
users_cache table ถูกใช้โดย:
    ├─ PostService (author info)
    ├─ CommentService (author info)
    ├─ MessageService (sender/receiver info)
    ├─ NotificationService (user info)
    ├─ VoteService (voter info)
    ├─ FollowService (follower/following info)
    └─ ... (all services!)

Problem: เปลี่ยน users_cache schema → ต้อง update ทุก service!
```

**Migration Risk**:
```sql
-- Add column to users_cache
ALTER TABLE users_cache ADD COLUMN phone_number VARCHAR;

-- Impact:
-- ✗ PostService: need to update DTOs
-- ✗ CommentService: need to update DTOs
-- ✗ MessageService: need to update DTOs
-- ✗ ... (all services need changes!)
```

---

### 🚨 Problem 3.3: Monolithic Deployment

**ปัญหา**:
- เปลี่ยน 1 service → ต้อง deploy ทั้ง app
- Downtime ทุกครั้งที่ deploy
- Rollback ยาก (ต้อง rollback ทั้งหมด)

**Deployment Scenario**:
```
Change: แก้ bug ใน CommentService

Required:
1. Rebuild entire application
2. Stop entire application
3. Deploy new version
4. Restart application
5. All services down during deploy (~2-5 minutes)

Impact:
- Chat users disconnected (WebSocket closed)
- Feed loading fails
- Notification delivery stopped
- Message sending fails
```

---

## 4️⃣ Deployment Challenges

### 🚨 Problem 4.1: All-or-Nothing Deployment

**ปัญหา**:
- ไม่สามารถ deploy service ทีละตัวได้
- Risk สูง (ถ้า deploy fail → ทุกอย่าง fail)
- Rollback ทั้งหมด (ไม่สามารถ rollback เฉพาะ service)

**Deployment Risk**:
```
Scenario: Deploy new feature in PostService

Risk Level: HIGH
- If PostService has bug → entire app crashes
- If database migration fails → entire app can't start
- If new code has memory leak → all services affected
```

---

### 🚨 Problem 4.2: Long Build Time

**ปัญหา**:
- Build ทั้ง monolith ทุกครั้ง
- Test ทั้ง monolith ทุกครั้ง
- CI/CD pipeline ช้า

**Build Time**:
```
Current Build Time:
├─ Compile Go code: ~2 minutes
├─ Run unit tests: ~5 minutes
├─ Run integration tests: ~10 minutes
├─ Build Docker image: ~3 minutes
└─ Total: ~20 minutes

ถ้าเป็น Microservices:
└─ Build only changed service: ~3 minutes
```

---

### 🚨 Problem 4.3: No Gradual Rollout

**ปัญหา**:
- ไม่สามารถ deploy แบบ canary/blue-green ได้
- ไม่สามารถ A/B test features ได้
- ทุก user ได้ version เดียวกัน

---

## 5️⃣ Development Complexity

### 🚨 Problem 5.1: Large Codebase

**ปัญหา**:
- Developer ต้องเข้าใจ entire codebase
- Hard to onboard new developers
- Difficult to navigate code

**Codebase Size**:
```
Total:
├─ 21 Services
├─ 27 Database Tables
├─ 80+ API Endpoints
├─ 50+ Repository Implementations
├─ 10+ Background Workers/Hubs
└─ ~30,000 lines of code

New developer must understand:
- All service interactions
- All database relationships
- All API contracts
- All business logic
```

---

### 🚨 Problem 5.2: Difficult to Add New Features

**ปัญหา**:
- ต้องเข้าใจ existing services ก่อน
- ต้อง worry about breaking existing features
- ต้อง coordinate changes across multiple modules

**Example**:
```
Task: Add "Story" feature (like Instagram Stories)

Required Changes:
1. Add story table
2. Add StoryService
3. Update NotificationService (notify followers)
4. Update MediaService (handle story media)
5. Update PostService (promote story to post)
6. Update UserProfileService (story count)
7. Update FeedService (show stories in feed)
8. Update ChatHub (notify real-time)
9. Update Cache invalidation logic
10. Update all DTOs

Estimated time: 2-3 weeks (because of coupling!)
```

---

### 🚨 Problem 5.3: Testing Complexity

**ปัญหา**:
- ต้อง test entire application
- Integration tests ช้า
- Flaky tests (dependencies on external services)

**Testing Challenges**:
```
Unit tests: OK (fast, isolated)

Integration tests: SLOW
├─ Need full database setup
├─ Need Redis setup
├─ Need mock external APIs (OpenAI, Bunny)
├─ Need WebSocket connections
└─ Total test time: ~10 minutes

E2E tests: VERY SLOW
└─ Total test time: ~30 minutes
```

---

## 6️⃣ Operational Risks

### 🚨 Problem 6.1: Single Point of Failure

**ปัญหา**:
- ถ้า application crashes → ทุกอย่าง down
- ถ้า database down → ทุกอย่าง down
- No fault isolation

**Failure Scenarios**:
```
Scenario 1: Memory leak in ChatHub
Result: Application OOM → entire app crashes
Impact: All services down (posts, comments, notifications, chat)

Scenario 2: Slow query in PostService
Result: Database connection pool exhausted
Impact: All services timeout (chat, notifications, etc.)

Scenario 3: PostgreSQL crashes
Result: Entire application can't function
Impact: 100% downtime
```

---

### 🚨 Problem 6.2: No Service-Level Monitoring

**ปัญหา**:
- ไม่สามารถ monitor แยก service ได้
- ไม่รู้ว่า service ไหนทำให้ app ช้า
- ยากต่อการ debug performance issues

**Current Monitoring**:
```
Available:
- Application-level metrics (CPU, Memory, Requests/s)
- Database metrics

Missing:
- PostService-specific metrics
- CommentService latency
- ChatHub connection count per service
- NotificationService delivery rate
```

---

### 🚨 Problem 6.3: No Circuit Breaker

**ปัญหา**:
- ถ้า external service down → requests timeout
- ไม่มี fallback mechanism
- ไม่มี retry logic

**Failure Propagation**:
```
Scenario: OpenAI API down

Current behavior:
1. AutoPostService calls OpenAI → timeout (30s)
2. EventScheduler blocked for 30s
3. All other scheduled jobs delayed
4. No fallback → just log error
5. No retry → lost opportunity to generate post

Better behavior (with Circuit Breaker):
1. OpenAI API down → circuit opens
2. AutoPostService skips OpenAI call (fail fast)
3. EventScheduler continues other jobs
4. Retry after cooldown period
```

---

## 📊 สรุปปัญหาทั้งหมด

| Category | Problem | Severity | Impact |
|----------|---------|----------|--------|
| **Scalability** | Cannot scale independently | 🔴 Critical | High cost, resource waste |
| **Scalability** | Shared database bottleneck | 🔴 Critical | Performance degradation |
| **Scalability** | Resource contention | 🟡 High | OOM crashes |
| **Performance** | Complex database queries | 🟡 High | Slow responses |
| **Performance** | Cache invalidation complexity | 🟡 High | Stale data |
| **Performance** | Synchronous external calls | 🟡 High | Blocking, timeouts |
| **Coupling** | Service dependencies | 🟡 High | Hard to maintain |
| **Coupling** | Shared database schema | 🟡 High | Risky migrations |
| **Coupling** | Monolithic deployment | 🔴 Critical | High deployment risk |
| **Deployment** | All-or-nothing deployment | 🔴 Critical | Downtime |
| **Deployment** | Long build time | 🟢 Medium | Slow CI/CD |
| **Deployment** | No gradual rollout | 🟡 High | Can't A/B test |
| **Development** | Large codebase | 🟢 Medium | Slow onboarding |
| **Development** | Difficult to add features | 🟡 High | Slow development |
| **Development** | Testing complexity | 🟢 Medium | Slow tests |
| **Operations** | Single point of failure | 🔴 Critical | 100% downtime risk |
| **Operations** | No service-level monitoring | 🟡 High | Hard to debug |
| **Operations** | No circuit breaker | 🟡 High | Cascading failures |

**Legend**:
- 🔴 Critical: Must fix immediately
- 🟡 High: Should fix soon
- 🟢 Medium: Nice to have

---

## 🎯 Top 5 Most Critical Issues

1. **Shared Database Bottleneck** → Need database per service
2. **Cannot Scale Independently** → Need service separation
3. **Single Point of Failure** → Need fault isolation
4. **All-or-Nothing Deployment** → Need independent deployment
5. **Service Dependencies** → Need decoupling via events

---

## 💡 Recommended Solutions (High-Level)

| Problem | Solution |
|---------|----------|
| Shared database | Database per service pattern |
| Cannot scale independently | Microservices architecture |
| Complex queries | CQRS + Read Models |
| Cache invalidation | Event-driven invalidation |
| Synchronous external calls | Async message queue |
| Service dependencies | Event-driven architecture |
| All-or-nothing deployment | Independent service deployment |
| Single point of failure | Service mesh + circuit breaker |
| No monitoring | Service-level observability |

---

**Next Step**: ใช้ข้อมูลนี้เพื่อออกแบบ Microservices Architecture ที่แก้ปัญหาเหล่านี้
