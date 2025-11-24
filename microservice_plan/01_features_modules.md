# 1️⃣ Features / Modules

สรุป Features และ Modules ทั้งหมดใน Monolith ปัจจุบัน

---

## ภาพรวม Modules หลัก

ระบบปัจจุบันแบ่งออกเป็น **7 โมดูลหลัก**:

1. **User & Identity Management** - จัดการข้อมูลผู้ใช้และโปรไฟล์
2. **Social Media Features** - โพสต์, คอมเมนต์, โหวต, แท็ก, ฟอลโล่
3. **Chat & Messaging** - แชทแบบ 1-on-1
4. **Notifications** - การแจ้งเตือนและ Web Push
5. **Media & Files** - จัดการรูปภาพ, วิดีโอ, ไฟล์
6. **Auto-Posting (AI)** - สร้างโพสต์อัตโนมัติด้วย AI
7. **Legacy Features** - Task, File, Job Management (เก่า)

---

## 📋 รายละเอียด Features แต่ละโมดูล

### 1. User & Identity Management

**📌 วัตถุประสงค์**: จัดการข้อมูลผู้ใช้ โปรไฟล์ และ Authentication

**🔧 Services**:
- `UsersCacheService` - ซิงค์ข้อมูลผู้ใช้จาก Auth Service
- `UserProfileService` - จัดการโปรไฟล์ส่วนตัวของผู้ใช้

**📊 Database Tables**:
| Table | Purpose | Key Fields |
|-------|---------|------------|
| `users_cache` | เก็บข้อมูลผู้ใช้ที่ซิงค์จาก Auth Service | ID, Email, Username, DisplayName, Avatar, Role, IsActive, SyncedAt |
| `user_profiles` | เก็บข้อมูลโปรไฟล์ Social (แยกจาก Auth) | UserID, Bio, Location, Website, DisplayName, Avatar, Karma, FollowersCount, FollowingCount |

**🌐 API Endpoints**:
- `GET /api/v1/users/{userId}/profile` - ดูโปรไฟล์
- `PUT /api/v1/users/{userId}/profile` - แก้ไขโปรไฟล์
- `GET /api/v1/users/{userId}/stats` - สถิติผู้ใช้

**🔗 Dependencies**:
- เชื่อมต่อกับ **Auth Service** (ภายนอก) ผ่าน Webhook
- ใช้ `users_cache` เป็น Read Model

---

### 2. Social Media Features

**📌 วัตถุประสงค์**: ฟีเจอร์หลักของ Social Network

**🔧 Services**:
- `PostService` - CRUD โพสต์, Feed Generation
- `CommentService` - คอมเมนต์แบบซ้อนได้ (max depth 10)
- `VoteService` - Upvote/Downvote โพสต์และคอมเมนต์
- `FollowService` - ระบบติดตามผู้ใช้
- `SavedPostService` - บุ๊คมาร์คโพสต์
- `TagService` - จัดการแท็ก
- `SearchService` - ค้นหาโพสต์และผู้ใช้

**📊 Database Tables**:
| Table | Purpose | Relationships |
|-------|---------|---------------|
| `posts` | โพสต์หลัก | → users_cache (author), → comments, → votes, → media (M2M), → tags (M2M), → saved_posts |
| `comments` | คอมเมนต์แบบซ้อน | → posts, → users_cache (author), → votes, Self-referencing (parent) |
| `votes` | โหวตโพสต์/คอมเมนต์ | → posts, → comments, → users_cache |
| `tags` | แท็กหมวดหมู่ | → posts (M2M) |
| `post_tags` | Junction table | posts ↔ tags |
| `post_media` | Junction table | posts ↔ media |
| `follows` | ความสัมพันธ์การติดตาม | users_cache ↔ users_cache |
| `saved_posts` | บุ๊คมาร์ค | → posts, → users_cache |
| `search_history` | ประวัติการค้นหา | → users_cache |

**🌐 API Endpoints**:
- `POST /api/v1/posts` - สร้างโพสต์
- `GET /api/v1/posts` - ดูรายการโพสต์
- `GET /api/v1/posts/{id}` - ดูโพสต์
- `PUT /api/v1/posts/{id}` - แก้ไขโพสต์
- `DELETE /api/v1/posts/{id}` - ลบโพสต์
- `GET /api/v1/feed` - ดูฟีดส่วนตัว
- `POST /api/v1/posts/{postId}/comments` - สร้างคอมเมนต์
- `GET /api/v1/posts/{postId}/comments` - ดูคอมเมนต์
- `POST /api/v1/votes` - โหวต
- `DELETE /api/v1/votes/{id}` - ยกเลิกโหวต
- `POST /api/v1/users/{userId}/follow` - ติดตาม
- `DELETE /api/v1/users/{userId}/follow` - เลิกติดตาม
- `GET /api/v1/users/{userId}/followers` - รายชื่อผู้ติดตาม
- `GET /api/v1/users/{userId}/following` - รายชื่อกำลังติดตาม
- `POST /api/v1/saved-posts` - บันทึกโพสต์
- `GET /api/v1/saved-posts` - ดูโพสต์ที่บันทึก
- `GET /api/v1/tags` - ดูแท็กทั้งหมด
- `GET /api/v1/search/posts` - ค้นหาโพสต์
- `GET /api/v1/search/users` - ค้นหาผู้ใช้

**🔗 Dependencies**:
- `PostService` → TagService, MediaService, VoteService, NotificationHub, Redis
- `CommentService` → PostService, NotificationService
- `VoteService` → PostService, CommentService, NotificationService
- `FollowService` → NotificationService

---

### 3. Chat & Messaging

**📌 วัตถุประสงค์**: แชทแบบ Real-time 1-on-1

**🔧 Services**:
- `ConversationService` - จัดการห้องแชท
- `MessageService` - ส่ง/รับข้อความ
- `BlockService` - บล็อกผู้ใช้
- `ChatHub` - WebSocket Hub สำหรับ Real-time

**📊 Database Tables**:
| Table | Purpose | Special Features |
|-------|---------|------------------|
| `conversations` | ห้องแชท 1-on-1 | Denormalized: LastMessageID, UnreadCount (per user) |
| `messages` | ข้อความ | Support: Text, Image, Video, File, JSONB Media field |
| `blocks` | บล็อกผู้ใช้ | Prevent blocked users from messaging |

**🌐 API Endpoints**:
- `POST /api/v1/conversations` - สร้างห้องแชท
- `GET /api/v1/conversations` - ดูรายการห้องแชท
- `GET /api/v1/conversations/{id}` - ดูห้องแชท
- `POST /api/v1/conversations/{id}/messages` - ส่งข้อความ
- `GET /api/v1/conversations/{id}/messages` - ดูข้อความ
- `PUT /api/v1/messages/{id}` - แก้ไขข้อความ
- `DELETE /api/v1/messages/{id}` - ลบข้อความ
- `POST /api/v1/blocks` - บล็อกผู้ใช้
- `GET /api/v1/blocks` - ดูรายการบล็อก
- `DELETE /api/v1/blocks/{id}` - ยกเลิกบล็อก
- **WebSocket**: `/ws/chat` - Real-time chat

**🔗 Dependencies**:
- `MessageService` → ConversationService, BlockService, ChatHub, PushService
- `ConversationService` → BlockService

---

### 4. Notifications

**📌 วัตถุประสงค์**: แจ้งเตือนกิจกรรมและ Web Push

**🔧 Services**:
- `NotificationService` - จัดการการแจ้งเตือน
- `PushService` - Web Push Notification (VAPID)
- `NotificationHub` - WebSocket Hub สำหรับ Real-time

**📊 Database Tables**:
| Table | Purpose |
|-------|---------|
| `notifications` | การแจ้งเตือน (reply, vote, mention, follow) |
| `notification_settings` | ตั้งค่าการแจ้งเตือนของผู้ใช้ |
| `push_subscriptions` | Web Push endpoints (PWA) |

**🌐 API Endpoints**:
- `GET /api/v1/notifications` - ดูการแจ้งเตือน
- `PUT /api/v1/notifications/mark-read` - ทำเครื่องหมายอ่านแล้ว
- `GET /api/v1/notifications/settings` - ดูการตั้งค่า
- `PUT /api/v1/notifications/settings` - แก้ไขการตั้งค่า
- `POST /api/v1/push/subscribe` - Subscribe Web Push
- `DELETE /api/v1/push/unsubscribe` - Unsubscribe
- **WebSocket**: `/ws/notifications` - Real-time notifications

**🔗 Dependencies**:
- ใช้โดย: PostService, CommentService, VoteService, FollowService, MessageService

---

### 5. Media & Files

**📌 วัตถุประสงค์**: จัดการอัปโหลดและเก็บไฟล์

**🔧 Services**:
- `MediaService` - CRUD Media
- `FileUploadService` - อัปโหลดไฟล์
- `MediaUploadService` - อัปโหลดแบบ Multi-storage
- `VideoEncoderWorker` - ติดตามสถานะการเข้ารหัสวิดีโอ

**📊 Database Tables**:
| Table | Purpose | Storage Backend |
|-------|---------|-----------------|
| `media` | Metadata ของไฟล์ทั้งหมด | Bunny CDN (images), Bunny Stream (video), Cloudflare R2 |
| `files` (Legacy) | Metadata ไฟล์เก่า | Bunny CDN |

**🌐 API Endpoints**:
- `POST /api/v1/media` - อัปโหลดรูป/วิดีโอ
- `GET /api/v1/media/{id}` - ดูข้อมูล Media
- `DELETE /api/v1/media/{id}` - ลบ Media
- `POST /api/v1/upload` - อัปโหลดไฟล์ทั่วไป
- `GET /api/v1/upload/presigned` - ขอ Presigned URL (R2)
- **Webhook**: `/webhook/bunny/video-completed` - Callback เมื่อเข้ารหัสวิดีโอเสร็จ

**🔗 Dependencies**:
- External: Bunny Storage, Bunny Stream, Cloudflare R2
- Internal: Redis (queue), NotificationService

---

### 6. Auto-Posting (AI)

**📌 วัตถุประสงค์**: สร้างโพสต์อัตโนมัติด้วย OpenAI

**🔧 Services**:
- `AutoPostService` - สร้างโพสต์แบบ Cron-based
- `SimpleAutoPostService` - สร้างโพสต์แบบ Queue-based
- `EventScheduler` - ตัว Scheduler หลัก

**📊 Database Tables**:
| Table | Purpose |
|-------|---------|
| `auto_post_settings` | การตั้งค่าการสร้างโพสต์อัตโนมัติ |
| `auto_post_logs` | ประวัติการสร้างโพสต์ |

**🌐 API Endpoints**:
- `GET /api/v1/auto-post/settings` - ดูการตั้งค่า
- `PUT /api/v1/auto-post/settings` - แก้ไขการตั้งค่า
- `POST /api/v1/auto-post/generate` - สร้างโพสต์ด้วยตนเอง
- `GET /api/v1/auto-post/logs` - ดูประวัติ
- `POST /api/v1/simple-auto-post/upload-topics` - อัปโหลดหัวข้อ
- `POST /api/v1/simple-auto-post/process-next` - สร้างโพสต์ถัดไป

**🔗 Dependencies**:
- External: OpenAI API (gpt-4o-mini)
- Internal: PostService, EventScheduler

**⏰ Scheduling**:
- Cron: ทุก 1 ชั่วโมง (`0 * * * *`)

---

### 7. Legacy Features

**📌 วัตถุประสงค์**: ฟีเจอร์เก่าที่อาจถูกยกเลิกในอนาคต

**🔧 Services**:
- `TaskService` - จัดการ Tasks
- `FileService` - จัดการไฟล์ (เก่า)
- `JobService` - Background Jobs

**📊 Database Tables**:
| Table | Purpose |
|-------|---------|
| `tasks` | Task management |
| `files` | File storage (legacy) |
| `jobs` | Background job scheduler |

**🌐 API Endpoints**:
- `/api/v1/tasks/*` - Task CRUD
- `/api/v1/files/*` - File CRUD (Legacy)
- `/api/v1/jobs/*` - Job CRUD

---

## 📊 สรุปจำนวน Components

| Category | Count |
|----------|-------|
| **Services** | 21 services |
| **Database Tables** | 27 tables |
| **HTTP Handlers** | ~80 endpoints |
| **WebSocket Hubs** | 2 (Chat, Notification) |
| **Background Workers** | 2 (VideoEncoder, EventScheduler) |
| **External Integrations** | 4 (Bunny CDN, Bunny Stream, Cloudflare R2, OpenAI) |

---

## 🔄 ความสัมพันธ์ระหว่าง Modules

```
┌─────────────────────┐
│   User & Profile    │◄────────────┐
└─────────────────────┘             │
         ▲                          │
         │ uses                     │
         │                          │
┌─────────────────────┐    ┌────────────────┐
│  Social Features    │───►│ Notifications  │
│ (Post/Comment/Vote) │    └────────────────┘
└─────────────────────┘             ▲
         │                          │
         │ generates                │
         │                          │
┌─────────────────────┐             │
│   Auto-Posting AI   │─────────────┘
└─────────────────────┘

┌─────────────────────┐    ┌────────────────┐
│   Chat & Message    │───►│ Notifications  │
└─────────────────────┘    └────────────────┘
         │
         │ uses
         ▼
┌─────────────────────┐
│   Media & Files     │
└─────────────────────┘
```

---

## 🎯 Key Insights

1. **Social Features** เป็น Core Module ที่ใหญ่ที่สุด (7 services, 10 tables)
2. **User Data** แยกเป็น 2 ส่วน: Authentication (Auth Service) และ Social Profile (UserProfile)
3. **Real-time** ใช้ WebSocket Hubs (2 hubs) + WebSocket connections
4. **AI Integration** ใช้ OpenAI สร้างโพสต์อัตโนมัติ
5. **Media Processing** ใช้ Bunny Stream + Background Worker แบบ async
6. **Legacy Features** มี 3 services ที่อาจถูกเลิกใช้
