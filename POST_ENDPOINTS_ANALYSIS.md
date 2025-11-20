# การวิเคราะห์ POST Endpoints ทั้งหมดในระบบ

วันที่วิเคราะห์: 2025-11-14
โปรเจกต์: Go Fiber Backend (Social Media Platform)

---

## สารบัญ

1. [ภาพรวมของระบบ](#ภาพรวมของระบบ)
2. [Authentication & OAuth](#1-authentication--oauth)
3. [Posts](#2-posts)
4. [Comments](#3-comments)
5. [Votes](#4-votes)
6. [Follows](#5-follows)
7. [Saved Posts](#6-saved-posts)
8. [Media Upload](#7-media-upload)
9. [Push Notifications](#8-push-notifications)
10. [Chat & Messaging](#9-chat--messaging)
11. [File Upload](#10-file-upload)
12. [Webhooks](#11-webhooks)
13. [Legacy Endpoints](#12-legacy-endpoints)
14. [Admin Jobs](#13-admin-jobs)
15. [สรุปและข้อสังเกต](#สรุปและข้อสังเกต)

---

## ภาพรวมของระบบ

### สถิติ POST Endpoints
- **จำนวน POST Endpoints ทั้งหมด**: 32 endpoints
- **Endpoints ที่ต้อง Authentication**: 30 endpoints
- **Public Endpoints**: 2 endpoints (webhook และ OAuth callback)
- **จำนวนโมดูล**: 15 modules
- **Handler Files**: 14 files

### โครงสร้าง URL Base Path
```
/api/v1/{module}/{action}
```

### Authentication Method
- **JWT Bearer Token** ผ่าน middleware `middleware.Protected()`
- Token ถูกเก็บใน Header: `Authorization: Bearer <token>`
- UserID ถูกดึงจาก `c.Locals("userID").(uuid.UUID)`

---

## 1. Authentication & OAuth

### 1.1 POST /api/v1/auth/register
**ไฟล์**: `interfaces/api/handlers/user_handler.go:32`

**คำอธิบาย**: สร้างบัญชีผู้ใช้ใหม่

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "username": "username",
  "full_name": "John Doe"
}
```

**Response**:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "username": "username",
    "full_name": "John Doe",
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับข้อมูลจาก request body แล้วทำ validation
2. เรียก `userService.Register()` เพื่อสร้างบัญชี
3. Hash password ก่อนบันทึกลง database
4. ส่งข้อมูล user กลับโดยไม่รวม password

**Validation**:
- Email ต้องเป็นรูปแบบอีเมลที่ถูกต้อง
- Password ต้องมีความยาวขั้นต่ำ
- Username ต้องไม่ซ้ำในระบบ

---

### 1.2 POST /api/v1/auth/login
**ไฟล์**: `interfaces/api/handlers/user_handler.go:66`

**คำอธิบาย**: เข้าสู่ระบบด้วย email และ password

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "jwt-token-here",
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "username",
      "full_name": "John Doe"
    }
  }
}
```

**การทำงาน**:
1. รับ credentials จาก request
2. ตรวจสอบ email และ password กับฐานข้อมูล
3. สร้าง JWT token สำหรับ session
4. ส่ง token และข้อมูล user กลับ

**Authentication Flow**:
```
Client -> Login Endpoint -> Validate Credentials -> Generate JWT -> Return Token
```

---

### 1.3 POST /api/v1/auth/exchange
**ไฟล์**: `interfaces/api/handlers/oauth_handler.go:157`

**คำอธิบาย**: แลกเปลี่ยน OAuth authorization code เป็น JWT token

**Request Body**:
```json
{
  "code": "authorization-code",
  "state": "csrf-protection-state"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Token exchanged successfully",
  "data": {
    "token": "jwt-token-here",
    "is_new_user": false,
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "username": "username"
    }
  }
}
```

**การทำงาน**:
1. รับ code และ state จาก OAuth callback
2. Validate code กับ temporary store
3. แลก code เป็น JWT token
4. ระบุว่าเป็นผู้ใช้ใหม่หรือไม่

**OAuth Flow**:
```
Frontend -> /auth/google -> Google OAuth -> Callback -> Exchange Code -> JWT Token
```

---

## 2. Posts

### 2.1 POST /api/v1/posts
**ไฟล์**: `interfaces/api/handlers/post_handler.go:38`

**คำอธิบาย**: สร้างโพสต์ใหม่

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "title": "My First Post",
  "content": "This is the content of my post",
  "tags": ["golang", "programming"],
  "media": [
    {
      "url": "https://cdn.example.com/image.jpg",
      "type": "image",
      "mime_type": "image/jpeg"
    }
  ]
}
```

**Response**:
```json
{
  "success": true,
  "message": "Post created successfully",
  "data": {
    "id": "uuid",
    "title": "My First Post",
    "content": "This is the content...",
    "author": {
      "id": "uuid",
      "username": "john_doe"
    },
    "tags": ["golang", "programming"],
    "media": [...],
    "vote_count": 0,
    "comment_count": 0,
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. ดึง userID จาก authentication context
2. Validate request body (title, content, tags)
3. สร้างโพสต์พร้อม media และ tags
4. บันทึกลง database
5. ส่งข้อมูลโพสต์ที่สร้างกลับ

**Features**:
- รองรับ multiple media files (รูปภาพ, วิดีโอ)
- รองรับ tags สำหรับจัดหมวดหมู่
- คำนวณ vote และ comment count แบบ real-time

---

### 2.2 POST /api/v1/posts/:id/crosspost
**ไฟล์**: `interfaces/api/handlers/post_handler.go:352`

**คำอธิบาย**: สร้าง crosspost (แชร์โพสต์ไปยัง profile ของตัวเอง)

**Authentication**: ✅ Required

**URL Parameters**:
- `id` (UUID) - ID ของโพสต์ต้นทาง

**Request Body**:
```json
{
  "title": "Custom title for crosspost",
  "content": "Additional comment on this crosspost"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Crosspost created successfully",
  "data": {
    "id": "new-post-uuid",
    "original_post": {
      "id": "original-post-uuid",
      "author": {...}
    },
    "author": {
      "id": "current-user-uuid"
    },
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. ตรวจสอบว่าโพสต์ต้นทางมีอยู่จริง
2. สร้างโพสต์ใหม่ที่อ้อิงถึงโพสต์ต้นทาง
3. เก็บ reference ไปยังโพสต์เดิม
4. ผู้ใช้สามารถเพิ่มความคิดเห็นของตัวเองได้

---

## 3. Comments

### 3.1 POST /api/v1/comments
**ไฟล์**: `interfaces/api/handlers/comment_handler.go:26`

**คำอธิบาย**: สร้างคอมเมนต์หรือตอบกลับคอมเมนต์

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "post_id": "post-uuid",
  "parent_id": "parent-comment-uuid", // optional: สำหรับการตอบกลับ
  "content": "This is my comment"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Comment created successfully",
  "data": {
    "id": "comment-uuid",
    "post_id": "post-uuid",
    "parent_id": "parent-comment-uuid",
    "content": "This is my comment",
    "author": {
      "id": "uuid",
      "username": "john_doe"
    },
    "vote_count": 0,
    "reply_count": 0,
    "depth": 1,
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับข้อมูล post_id และ content (parent_id optional)
2. Validate ว่าโพสต์มีอยู่จริง
3. ถ้ามี parent_id จะสร้างเป็น reply (nested comment)
4. บันทึก comment พร้อมคำนวณ depth level
5. อัพเดท comment_count ของโพสต์

**Features**:
- รองรับ nested comments (แบบมีลำดับชั้น)
- คำนวณ depth level อัตโนมัติ
- ระบบ vote สำหรับคอมเมนต์

---

## 4. Votes

### 4.1 POST /api/v1/votes
**ไฟล์**: `interfaces/api/handlers/vote_handler.go:25`

**คำอธิบาย**: โหวต (upvote/downvote) โพสต์หรือคอมเมนต์

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "target_id": "post-or-comment-uuid",
  "target_type": "post", // หรือ "comment"
  "vote_type": 1 // 1 = upvote, -1 = downvote
}
```

**Response**:
```json
{
  "success": true,
  "message": "Voted successfully",
  "data": {
    "id": "vote-uuid",
    "target_id": "target-uuid",
    "target_type": "post",
    "vote_type": 1,
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับข้อมูล target (post/comment) และ vote_type
2. ตรวจสอบว่าผู้ใช้โหวตไปแล้วหรือยง
3. ถ้าโหวตแล้ว:
   - ถ้า vote_type เดิม = ใหม่ -> ลบโหวต (unvote)
   - ถ้า vote_type เดิม ≠ ใหม่ -> เปลี่ยนโหวต
4. ถ้ายังไม่โหวต -> สร้างโหวตใหม่
5. อัพเดท vote_count ของ target แบบ real-time

**Business Logic**:
```
เดิมไม่มีโหวต + upvote -> vote_count +1
เดิม upvote + upvote -> ลบโหวต (vote_count -1)
เดิม upvote + downvote -> เปลี่ยนเป็น downvote (vote_count -2)
เดิม downvote + upvote -> เปลี่ยนเป็น upvote (vote_count +2)
```

---

## 5. Follows

### 5.1 POST /api/v1/follows/user/:userId
**ไฟล์**: `interfaces/api/handlers/follow_handler.go:24`

**คำอธิบาย**: ติดตามผู้ใช้คนอื่น

**Authentication**: ✅ Required

**URL Parameters**:
- `userId` (UUID) - ID ของผู้ใช้ที่ต้องการติดตาม

**Response**:
```json
{
  "success": true,
  "message": "Followed successfully",
  "data": {
    "id": "follow-uuid",
    "follower_id": "current-user-uuid",
    "following_id": "target-user-uuid",
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับ userId จาก URL parameter
2. ตรวจสอบว่าผู้ใช้นั้นมีอยู่จริง
3. ตรวจสอบว่าไม่ได้ติดตามตัวเอง
4. สร้าง follow relationship
5. อัพเดท follower_count และ following_count

**Business Rules**:
- ไม่สามารถติดตามตัวเองได้
- ถ้าติดตามอยู่แล้ว จะไม่สร้าง duplicate record
- สามารถ unfollow ได้โดยใช้ DELETE endpoint

---

## 6. Saved Posts

### 6.1 POST /api/v1/saved/posts/:postId
**ไฟล์**: `interfaces/api/handlers/saved_post_handler.go:24`

**คำอธิบาย**: บันทึกโพสต์เพื่ออ่านภายหลัง

**Authentication**: ✅ Required

**URL Parameters**:
- `postId` (UUID) - ID ของโพสต์ที่ต้องการบันทึก

**Response**:
```json
{
  "success": true,
  "message": "Post saved successfully",
  "data": {
    "id": "saved-post-uuid",
    "user_id": "current-user-uuid",
    "post_id": "post-uuid",
    "saved_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับ postId จาก URL parameter
2. ตรวจสอบว่าโพสต์มีอยู่จริง
3. สร้างการบันทึกโพสต์
4. เก็บ timestamp สำหรับจัดเรียง

**Features**:
- สามารถบันทึกโพสต์ได้ไม่จำกัด
- มี endpoint สำหรับดูโพสต์ที่บันทึกทั้งหมด
- สามารถ unsave ได้โดยใช้ DELETE endpoint

---

## 7. Media Upload

### 7.1 POST /api/v1/media/upload/image
**ไฟล์**: `interfaces/api/handlers/media_handler.go:24`

**คำอธิบาย**: อัปโหลดรูปภาพ

**Authentication**: ✅ Required

**Request**: multipart/form-data
- `image`: ไฟล์รูปภาพ

**Response**:
```json
{
  "success": true,
  "message": "Image uploaded successfully",
  "data": {
    "id": "media-uuid",
    "type": "image",
    "url": "https://cdn.example.com/image.jpg",
    "thumbnail": "https://cdn.example.com/thumb.jpg",
    "width": 1920,
    "height": 1080,
    "size": 1024000,
    "mime_type": "image/jpeg"
  }
}
```

**การทำงาน**:
1. รับไฟล์จาก multipart form
2. Validate file type (jpeg, png, gif, webp)
3. Validate file size (max 10MB)
4. สร้าง thumbnail อัตโนมัติ
5. อัปโหลดไปยัง cloud storage (Bunny CDN/R2)
6. บันทึก metadata ลง database

**Supported Formats**:
- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- WebP (.webp)

**File Size Limit**: 10MB

---

### 7.2 POST /api/v1/media/upload/video
**ไฟล์**: `interfaces/api/handlers/media_handler.go:53`

**คำอธิบาย**: อัปโหลดวิดีโอ

**Authentication**: ✅ Required

**Request**: multipart/form-data
- `video`: ไฟล์วิดีโอ

**Response**:
```json
{
  "success": true,
  "message": "Video uploaded successfully",
  "data": {
    "id": "media-uuid",
    "type": "video",
    "url": "https://cdn.example.com/video.mp4",
    "thumbnail": "https://cdn.example.com/thumb.jpg",
    "width": 1920,
    "height": 1080,
    "duration": 120.5,
    "size": 50000000,
    "mime_type": "video/mp4",
    "encoding_status": "processing"
  }
}
```

**การทำงาน**:
1. รับไฟล์วิดีโอจาก multipart form
2. Validate file type (mp4, webm, ogg)
3. Validate file size (max 300MB)
4. อัปโหลดไปยัง R2 storage
5. สร้าง media record พร้อมสถานะ "processing"
6. ระบบจะประมวลผลวิดีโออัตโนมัติ (ถ้ามี Bunny Stream)

**Supported Formats**:
- MP4 (.mp4)
- WebM (.webm)
- OGG (.ogg)

**File Size Limit**: 300MB

**Video Processing**:
- ระบบจะประมวลผลวิดีโอใน background
- สร้าง thumbnail อัตโนมัติ
- สามารถติดตาม encoding status ผ่าน WebSocket

---

## 8. Push Notifications

### 8.1 POST /api/v1/push/subscribe
**ไฟล์**: `interfaces/api/handlers/push_handler.go:23`

**คำอธิบาย**: ลงทะเบียนรับ push notifications

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "endpoint": "https://fcm.googleapis.com/fcm/send/...",
  "keys": {
    "p256dh": "public-key",
    "auth": "auth-secret"
  },
  "device_type": "web" // web, ios, android
}
```

**Response**:
```json
{
  "success": true,
  "message": "Subscription saved successfully",
  "data": {
    "id": "subscription-uuid",
    "user_id": "user-uuid",
    "endpoint": "https://...",
    "device_type": "web",
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. รับ subscription data จาก client
2. ตรวจสอบว่า endpoint ถูกต้อง
3. บันทึก subscription ลง database
4. ผูก subscription กับ user

**Supported Device Types**:
- `web` - Web push (Chrome, Firefox, Safari)
- `ios` - iOS APNs
- `android` - Android FCM

---

### 8.2 POST /api/v1/push/unsubscribe
**ไฟล์**: `interfaces/api/handlers/push_handler.go:49`

**คำอธิบาย**: ยกเลิกการรับ push notifications

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "endpoint": "https://fcm.googleapis.com/fcm/send/..."
}
```

**Response**:
```json
{
  "success": true,
  "message": "Subscription removed successfully",
  "data": null
}
```

**การทำงาน**:
1. รับ endpoint ที่ต้องการยกเลิก
2. ค้นหา subscription ที่ตรงกับ endpoint และ user
3. ลบ subscription จาก database

---

## 9. Chat & Messaging

### 9.1 POST /api/v1/chat/conversations/:conversationId/messages
**ไฟล์**: `interfaces/api/handlers/message_handler.go:40`

**คำอธิบาย**: ส่งข้อความในการสนทนา

**Authentication**: ✅ Required

**URL Parameters**:
- `conversationId` (UUID) - ID ของการสนทนา

**Request Body** (Text Message):
```json
{
  "content": "Hello, how are you?",
  "type": "text"
}
```

**Request Body** (Media Message - multipart/form-data):
- `type`: "image" | "video" | "file"
- `content`: Caption (optional)
- `media[]`: ไฟล์ media (1-10 files for images, 1 for video, 1-5 for files)

**Response**:
```json
{
  "success": true,
  "message": "Message sent successfully",
  "data": {
    "id": "message-uuid",
    "conversation_id": "conversation-uuid",
    "sender": {
      "id": "sender-uuid",
      "username": "john_doe"
    },
    "receiver": {
      "id": "receiver-uuid",
      "username": "jane_doe"
    },
    "type": "text",
    "content": "Hello, how are you?",
    "media": [],
    "is_read": false,
    "created_at": "timestamp"
  }
}
```

**การทำงาน**:
1. ตรวจสอบว่า conversationId ถูกต้องและผู้ใช้เป็นส่วนหนึ่งของการสนทนา
2. ตรวจสอบ Content-Type:
   - `application/json` -> Text message
   - `multipart/form-data` -> Media message
3. สำหรับ media:
   - Validate file type และ size
   - อัปโหลดไปยัง storage
   - สร้าง media record
4. บันทึก message ลง database
5. ส่ง real-time notification ผ่าน WebSocket
6. อัพเดท last_message และ unread_count ของ conversation

**Message Types**:
- `text` - ข้อความธรรมดา
- `image` - รูปภาพ (max 10 files, 20MB each)
- `video` - วิดีโอ (max 1 file, 500MB)
- `file` - ไฟล์เอกสาร (max 5 files, 100MB each)

**Media Validation**:
```
Image: .jpg, .jpeg, .png, .gif, .webp (max 20MB each)
Video: .mp4, .mov, .mkv (max 500MB)
File: .pdf, .doc, .docx, .xls, .xlsx, .zip, .txt (max 100MB each)
```

---

### 9.2 POST /api/v1/chat/conversations/:conversationId/read
**ไฟล์**: `interfaces/api/handlers/conversation_handler.go:95`

**คำอธิบาย**: ทำเครื่องหมายข้อความทั้งหมดในการสนทนาว่าอ่านแล้ว

**Authentication**: ✅ Required

**URL Parameters**:
- `conversationId` (UUID) - ID ของการสนทนา

**Response**:
```json
{
  "success": true,
  "message": "Conversation marked as read",
  "data": null
}
```

**การทำงาน**:
1. ตรวจสอบว่าผู้ใช้เป็นส่วนหนึ่งของการสนทนา
2. อัพเดท `is_read = true` สำหรับข้อความทั้งหมดที่ผู้ใช้ยังไม่ได้อ่าน
3. อัพเดท `unread_count` ของ conversation เป็น 0
4. ส่ง read receipt notification ผ่าน WebSocket ไปยังผู้ส่ง

**WebSocket Event**:
```json
{
  "type": "message.read_update",
  "payload": {
    "conversationId": "uuid",
    "readBy": "user-uuid",
    "readAt": "timestamp"
  }
}
```

---

### 9.3 POST /api/v1/chat/blocks/
**ไฟล์**: `interfaces/api/handlers/block_handler.go:26`

**คำอธิบาย**: บล็อกผู้ใช้ (ป้องกันการส่งข้อความ)

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "username": "user_to_block"
}
```

**Response**:
```json
{
  "success": true,
  "message": "User blocked successfully",
  "data": null
}
```

**การทำงาน**:
1. ค้นหาผู้ใช้จาก username
2. สร้าง block relationship
3. ลบ conversation ที่มีอยู่ (optional)
4. ป้องกันการสร้าง conversation ใหม่

**Effects of Blocking**:
- ไม่สามารถส่งข้อความหากันได้
- ไม่สามารถเห็นโพสต์และคอมเมนต์ของกัน
- ลบ follow relationship (ถ้ามี)

---

## 10. File Upload

### 10.1 POST /api/v1/upload/file
**ไฟล์**: `interfaces/api/handlers/file_upload_handler.go:21`

**คำอธิบาย**: อัปโหลดไฟล์โดยตรง (direct upload)

**Authentication**: ✅ Required

**Request**: multipart/form-data
- `file`: ไฟล์ที่ต้องการอัปโหลด

**Response**:
```json
{
  "id": "media-uuid",
  "filename": "document.pdf",
  "extension": "pdf",
  "size": 1024000,
  "mime_type": "application/pdf",
  "url": "https://cdn.example.com/file.pdf",
  "uploaded_at": "timestamp"
}
```

**การทำงาน**:
1. รับไฟล์จาก multipart form
2. Validate file type และ size (max 50MB)
3. อัปโหลดไปยัง storage
4. สร้าง media record
5. ส่ง URL กลับทันที

**Supported File Types**:
- Documents: .pdf, .doc, .docx, .xls, .xlsx
- Archives: .zip, .rar
- Text: .txt

---

### 10.2 POST /api/v1/upload/presigned-url
**ไฟล์**: `interfaces/api/handlers/presigned_upload_handler.go:59`

**คำอธิบาย**: สร้าง presigned URL สำหรับอัปโหลดไฟล์จาก client โดยตรง

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "filename": "photo.jpg",
  "contentType": "image/jpeg",
  "fileSize": 1024000,
  "mediaType": "image"
}
```

**Response**:
```json
{
  "uploadUrl": "https://r2.cloudflare.com/presigned-url",
  "fileUrl": "https://cdn.example.com/images/uuid/photo.jpg",
  "fileKey": "images/user-uuid/media-uuid.jpg",
  "mediaId": "media-uuid",
  "expiresAt": "timestamp"
}
```

**การทำงาน**:
1. รับข้อมูลไฟล์ที่ต้องการอัปโหลด
2. Validate file size และ extension
3. สร้าง unique file key
4. สร้าง presigned URL สำหรับอัปโหลด (valid 15 นาที)
5. ส่ง URL กลับให้ client อัปโหลดเองโดยตรง

**Advantages**:
- ลดโหลดบน server (client อัปโหลดตรงไปยัง storage)
- รองรับไฟล์ขนาดใหญ่
- มี progress tracking
- ไม่มี timeout issues

**Upload Flow**:
```
1. Client ขอ presigned URL
2. Server สร้าง URL (valid 15 min)
3. Client อัปโหลดไฟล์โดยตรงไปยัง R2
4. Client เรียก /upload/confirm เพื่อแจ้ง server
5. Server สร้าง media record
```

---

### 10.3 POST /api/v1/upload/presigned-url-batch
**ไฟล์**: `interfaces/api/handlers/presigned_upload_handler.go:142`

**คำอธิบาย**: สร้าง presigned URLs หลายไฟล์พร้อมกัน (max 200 files)

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "files": [
    {
      "filename": "photo1.jpg",
      "contentType": "image/jpeg",
      "fileSize": 1024000,
      "mediaType": "image"
    },
    {
      "filename": "photo2.jpg",
      "contentType": "image/jpeg",
      "fileSize": 2048000,
      "mediaType": "image"
    }
  ]
}
```

**Response**:
```json
{
  "uploads": [
    {
      "uploadUrl": "https://r2.cloudflare.com/presigned-url-1",
      "fileUrl": "https://cdn.example.com/images/uuid/photo1.jpg",
      "fileKey": "images/user-uuid/media-uuid-1.jpg",
      "mediaId": "media-uuid-1",
      "expiresAt": "timestamp"
    },
    {
      "uploadUrl": "https://r2.cloudflare.com/presigned-url-2",
      "fileUrl": "https://cdn.example.com/images/uuid/photo2.jpg",
      "fileKey": "images/user-uuid/media-uuid-2.jpg",
      "mediaId": "media-uuid-2",
      "expiresAt": "timestamp"
    }
  ],
  "total": 2
}
```

**การทำงาน**:
1. รับ array ของไฟล์ที่ต้องการอัปโหลด (max 200 files)
2. Validate แต่ละไฟล์
3. สร้าง presigned URLs สำหรับทุกไฟล์
4. ส่ง URLs ทั้งหมดกลับพร้อมกัน

**Use Cases**:
- อัปโหลดรูปภาพหลายรูปในโพสต์
- อัปโหลดไฟล์แนบหลายไฟล์
- Bulk upload operations

---

### 10.4 POST /api/v1/upload/confirm
**ไฟล์**: `interfaces/api/handlers/presigned_upload_handler.go:278`

**คำอธิบาย**: ยืนยันว่าอัปโหลดไฟล์สำเร็จแล้ว

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "mediaId": "media-uuid",
  "fileKey": "images/user-uuid/media-uuid.jpg",
  "fileSize": 1024000,
  "contentType": "image/jpeg",
  "width": 1920,
  "height": 1080,
  "sourceType": "post",
  "sourceId": "post-uuid"
}
```

**Response**:
```json
{
  "success": true,
  "mediaId": "media-uuid",
  "message": "upload confirmed",
  "fileUrl": "https://cdn.example.com/images/uuid/photo.jpg"
}
```

**การทำงาน**:
1. รับข้อมูล metadata หลังจากอัปโหลดเสร็จ
2. สร้าง media record ใน database
3. เชื่อมโยงกับ source (post, message, etc.)
4. ส่ง final URL กลับ

---

### 10.5 POST /api/v1/upload/confirm-batch
**ไฟล์**: `interfaces/api/handlers/presigned_upload_handler.go:395`

**คำอธิบาย**: ยืนยันการอัปโหลดหลายไฟล์พร้อมกัน

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "uploads": [
    {
      "mediaId": "media-uuid-1",
      "fileKey": "images/user-uuid/media-uuid-1.jpg",
      "fileSize": 1024000,
      "contentType": "image/jpeg"
    },
    {
      "mediaId": "media-uuid-2",
      "fileKey": "images/user-uuid/media-uuid-2.jpg",
      "fileSize": 2048000,
      "contentType": "image/jpeg"
    }
  ]
}
```

**Response**:
```json
{
  "successful": [
    {
      "mediaId": "media-uuid-1",
      "fileUrl": "https://cdn.example.com/images/uuid/photo1.jpg",
      "fileKey": "images/user-uuid/media-uuid-1.jpg",
      "success": true
    },
    {
      "mediaId": "media-uuid-2",
      "fileUrl": "https://cdn.example.com/images/uuid/photo2.jpg",
      "fileKey": "images/user-uuid/media-uuid-2.jpg",
      "success": true
    }
  ],
  "failed": [],
  "total": 2,
  "successCount": 2,
  "failCount": 0
}
```

**การทำงาน**:
1. รับ array ของไฟล์ที่อัปโหลดเสร็จแล้ว
2. สร้าง media records สำหรับแต่ละไฟล์
3. แยกผลลัพธ์เป็น successful และ failed
4. ส่ง summary กลับ

---

## 11. Webhooks

### 11.1 POST /api/v1/webhooks/bunny/video-status
**ไฟล์**: `interfaces/api/handlers/webhook_handler.go:35`

**คำอธิบาย**: รับ webhook จาก Bunny Stream เกี่ยวกับสถานะการ encode วิดีโอ

**Authentication**: ❌ Not Required (public webhook)

**Request Body** (from Bunny Stream):
```json
{
  "VideoLibraryId": 12345,
  "VideoGuid": "video-uuid",
  "Status": 3,
  "EncodeProgress": 100,
  "Width": 1920,
  "Height": 1080,
  "Length": 120,
  "AvailableResolutions": "240p,360p,720p,1080p"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Webhook received"
}
```

**การทำงาน**:
1. รับ webhook payload จาก Bunny Stream
2. ตอบกลับ 200 OK ทันที (ไม่ให้ Bunny รอ)
3. ประมวลผลใน background goroutine:
   - อัพเดท encoding_status และ progress
   - อัพเดท video metadata (width, height, duration)
   - ส่ง WebSocket notification ไปยังผู้ใช้
   - ถ้าเสร็จสมบูรณ์:
     - อัพเดท message/post ที่เกี่ยวข้อง
     - Auto-publish draft posts

**Bunny Stream Status Codes**:
```
0 = Queued
1 = Processing
2 = Encoding
3 = Finished (Ready to play)
4 = Resolution Finished
5 = Failed
6 = PresignedUploadStarted
7 = PresignedUploadFinished
8 = PresignedUploadFailed
```

**WebSocket Notifications**:
- `video.encoding.progress` - อัพเดทความคืบหน้า
- `video.encoding.completed` - เสร็จสมบูรณ์
- `video.encoding.failed` - ล้มเหลว

---

## 12. Legacy Endpoints

### 12.1 POST /api/v1/tasks/
**ไฟล์**: `interfaces/api/handlers/task_handler.go:23`

**คำอธิบาย**: สร้าง task ใหม่ (legacy endpoint สำหรับ task management)

**Authentication**: ✅ Required

**Request Body**:
```json
{
  "title": "Complete documentation",
  "description": "Write API documentation",
  "status": "pending",
  "priority": "high"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "id": "task-uuid",
    "title": "Complete documentation",
    "description": "Write API documentation",
    "status": "pending",
    "priority": "high",
    "user": {...},
    "created_at": "timestamp"
  }
}
```

**หมายเหตุ**: Endpoint นี้เป็น legacy และอาจถูกลบออกในอนาคต

---

### 12.2 POST /api/v1/files/upload
**ไฟล์**: `interfaces/api/handlers/file_handler.go:34`

**คำอธิบาย**: อัปโหลดไฟล์ (legacy endpoint)

**Authentication**: ✅ Required

**Request**: multipart/form-data
- `file`: ไฟล์ที่ต้องการอัปโหลด
- `custom_path`: (optional) กำหนด path เอง
- `category`: (optional) หมวดหมู่
- `entity_id`: (optional) ID ของ entity
- `file_type`: (optional) ประเภทไฟล์

**Response**:
```json
{
  "success": true,
  "message": "File uploaded successfully",
  "data": {
    "fileId": "file-uuid",
    "fileName": "document.pdf",
    "url": "https://cdn.example.com/file.pdf",
    "cdnPath": "/files/...",
    "fileSize": 1024000,
    "mimeType": "application/pdf",
    "pathType": "structured"
  }
}
```

**หมายเหตุ**: ควรใช้ `/api/v1/upload/file` หรือ presigned URL endpoints แทน

---

## 13. Admin Jobs

### 13.1 POST /api/v1/jobs/
**ไฟล์**: `interfaces/api/handlers/job_handler.go:23`

**คำอธิบาย**: สร้าง background job ใหม่ (admin only)

**Authentication**: ✅ Required + Admin Role

**Request Body**:
```json
{
  "name": "cleanup_old_media",
  "schedule": "0 2 * * *",
  "enabled": true,
  "config": {
    "retention_days": 90
  }
}
```

**Response**:
```json
{
  "success": true,
  "message": "Job created successfully",
  "data": {
    "id": "job-uuid",
    "name": "cleanup_old_media",
    "schedule": "0 2 * * *",
    "enabled": true,
    "last_run": null,
    "next_run": "timestamp",
    "created_at": "timestamp"
  }
}
```

**Use Cases**:
- Scheduled cleanup tasks
- Data processing jobs
- Report generation
- Maintenance operations

---

### 13.2 POST /api/v1/jobs/:id/start
**ไฟล์**: `interfaces/api/handlers/job_handler.go:99`

**คำอธิบาย**: เริ่มการทำงานของ job (admin only)

**Authentication**: ✅ Required + Admin Role

**URL Parameters**:
- `id` (UUID) - ID ของ job

**Response**:
```json
{
  "success": true,
  "message": "Job started successfully",
  "data": null
}
```

---

### 13.3 POST /api/v1/jobs/:id/stop
**ไฟล์**: `interfaces/api/handlers/job_handler.go:114`

**คำอธิบาย**: หยุดการทำงานของ job (admin only)

**Authentication**: ✅ Required + Admin Role

**URL Parameters**:
- `id` (UUID) - ID ของ job

**Response**:
```json
{
  "success": true,
  "message": "Job stopped successfully",
  "data": null
}
```

---

## สรุปและข้อสังเกต

### 1. Architecture Patterns

#### 1.1 Clean Architecture
ระบบใช้ Clean Architecture แบบแบ่งชั้นชัดเจน:
```
interfaces/api/handlers -> domain/services -> domain/repositories
```

#### 1.2 Dependency Injection
ทุก handler รับ service ผ่าน constructor:
```go
func NewPostHandler(postService services.PostService) *PostHandler {
    return &PostHandler{postService: postService}
}
```

#### 1.3 Repository Pattern
แยก business logic (services) ออกจาก data access (repositories)

---

### 2. Security & Authentication

#### 2.1 JWT Authentication
- ส่วนใหญ่ใช้ JWT Bearer Token
- Token ถูก validate ผ่าน middleware
- UserID ถูกเก็บใน `c.Locals("userID")`

#### 2.2 OAuth 2.0 Flow
- รองรับ Google OAuth
- ใช้ authorization code flow
- มี CSRF protection ผ่าน state parameter

#### 2.3 Input Validation
- ใช้ struct tags สำหรับ validation
- มี custom validators
- Error messages ชัดเจนและเป็นมิตรกับผู้ใช้

---

### 3. File Upload Strategies

ระบบมี 3 วิธีในการอัปโหลดไฟล์:

#### 3.1 Direct Upload
```
POST /api/v1/upload/file
```
- Server รับไฟล์และอัปโหลดเอง
- เหมาะสำหรับไฟล์ขนาดเล็ก
- มี timeout risk สำหรับไฟล์ใหญ่

#### 3.2 Presigned URL (Single File)
```
POST /api/v1/upload/presigned-url
POST /api/v1/upload/confirm
```
- Client อัปโหลดตรงไปยัง R2
- ลดโหลดบน server
- รองรับไฟล์ขนาดใหญ่

#### 3.3 Presigned URL (Batch)
```
POST /api/v1/upload/presigned-url-batch
POST /api/v1/upload/confirm-batch
```
- อัปโหลดหลายไฟล์พร้อมกัน
- มี partial success handling
- เหมาะสำหรับ bulk operations

**แนะนำ**: ใช้ Presigned URL สำหรับไฟล์ใหญ่และ batch operations

---

### 4. Real-time Features

#### 4.1 WebSocket Notifications
ระบบมี 2 WebSocket hubs:
- **ChatHub**: สำหรับข้อความ chat
- **NotificationHub**: สำหรับ notifications ทั่วไป

#### 4.2 Events
```
- message.new - ข้อความใหม่
- message.read_update - อ่านข้อความแล้ว
- video.encoding.progress - progress การ encode
- video.encoding.completed - encode เสร็จ
- video.encoding.failed - encode ล้มเหลว
```

---

### 5. Media Processing

#### 5.1 Image Processing
- สร้าง thumbnail อัตโนมัติ
- รองรับหลาย format (JPEG, PNG, GIF, WebP)
- Max size: 10MB

#### 5.2 Video Processing
- อัปโหลดไปยัง R2 storage
- Optional: ส่งไปยัง Bunny Stream สำหรับ encoding
- สร้าง thumbnail จาก frame
- Track encoding progress
- Max size: 300MB (direct) / 500MB (presigned)

#### 5.3 Video Encoding Workflow
```
1. Upload video to R2
2. Create media record (status: "processing")
3. Optionally send to Bunny Stream
4. Bunny sends webhook updates
5. Update encoding status via webhook
6. Notify user via WebSocket
7. Auto-publish related content when done
```

---

### 6. Data Validation

#### 6.1 Request Validation
ทุก endpoint มี validation:
```go
if err := utils.ValidateStruct(&req); err != nil {
    errors := utils.GetValidationErrors(err)
    return c.Status(fiber.StatusBadRequest).JSON(...)
}
```

#### 6.2 Validation Rules
- **Email**: รูปแบบ email ที่ถูกต้อง
- **Password**: ความยาวขั้นต่ำ
- **File Size**: จำกัดตาม type
- **File Type**: ตรวจสอบ MIME type
- **UUID**: ตรวจสอบรูปแบบ UUID

---

### 7. Error Handling

#### 7.1 Error Response Format
```json
{
  "success": false,
  "message": "Error description",
  "errors": {
    "field": "Field-specific error"
  }
}
```

#### 7.2 HTTP Status Codes
- `200` - Success
- `400` - Bad Request / Validation Error
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

---

### 8. Performance Optimizations

#### 8.1 Pagination
ทุก list endpoints รองรับ pagination:
```
?offset=0&limit=20
```

#### 8.2 Cursor-based Pagination
สำหรับ real-time data (messages, notifications):
```
?cursor=timestamp&limit=50
```

#### 8.3 CDN & Caching
- ใช้ Cloudflare R2 + CDN
- Static assets ถูก cache
- Presigned URLs มี expiration

---

### 9. Testing Recommendations

#### 9.1 Unit Tests
ควร test:
- Handler logic
- Service business logic
- Repository queries
- Validation rules

#### 9.2 Integration Tests
ควร test:
- API endpoints
- Database operations
- File upload flows
- WebSocket connections

#### 9.3 Load Tests
ควร test:
- Concurrent file uploads
- Real-time messaging
- Media processing load

---

### 10. API Best Practices

#### ✅ สิ่งที่ทำดีแล้ว
1. **Consistent Response Format**: ทุก endpoint ใช้รูปแบบเดียวกัน
2. **Clear Error Messages**: ข้อความ error ชัดเจน
3. **Proper HTTP Status Codes**: ใช้ status codes อย่างถูกต้อง
4. **Input Validation**: มี validation ครบถ้วน
5. **Authentication Middleware**: มี middleware ป้องกัน unauthorized access
6. **Swagger Documentation**: มี API docs (บางส่วน)
7. **UUID for IDs**: ใช้ UUID แทน sequential IDs (ปลอดภัยกว่า)

#### ⚠️ ข้อควรปรับปรุง
1. **Rate Limiting**: ควรเพิ่ม rate limiting สำหรับ public endpoints
2. **API Versioning**: มี `/api/v1` แต่ยังไม่มีแผนสำหรับ v2
3. **Webhook Signature**: Webhook ควรมี signature verification
4. **Idempotency**: ควรรองรับ idempotency keys สำหรับ POST requests
5. **Partial Updates**: ควรใช้ PATCH แทน PUT สำหรับ partial updates
6. **Soft Delete**: พิจารณาใช้ soft delete แทนการลบถาวร
7. **Audit Logs**: ควรมี audit logs สำหรับ sensitive operations

---

### 11. Security Considerations

#### ✅ มาตรการความปลอดภัย
1. JWT authentication
2. Password hashing
3. Input validation
4. CORS configuration
5. UUID instead of sequential IDs

#### ⚠️ ควรเพิ่ม
1. **Rate Limiting**: ป้องกัน brute force และ DDoS
2. **CSRF Protection**: สำหรับ web clients
3. **API Key Rotation**: สำหรับ third-party integrations
4. **Request Logging**: log ทุก request สำหรับ audit
5. **Webhook Signatures**: verify ว่า webhook มาจากแหล่งที่ถูกต้อง
6. **File Scanning**: scan uploaded files สำหรับ malware
7. **Content Security Policy**: headers สำหรับ XSS protection

---

### 12. Scalability Considerations

#### 12.1 Database
- ใช้ connection pooling
- มี indexes สำหรับ query ที่ใช้บ่อย
- Pagination ทุก list endpoints

#### 12.2 File Storage
- ใช้ cloud storage (R2)
- CDN สำหรับ static assets
- Presigned URLs ลดโหลดบน server

#### 12.3 Background Jobs
- Video encoding ทำงานแบบ async
- Webhook processing ไม่ block main thread
- มี job queue system (jobs endpoints)

---

### 13. Monitoring & Observability

#### ควรเพิ่ม
1. **Metrics**: track API response times, error rates
2. **Distributed Tracing**: trace requests ผ่านทั้ง system
3. **Health Check Endpoint**: `/health` สำหรับ monitoring
4. **Structured Logging**: log ในรูปแบบ JSON
5. **Error Tracking**: integration กับ Sentry หรือ similar

---

## สรุปท้าย

### จำนวน POST Endpoints ตามหมวดหมู่

| หมวดหมู่ | จำนวน Endpoints |
|---------|----------------|
| Authentication | 3 |
| Posts | 2 |
| Comments | 1 |
| Votes | 1 |
| Follows | 1 |
| Saved Posts | 1 |
| Media Upload | 2 |
| Push Notifications | 2 |
| Chat & Messaging | 3 |
| File Upload | 5 |
| Webhooks | 1 |
| Legacy | 2 |
| Admin Jobs | 3 |
| **รวมทั้งหมด** | **32** |

### คุณสมบัติเด่นของระบบ

1. **Social Media Platform**: ครบครันทั้ง posts, comments, votes, follows
2. **Real-time Chat**: รองรับข้อความแบบ real-time ผ่าน WebSocket
3. **Media Processing**: ประมวลผลรูปภาพและวิดีโออัตโนมัติ
4. **Multiple Upload Methods**: รองรับทั้ง direct upload และ presigned URLs
5. **OAuth Integration**: รองรับ Google OAuth
6. **Push Notifications**: ระบบ push notifications แบบครบวงจร
7. **Admin Tools**: มี tools สำหรับ admin จัดการ background jobs

### ข้อเสนอแนะสำหรับการพัฒนาต่อ

1. เพิ่ม rate limiting และ security headers
2. ปรับปรุง documentation ให้ครบทุก endpoint
3. เพิ่ม monitoring และ observability
4. พัฒนา automated tests
5. เพิ่ม API versioning strategy
6. พิจารณา GraphQL สำหรับ complex queries
7. เพิ่ม caching layer (Redis) สำหรับข้อมูลที่ query บ่อย

---

**วันที่อัพเดทล่าสุด**: 2025-11-14
**เวอร์ชัน API**: v1
**จำนวน POST Endpoints**: 32 endpoints
