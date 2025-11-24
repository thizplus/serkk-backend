# 2️⃣ Database / Tables

สรุป Database Schema ทั้งหมดใน Monolith ปัจจุบัน

---

## ภาพรวม Database

- **Database Type**: PostgreSQL
- **ORM**: GORM
- **Total Tables**: 27 tables
- **Migration Files**: 25+ migration files

---

## 📋 รายละเอียด Tables ทั้งหมด

### 1. User & Identity Management (2 tables)

#### `users_cache`
**Purpose**: เก็บข้อมูลผู้ใช้ที่ซิงค์จาก Auth Service (Read Model)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | User ID (จาก Auth Service) |
| `email` | VARCHAR | UNIQUE, NOT NULL | อีเมล |
| `username` | VARCHAR | UNIQUE, NOT NULL | ชื่อผู้ใช้ |
| `display_name` | VARCHAR | | ชื่อแสดง |
| `avatar` | VARCHAR | | URL รูปโปรไฟล์ |
| `role` | VARCHAR | DEFAULT 'user' | บทบาท (user, admin) |
| `is_active` | BOOLEAN | DEFAULT true | สถานะการใช้งาน |
| `synced_at` | TIMESTAMP | | เวลาซิงค์ข้อมูลล่าสุด |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_users_cache_username` (username)
- `idx_users_cache_email` (email)

**Relationships**:
- **Used by**: posts, comments, votes, follows, messages, notifications, user_profiles

**Notes**:
- ซิงค์ผ่าน Internal Webhook จาก Auth Service
- เป็น **Read Model** เท่านั้น (ไม่แก้ไขโดยตรง)

---

#### `user_profiles`
**Purpose**: เก็บข้อมูลโปรไฟล์ Social-specific (แยกจาก Auth)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Profile ID |
| `user_id` | UUID | FK (users_cache), UNIQUE | User ID |
| `bio` | TEXT | | แนะนำตัว |
| `location` | VARCHAR | | ที่อยู่ |
| `website` | VARCHAR | | เว็บไซต์ |
| `display_name` | VARCHAR | | ชื่อแสดง (ซิงค์กับ users_cache) |
| `avatar` | VARCHAR | | URL รูปโปรไฟล์ (ซิงค์กับ users_cache) |
| `karma` | INTEGER | DEFAULT 0 | คะแนนชื่อเสียง |
| `followers_count` | INTEGER | DEFAULT 0 | จำนวนผู้ติดตาม (denormalized) |
| `following_count` | INTEGER | DEFAULT 0 | จำนวนกำลังติดตาม (denormalized) |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_user_profiles_user_id` (user_id)

**Relationships**:
- `user_id` → `users_cache.id` (FK)

**Notes**:
- เก็บข้อมูลเฉพาะ Social features
- Denormalized counts เพื่อ performance

---

### 2. Social Media Features (10 tables)

#### `posts`
**Purpose**: โพสต์หลัก

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Post ID |
| `title` | VARCHAR | NOT NULL | หัวข้อ |
| `content` | TEXT | | เนื้อหา |
| `author_id` | UUID | FK (users_cache) | ผู้เขียน |
| `votes` | INTEGER | DEFAULT 0 | คะแนนโหวตรวม (denormalized) |
| `comment_count` | INTEGER | DEFAULT 0 | จำนวนคอมเมนต์ (denormalized) |
| `type` | VARCHAR | DEFAULT 'text' | ประเภท (text, image, video, link, repost) |
| `source_post_id` | UUID | FK (posts), NULLABLE | โพสต์ต้นฉบับ (สำหรับ repost) |
| `status` | VARCHAR | DEFAULT 'published' | สถานะ (published, deleted, pending) |
| `client_post_id` | VARCHAR | UNIQUE, NULLABLE | Idempotency key |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |
| `deleted_at` | TIMESTAMP | NULLABLE | Soft delete |

**Indexes**:
- `idx_posts_author_id` (author_id)
- `idx_posts_created_at` (created_at DESC)
- `idx_posts_client_post_id` (client_post_id)
- `idx_posts_type` (type)

**Relationships**:
- `author_id` → `users_cache.id` (FK)
- `source_post_id` → `posts.id` (FK, self-referencing)
- Has many: `comments`, `votes`, `saved_posts`
- Many-to-many: `tags` (via post_tags), `media` (via post_media)

---

#### `comments`
**Purpose**: คอมเมนต์แบบซ้อนได้ (max depth 10)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Comment ID |
| `post_id` | UUID | FK (posts) | โพสต์ที่คอมเมนต์ |
| `author_id` | UUID | FK (users_cache) | ผู้เขียน |
| `content` | TEXT | NOT NULL | เนื้อหา |
| `votes` | INTEGER | DEFAULT 0 | คะแนนโหวตรวม (denormalized) |
| `parent_id` | UUID | FK (comments), NULLABLE | Comment ที่ตอบกลับ |
| `depth` | INTEGER | DEFAULT 0 | ระดับความซ้อน (0-10) |
| `is_deleted` | BOOLEAN | DEFAULT false | ถูกลบหรือไม่ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_comments_post_id` (post_id)
- `idx_comments_author_id` (author_id)
- `idx_comments_parent_id` (parent_id)

**Relationships**:
- `post_id` → `posts.id` (FK)
- `author_id` → `users_cache.id` (FK)
- `parent_id` → `comments.id` (FK, self-referencing)
- Has many: `votes`

---

#### `votes`
**Purpose**: โหวตโพสต์และคอมเมนต์

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Vote ID |
| `user_id` | UUID | FK (users_cache) | ผู้โหวต |
| `target_id` | UUID | NOT NULL | ID ของโพสต์/คอมเมนต์ |
| `target_type` | VARCHAR | NOT NULL | ประเภท (post, comment) |
| `vote_type` | VARCHAR | NOT NULL | ประเภทโหวต (upvote, downvote) |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_votes_user_target` (user_id, target_id, target_type) - UNIQUE
- `idx_votes_target` (target_id, target_type)

**Relationships**:
- `user_id` → `users_cache.id` (FK)
- Polymorphic: target_id → posts.id OR comments.id

---

#### `tags`
**Purpose**: แท็กหมวดหมู่

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Tag ID |
| `name` | VARCHAR | UNIQUE, NOT NULL | ชื่อแท็ก |
| `post_count` | INTEGER | DEFAULT 0 | จำนวนโพสต์ (denormalized) |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_tags_name` (name)

**Relationships**:
- Many-to-many: `posts` (via post_tags)

---

#### `post_tags`
**Purpose**: Junction table สำหรับ Posts ↔ Tags

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `post_id` | UUID | FK (posts) | |
| `tag_id` | UUID | FK (tags) | |
| `created_at` | TIMESTAMP | | |

**Indexes**:
- `idx_post_tags_post_tag` (post_id, tag_id) - UNIQUE (Composite PK)

---

#### `post_media`
**Purpose**: Junction table สำหรับ Posts ↔ Media

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `post_id` | UUID | FK (posts) | |
| `media_id` | UUID | FK (media) | |
| `display_order` | INTEGER | DEFAULT 0 | ลำดับการแสดงผล |
| `created_at` | TIMESTAMP | | |

**Indexes**:
- `idx_post_media_post_media` (post_id, media_id) - UNIQUE

---

#### `follows`
**Purpose**: ความสัมพันธ์การติดตามผู้ใช้

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Follow ID |
| `follower_id` | UUID | FK (users_cache) | ผู้ติดตาม |
| `following_id` | UUID | FK (users_cache) | ผู้ถูกติดตาม |
| `created_at` | TIMESTAMP | | |

**Indexes**:
- `idx_follows_follower_following` (follower_id, following_id) - UNIQUE

**Relationships**:
- `follower_id` → `users_cache.id` (FK)
- `following_id` → `users_cache.id` (FK)

---

#### `saved_posts`
**Purpose**: บุ๊คมาร์คโพสต์

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Saved Post ID |
| `user_id` | UUID | FK (users_cache) | ผู้บันทึก |
| `post_id` | UUID | FK (posts) | โพสต์ที่บันทึก |
| `created_at` | TIMESTAMP | | |

**Indexes**:
- `idx_saved_posts_user_post` (user_id, post_id) - UNIQUE

**Relationships**:
- `user_id` → `users_cache.id` (FK)
- `post_id` → `posts.id` (FK)

---

#### `search_history`
**Purpose**: ประวัติการค้นหา

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Search ID |
| `user_id` | UUID | FK (users_cache) | ผู้ค้นหา |
| `query` | VARCHAR | NOT NULL | คำค้นหา |
| `type` | VARCHAR | | ประเภท (posts, users, all) |
| `searched_at` | TIMESTAMP | | วันที่ค้นหา |

**Indexes**:
- `idx_search_history_user_id` (user_id)

**Relationships**:
- `user_id` → `users_cache.id` (FK)

---

### 3. Chat & Messaging (3 tables)

#### `conversations`
**Purpose**: ห้องแชท 1-on-1

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Conversation ID |
| `user1_id` | UUID | FK (users_cache) | ผู้ใช้คนที่ 1 (เรียงตาม UUID) |
| `user2_id` | UUID | FK (users_cache) | ผู้ใช้คนที่ 2 |
| `last_message_id` | UUID | FK (messages), NULLABLE | ข้อความล่าสุด (denormalized) |
| `last_message_at` | TIMESTAMP | NULLABLE | เวลาข้อความล่าสุด (denormalized) |
| `user1_unread_count` | INTEGER | DEFAULT 0 | ข้อความที่ยังไม่อ่าน (user1) |
| `user2_unread_count` | INTEGER | DEFAULT 0 | ข้อความที่ยังไม่อ่าน (user2) |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_conversations_user1_user2` (user1_id, user2_id) - UNIQUE
- `idx_conversations_last_message_at` (last_message_at DESC)

**Relationships**:
- `user1_id` → `users_cache.id` (FK)
- `user2_id` → `users_cache.id` (FK)
- `last_message_id` → `messages.id` (FK)
- Has many: `messages`

**Notes**:
- User ordering: UUID ของ user1 < user2 เสมอ (เพื่อป้องกัน duplicate)
- Denormalized fields เพื่อ performance

---

#### `messages`
**Purpose**: ข้อความในแชท

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Message ID |
| `conversation_id` | UUID | FK (conversations) | ห้องแชท |
| `sender_id` | UUID | FK (users_cache) | ผู้ส่ง |
| `receiver_id` | UUID | FK (users_cache) | ผู้รับ |
| `type` | VARCHAR | DEFAULT 'text' | ประเภท (text, image, video, file) |
| `content` | TEXT | | เนื้อหาข้อความ |
| `media` | JSONB | NULLABLE | ข้อมูล Media (URL, Thumbnail, etc.) |
| `is_read` | BOOLEAN | DEFAULT false | อ่านแล้วหรือไม่ |
| `read_at` | TIMESTAMP | NULLABLE | เวลาที่อ่าน |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |
| `deleted_at` | TIMESTAMP | NULLABLE | Soft delete |

**Indexes**:
- `idx_messages_conversation_id` (conversation_id)
- `idx_messages_sender_id` (sender_id)
- `idx_messages_created_at` (created_at DESC)

**Relationships**:
- `conversation_id` → `conversations.id` (FK)
- `sender_id` → `users_cache.id` (FK)
- `receiver_id` → `users_cache.id` (FK)

**JSONB Media Structure**:
```json
{
  "url": "https://cdn.example.com/video.mp4",
  "thumbnail": "https://cdn.example.com/thumb.jpg",
  "type": "video",
  "filename": "video.mp4",
  "mime_type": "video/mp4",
  "size": 1048576,
  "width": 1920,
  "height": 1080,
  "duration": 120,
  "video_id": "bunny-stream-id",
  "hls_url": "https://stream.example.com/playlist.m3u8",
  "encoding_status": "completed",
  "encoding_progress": 100
}
```

---

#### `blocks`
**Purpose**: บล็อกผู้ใช้

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Block ID |
| `blocker_id` | UUID | FK (users_cache) | ผู้บล็อก |
| `blocked_id` | UUID | FK (users_cache) | ผู้ถูกบล็อก |
| `created_at` | TIMESTAMP | | |

**Indexes**:
- `idx_blocks_blocker_blocked` (blocker_id, blocked_id) - UNIQUE

**Relationships**:
- `blocker_id` → `users_cache.id` (FK)
- `blocked_id` → `users_cache.id` (FK)

---

### 4. Notifications (3 tables)

#### `notifications`
**Purpose**: การแจ้งเตือนกิจกรรม

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Notification ID |
| `user_id` | UUID | FK (users_cache) | ผู้รับการแจ้งเตือน |
| `sender_id` | UUID | FK (users_cache), NULLABLE | ผู้ทำกิจกรรม |
| `type` | VARCHAR | NOT NULL | ประเภท (reply, vote, mention, follow) |
| `message` | TEXT | | ข้อความ |
| `post_id` | UUID | FK (posts), NULLABLE | โพสต์ที่เกี่ยวข้อง |
| `comment_id` | UUID | FK (comments), NULLABLE | คอมเมนต์ที่เกี่ยวข้อง |
| `is_read` | BOOLEAN | DEFAULT false | อ่านแล้วหรือไม่ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_notifications_user_id` (user_id)
- `idx_notifications_is_read` (is_read)

**Relationships**:
- `user_id` → `users_cache.id` (FK)
- `sender_id` → `users_cache.id` (FK)
- `post_id` → `posts.id` (FK)
- `comment_id` → `comments.id` (FK)

---

#### `notification_settings`
**Purpose**: ตั้งค่าการแจ้งเตือนของผู้ใช้

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Setting ID |
| `user_id` | UUID | FK (users_cache), UNIQUE | ผู้ใช้ |
| `replies` | BOOLEAN | DEFAULT true | แจ้งเตือนการตอบกลับ |
| `mentions` | BOOLEAN | DEFAULT true | แจ้งเตือนการกล่าวถึง |
| `votes` | BOOLEAN | DEFAULT true | แจ้งเตือนการโหวต |
| `follows` | BOOLEAN | DEFAULT true | แจ้งเตือนการติดตาม |
| `email_notifications` | BOOLEAN | DEFAULT false | ส่งอีเมล |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_notification_settings_user_id` (user_id)

**Relationships**:
- `user_id` → `users_cache.id` (FK)

---

#### `push_subscriptions`
**Purpose**: Web Push Notification endpoints (PWA)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Subscription ID |
| `user_id` | UUID | FK (users_cache) | ผู้ใช้ |
| `endpoint` | TEXT | NOT NULL | Push endpoint URL |
| `p256dh` | TEXT | NOT NULL | Public key |
| `auth` | TEXT | NOT NULL | Auth secret |
| `expiration_time` | TIMESTAMP | NULLABLE | เวลาหมดอายุ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_push_subscriptions_user_id` (user_id)
- `idx_push_subscriptions_endpoint` (endpoint) - UNIQUE

**Relationships**:
- `user_id` → `users_cache.id` (FK)

---

### 5. Media & Files (2 tables)

#### `media`
**Purpose**: Metadata ของไฟล์ทั้งหมด (รูป, วิดีโอ, ไฟล์)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Media ID |
| `user_id` | UUID | FK (users_cache) | ผู้อัปโหลด |
| `type` | VARCHAR | NOT NULL | ประเภท (image, video, file) |
| `file_name` | VARCHAR | | ชื่อไฟล์ |
| `extension` | VARCHAR | | นามสกุล |
| `mime_type` | VARCHAR | | MIME type |
| `size` | BIGINT | | ขนาดไฟล์ (bytes) |
| `url` | VARCHAR | NOT NULL | URL ของไฟล์ |
| `thumbnail` | VARCHAR | | URL รูป Thumbnail |
| `width` | INTEGER | NULLABLE | ความกว้าง (รูป/วิดีโอ) |
| `height` | INTEGER | NULLABLE | ความสูง |
| `duration` | INTEGER | NULLABLE | ความยาว (วิดีโอ, วินาที) |
| `source_type` | VARCHAR | NULLABLE | แหล่งที่ใช้ (post, message, profile) |
| `source_id` | UUID | NULLABLE | ID ของแหล่งที่ใช้ |
| `usage_count` | INTEGER | DEFAULT 0 | จำนวนครั้งที่ใช้ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Indexes**:
- `idx_media_user_id` (user_id)
- `idx_media_type` (type)
- `idx_media_source` (source_type, source_id)

**Relationships**:
- `user_id` → `users_cache.id` (FK)
- Polymorphic: source_id → posts.id OR messages.id OR user_profiles.id
- Many-to-many: `posts` (via post_media)

**Storage Backends**:
- Images: Bunny CDN
- Videos: Bunny Stream
- Files: Cloudflare R2 (optional)

---

#### `files` (Legacy)
**Purpose**: Metadata ไฟล์เก่า (ก่อนมี media table)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | File ID |
| `file_name` | VARCHAR | | ชื่อไฟล์ |
| `file_size` | BIGINT | | ขนาดไฟล์ |
| `mime_type` | VARCHAR | | MIME type |
| `url` | VARCHAR | | URL |
| `cdn_path` | VARCHAR | | Path ใน CDN |
| `user_id` | UUID | FK (users_cache) | ผู้อัปโหลด |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Relationships**:
- `user_id` → `users_cache.id` (FK)

---

### 6. Auto-Posting (2 tables)

#### `auto_post_settings`
**Purpose**: การตั้งค่าการสร้างโพสต์อัตโนมัติด้วย AI

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Setting ID |
| `bot_user_id` | UUID | FK (users_cache) | บอทที่โพสต์ |
| `is_enabled` | BOOLEAN | DEFAULT false | เปิดใช้งาน |
| `cron_schedule` | VARCHAR | DEFAULT '0 * * * *' | Cron expression |
| `model` | VARCHAR | DEFAULT 'gpt-4o-mini' | OpenAI model |
| `topics` | JSONB | | หัวข้อที่จะสร้าง |
| `max_tokens` | INTEGER | DEFAULT 500 | จำนวน tokens สูงสุด |
| `temperature` | FLOAT | DEFAULT 0.7 | ความสร้างสรรค์ |
| `tone` | VARCHAR | DEFAULT 'neutral' | โทนเสียง |
| `enable_variations` | BOOLEAN | DEFAULT false | สร้างหัวข้อแบบหลากหลาย |
| `variation_style` | VARCHAR | | รูปแบบการสร้าง |
| `require_approval` | BOOLEAN | DEFAULT false | ต้องอนุมัติก่อนโพสต์ |
| `sensitive_topics` | JSONB | | หัวข้อที่หลีกเลี่ยง |
| `batch_size` | INTEGER | DEFAULT 1 | จำนวนโพสต์ต่อครั้ง |
| `use_batch_mode` | BOOLEAN | DEFAULT false | ใช้ Batch API |
| `total_posts_generated` | INTEGER | DEFAULT 0 | จำนวนโพสต์ทั้งหมด |
| `last_generated_at` | TIMESTAMP | NULLABLE | เวลาสร้างล่าสุด |
| `last_error` | TEXT | NULLABLE | Error ล่าสุด |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Relationships**:
- `bot_user_id` → `users_cache.id` (FK)

---

#### `auto_post_logs`
**Purpose**: ประวัติการสร้างโพสต์อัตโนมัติ

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Log ID |
| `setting_id` | UUID | FK (auto_post_settings) | การตั้งค่า |
| `post_id` | UUID | FK (posts), NULLABLE | โพสต์ที่สร้าง |
| `topic` | VARCHAR | | หัวข้อที่ใช้ |
| `generated_title` | VARCHAR | | หัวข้อที่ AI สร้าง |
| `status` | VARCHAR | | สถานะ (pending, success, failed, approved, rejected) |
| `error_message` | TEXT | NULLABLE | ข้อความ Error |
| `tokens_used` | INTEGER | | จำนวน tokens ที่ใช้ |
| `prompt_tokens` | INTEGER | | Prompt tokens |
| `completion_tokens` | INTEGER | | Completion tokens |
| `metadata` | JSONB | | ข้อมูลเพิ่มเติม |
| `title_variation_used` | VARCHAR | NULLABLE | รูปแบบที่ใช้ |
| `approved_by` | UUID | FK (users_cache), NULLABLE | ผู้อนุมัติ |
| `approved_at` | TIMESTAMP | NULLABLE | เวลาอนุมัติ |
| `rejected_by` | UUID | FK (users_cache), NULLABLE | ผู้ปฏิเสธ |
| `rejected_at` | TIMESTAMP | NULLABLE | เวลาปฏิเสธ |
| `rejection_reason` | TEXT | NULLABLE | เหตุผลปฏิเสธ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Relationships**:
- `setting_id` → `auto_post_settings.id` (FK)
- `post_id` → `posts.id` (FK)
- `approved_by` → `users_cache.id` (FK)
- `rejected_by` → `users_cache.id` (FK)

---

### 7. Legacy Features (2 tables)

#### `tasks`
**Purpose**: Task management (อาจถูกเลิกใช้)

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Task ID |
| `title` | VARCHAR | NOT NULL | หัวข้อ |
| `description` | TEXT | | รายละเอียด |
| `status` | VARCHAR | DEFAULT 'pending' | สถานะ (pending, completed) |
| `priority` | INTEGER | DEFAULT 0 | ความสำคัญ |
| `due_date` | TIMESTAMP | NULLABLE | วันครบกำหนด |
| `user_id` | UUID | FK (users_cache) | ผู้รับผิดชอบ |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

**Relationships**:
- `user_id` → `users_cache.id` (FK)

---

#### `jobs`
**Purpose**: Background job scheduler

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `id` | UUID | PK | Job ID |
| `name` | VARCHAR | NOT NULL | ชื่อ Job |
| `cron_expr` | VARCHAR | NOT NULL | Cron expression |
| `payload` | JSONB | | ข้อมูล Payload |
| `status` | VARCHAR | DEFAULT 'pending' | สถานะ |
| `last_run` | TIMESTAMP | NULLABLE | เวลาที่รันล่าสุด |
| `next_run` | TIMESTAMP | NULLABLE | เวลาที่จะรันถัดไป |
| `is_active` | BOOLEAN | DEFAULT true | เปิดใช้งาน |
| `created_at` | TIMESTAMP | | |
| `updated_at` | TIMESTAMP | | |

---

## 🔄 Entity Relationship Diagram (Summary)

```
users_cache
    ├─► user_profiles (1:1)
    ├─► posts (1:N)
    ├─► comments (1:N)
    ├─► votes (1:N)
    ├─► follows (M:N self)
    ├─► saved_posts (M:N via saved_posts)
    ├─► conversations (M:N via user1/user2)
    ├─► messages (1:N sender/receiver)
    ├─► blocks (M:N self)
    ├─► notifications (1:N user/sender)
    ├─► media (1:N)
    └─► auto_post_settings (1:N)

posts
    ├─► comments (1:N)
    ├─► votes (1:N polymorphic)
    ├─► tags (M:N via post_tags)
    ├─► media (M:N via post_media)
    ├─► saved_posts (M:N via saved_posts)
    └─► posts (1:N self-referencing for reposts)

comments
    ├─► votes (1:N polymorphic)
    └─► comments (1:N self-referencing for replies)

conversations
    └─► messages (1:N)

auto_post_settings
    └─► auto_post_logs (1:N)
```

---

## 📊 สรุป Table Categories

| Category | Tables | Total |
|----------|--------|-------|
| **User & Identity** | users_cache, user_profiles | 2 |
| **Social Media** | posts, comments, votes, tags, post_tags, post_media, follows, saved_posts, search_history | 10 |
| **Chat & Messaging** | conversations, messages, blocks | 3 |
| **Notifications** | notifications, notification_settings, push_subscriptions | 3 |
| **Media & Files** | media, files | 2 |
| **Auto-Posting** | auto_post_settings, auto_post_logs | 2 |
| **Legacy** | tasks, jobs | 2 |
| **Cache** | (Redis - not DB) | - |
| **TOTAL** | | **27** |

---

## 🎯 Key Insights

1. **Denormalization**: หลาย table ใช้ denormalized counts (votes, comment_count, followers_count) เพื่อ performance
2. **Polymorphic Relationships**: `votes` และ `media` ใช้ polymorphic relationships
3. **Soft Deletes**: `posts` และ `messages` ใช้ soft delete (deleted_at)
4. **JSONB Usage**: ใช้ JSONB ใน `messages.media`, `auto_post_settings.topics`, `jobs.payload`
5. **Idempotency**: `posts.client_post_id` สำหรับป้องกันโพสต์ซ้ำ
6. **Foreign Keys**: มี FK relationships ครบถ้วนระหว่าง tables
