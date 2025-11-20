# 🤖 คู่มือระบบ AI Auto-Post

## 📋 สารบัญ
1. [ภาพรวมระบบ](#ภาพรวมระบบ)
2. [สถาปัตยกรรม](#สถาปัตยกรรม)
3. [การติดตั้งและตั้งค่า](#การติดตั้งและตั้งค่า)
4. [การใช้งาน](#การใช้งาน)
5. [API Endpoints](#api-endpoints)
6. [ตัวอย่างการใช้งาน](#ตัวอย่างการใช้งาน)

---

## 📖 ภาพรวมระบบ

ระบบ AI Auto-Post เป็นฟีเจอร์ที่ใช้ AI (OpenAI) ในการสร้างเนื้อหาโพสต์อัตโนมัติตามหัวข้อที่กำหนด โดยระบบจะ:

### ✨ คุณสมบัติหลัก
- ใช้ OpenAI GPT models (gpt-4o-mini, gpt-4, etc.) ในการ generate เนื้อหา
- รองรับการกำหนดหัวข้อ (topics) หลายๆ หัวข้อได้
- สามารถตั้งเวลาโพสต์อัตโนมัติได้ด้วย Cron Expression
- มี Scheduler ทำงานทุกชั่วโมง (หรือตามที่ตั้งค่า)
- สามารถ trigger โพสต์ manual ได้
- บันทึก log ทุกครั้งที่มีการสร้างโพสต์

### 🔄 Flow การทำงาน
```
1. ผู้ใช้สร้าง Auto-Post Setting
   ↓
2. กำหนด Bot User (บัญชีที่จะใช้โพสต์)
   ↓
3. กำหนด Topics (หัวข้อที่ต้องการให้ AI generate)
   ↓
4. ตั้งค่า Schedule (เช่น ทุกชั่วโมง, ทุก 2 ชม.)
   ↓
5. เปิดใช้งาน (Enable Setting)
   ↓
6. Scheduler จะรันตาม Cron Expression
   ↓
7. เลือก topic แบบสุ่มจากที่กำหนดไว้
   ↓
8. ส่งคำขอไปยัง OpenAI API
   ↓
9. AI generate: Title, Content, Tags
   ↓
10. สร้างโพสต์ในระบบ
   ↓
11. บันทึก log (success/failed)
```

---

## 🏗️ สถาปัตยกรรม

### ไฟล์ที่สร้างขึ้นใหม่

#### 1. **OpenAI Service** (`pkg/ai/openai_service.go`)
- จัดการการเชื่อมต่อกับ OpenAI API
- Generate เนื้อหาจาก prompts
- Parse response เป็น structured data (Title, Content, Tags)

#### 2. **Models** (`domain/models/auto_post_setting.go`)
```go
// AutoPostSetting - การตั้งค่าระบบ auto-post
type AutoPostSetting struct {
    ID              uuid.UUID
    BotUserID       uuid.UUID  // บัญชีที่ใช้โพสต์
    IsEnabled       bool       // เปิด/ปิดการทำงาน
    CronSchedule    string     // เช่น "0 * * * *" (ทุกชั่วโมง)
    Model           string     // gpt-4o-mini, gpt-4, etc.
    Topics          JSON       // ["topic1", "topic2", ...]
    MaxTokens       int        // จำนวน tokens สูงสุด
    // ... statistics fields
}

// AutoPostLog - บันทึกประวัติการสร้างโพสต์
type AutoPostLog struct {
    ID             uuid.UUID
    SettingID      uuid.UUID
    PostID         *uuid.UUID
    Topic          string     // หัวข้อที่ใช้ generate
    GeneratedTitle string
    Status         string     // pending, success, failed
    ErrorMessage   *string
    // ... timestamp fields
}
```

#### 3. **DTOs** (`domain/dto/auto_post.go`)
- Request/Response structures สำหรับ API
- Validation rules

#### 4. **Repositories**
- `domain/repositories/auto_post_repository.go` - Interfaces
- `infrastructure/postgres/auto_post_repository_impl.go` - Implementation
- CRUD operations สำหรับ settings และ logs

#### 5. **Services**
- `domain/services/auto_post_service.go` - Interface
- `application/serviceimpl/auto_post_service_impl.go` - Business logic
  - CreateSetting, UpdateSetting, DeleteSetting
  - GenerateAndPost - สร้างโพสต์จาก AI
  - ProcessAllEnabledSettings - ประมวลผล settings ที่เปิดใช้งานทั้งหมด

#### 6. **API Handlers & Routes**
- `interfaces/api/handlers/auto_post_handler.go` - HTTP handlers
- `interfaces/api/routes/auto_post_routes.go` - Route definitions

#### 7. **Database Migration** (`migrations/018_create_auto_post_tables.sql`)
- สร้าง table `auto_post_settings`
- สร้าง table `auto_post_logs`
- สร้าง indexes สำหรับ performance

---

## ⚙️ การติดตั้งและตั้งค่า

### ขั้นตอนที่ 1: ตั้งค่า OpenAI API Key

1. ไปที่ https://platform.openai.com/api-keys
2. สร้าง API Key ใหม่
3. เพิ่มใน `.env` file:

```env
# OpenAI Configuration
OPENAI_API_KEY=sk-your-actual-api-key-here
OPENAI_MODEL=gpt-4o-mini
```

> **หมายเหตุ:** `gpt-4o-mini` เป็น model ที่ประหยัดค่าใช้จ่ายที่สุด แต่ให้ผลลัพธ์ดี

### ขั้นตอนที่ 2: Run Database Migration

เมื่อรัน server ครั้งแรก migration จะทำงานอัตโนมัติ หรือสามารถรัน manual:

```bash
# Server จะ run migration อัตโนมัติตอน startup
go run cmd/api/main.go
```

ระบบจะสร้าง tables:
- `auto_post_settings`
- `auto_post_logs`

### ขั้นตอนที่ 3: สร้าง Bot User

สร้าง user account ที่จะใช้สำหรับโพสต์อัตโนมัติ:

```bash
# ตัวอย่าง: สร้างผ่าน API
POST /api/v1/auth/register
{
  "username": "ai_bot",
  "email": "aibot@example.com",
  "password": "secure_password"
}

# บันทึก user ID สำหรับใช้งานต่อไป
```

### ขั้นตอนที่ 4: สร้าง Auto-Post Setting

```bash
POST /api/v1/auto-post/settings
Authorization: Bearer <your-token>
{
  "botUserId": "uuid-of-bot-user",
  "isEnabled": true,
  "cronSchedule": "0 * * * *",  // ทุกชั่วโมง
  "model": "gpt-4o-mini",
  "topics": [
    "Technology trends in 2025",
    "Tips for software developers",
    "AI and machine learning innovations"
  ],
  "maxTokens": 1500
}
```

---

## 🎮 การใช้งาน

### การจัดการ Settings

#### ดู Setting ทั้งหมด
```bash
GET /api/v1/auto-post/settings
Authorization: Bearer <token>
```

#### ดู Setting เฉพาะ
```bash
GET /api/v1/auto-post/settings/{id}
Authorization: Bearer <token>
```

#### แก้ไข Setting
```bash
PUT /api/v1/auto-post/settings/{id}
Authorization: Bearer <token>
{
  "topics": [
    "New topic 1",
    "New topic 2"
  ],
  "isEnabled": true
}
```

#### เปิด/ปิดการทำงาน
```bash
# เปิด
POST /api/v1/auto-post/settings/{id}/enable
Authorization: Bearer <token>

# ปิด
POST /api/v1/auto-post/settings/{id}/disable
Authorization: Bearer <token>
```

#### ลบ Setting
```bash
DELETE /api/v1/auto-post/settings/{id}
Authorization: Bearer <token>
```

### Trigger โพสต์ Manual

```bash
# ใช้ topic แบบสุ่ม
POST /api/v1/auto-post/settings/{id}/trigger
Authorization: Bearer <token>

# กำหนด topic เอง
POST /api/v1/auto-post/settings/{id}/trigger
Authorization: Bearer <token>
{
  "topic": "Custom topic here"
}
```

### ดู Logs

```bash
# ดูทั้งหมด
GET /api/v1/auto-post/logs
Authorization: Bearer <token>

# กรองตาม Setting ID
GET /api/v1/auto-post/logs?settingId={uuid}
Authorization: Bearer <token>

# กรองตาม Status
GET /api/v1/auto-post/logs?status=success
Authorization: Bearer <token>

# ดู log เฉพาะ
GET /api/v1/auto-post/logs/{id}
Authorization: Bearer <token>
```

---

## 📡 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auto-post/settings` | สร้าง setting ใหม่ |
| GET | `/api/v1/auto-post/settings` | ดู settings ทั้งหมด |
| GET | `/api/v1/auto-post/settings/{id}` | ดู setting เฉพาะ |
| PUT | `/api/v1/auto-post/settings/{id}` | แก้ไข setting |
| DELETE | `/api/v1/auto-post/settings/{id}` | ลบ setting |
| POST | `/api/v1/auto-post/settings/{id}/enable` | เปิดการทำงาน |
| POST | `/api/v1/auto-post/settings/{id}/disable` | ปิดการทำงาน |
| POST | `/api/v1/auto-post/settings/{id}/trigger` | สร้างโพสต์ทันที |
| GET | `/api/v1/auto-post/logs` | ดู logs |
| GET | `/api/v1/auto-post/logs/{id}` | ดู log เฉพาะ |

---

## 💡 ตัวอย่างการใช้งาน

### สถานการณ์ที่ 1: สร้างระบบโพสต์เนื้อหาเทคโนโลยีทุกวัน

```bash
# 1. สร้าง bot user (ทำครั้งเดียว)
POST /api/v1/auth/register
{
  "username": "tech_updates_bot",
  "email": "techbot@yoursite.com",
  "password": "secure_pass"
}

# 2. สร้าง auto-post setting
POST /api/v1/auto-post/settings
{
  "botUserId": "{bot-user-id}",
  "isEnabled": true,
  "cronSchedule": "0 9 * * *",  // ทุกวันเวลา 09:00
  "model": "gpt-4o-mini",
  "topics": [
    "Latest trends in web development",
    "Artificial intelligence breakthroughs",
    "Cloud computing best practices",
    "Cybersecurity tips for developers",
    "Mobile app development innovations"
  ],
  "maxTokens": 1500
}
```

### สถานการณ์ที่ 2: สร้างระบบโพสต์ Tips ทุกชั่วโมง

```bash
POST /api/v1/auto-post/settings
{
  "botUserId": "{bot-user-id}",
  "isEnabled": true,
  "cronSchedule": "0 * * * *",  // ทุกชั่วโมง
  "model": "gpt-4o-mini",
  "topics": [
    "Productivity tips for remote workers",
    "Programming best practices",
    "Time management techniques",
    "Career advice for developers"
  ],
  "maxTokens": 1000
}
```

### สถานการณ์ที่ 3: ทดสอบก่อนเปิดใช้งาน

```bash
# 1. สร้าง setting แต่ยังไม่เปิด
POST /api/v1/auto-post/settings
{
  "botUserId": "{bot-user-id}",
  "isEnabled": false,  // ปิดไว้ก่อน
  "cronSchedule": "0 * * * *",
  "topics": ["Test topic"]
}

# 2. ทดสอบ manual trigger
POST /api/v1/auto-post/settings/{id}/trigger

# 3. ตรวจสอบผลลัพธ์
GET /api/v1/auto-post/logs?settingId={id}

# 4. ถ้าพอใจแล้ว เปิดใช้งาน
POST /api/v1/auto-post/settings/{id}/enable
```

---

## 🎯 Cron Schedule Examples

```
"0 * * * *"     = ทุกชั่วโมง (minute 0)
"0 */2 * * *"   = ทุก 2 ชั่วโมง
"0 9 * * *"     = ทุกวันเวลา 09:00
"0 9,17 * * *"  = ทุกวันเวลา 09:00 และ 17:00
"0 9 * * 1-5"   = วันจันทร์-ศุกร์ เวลา 09:00
"*/30 * * * *"  = ทุก 30 นาที
```

---

## 📊 การตรวจสอบ Logs

### ผ่าน API
```bash
GET /api/v1/auto-post/logs?limit=10&offset=0
```

### ผ่าน Database
```sql
-- ดู logs ล่าสุด 10 รายการ
SELECT * FROM auto_post_logs
ORDER BY created_at DESC
LIMIT 10;

-- นับจำนวนโพสต์ที่สำเร็จ
SELECT status, COUNT(*)
FROM auto_post_logs
GROUP BY status;

-- ดู statistics ของ setting
SELECT
  s.id,
  s.total_posts_generated,
  s.last_generated_at,
  COUNT(l.id) as total_logs,
  SUM(CASE WHEN l.status = 'success' THEN 1 ELSE 0 END) as success_count
FROM auto_post_settings s
LEFT JOIN auto_post_logs l ON s.id = l.setting_id
GROUP BY s.id;
```

---

## ⚠️ ข้อควรระวัง

1. **ค่าใช้จ่าย OpenAI**: ทุกครั้งที่ generate จะมีค่าใช้จ่าย ควรเลือก model ที่เหมาะสม
   - `gpt-4o-mini`: ถูกที่สุด, เหมาะสำหรับ production
   - `gpt-4o`: สมดุลระหว่างราคาและคุณภาพ
   - `gpt-4`: คุณภาพสูงสุด แต่แพงที่สุด

2. **Rate Limits**: OpenAI มี rate limits ควร:
   - ไม่ตั้ง cron ถี่เกินไป
   - Handle errors อย่างเหมาะสม

3. **Content Moderation**: AI อาจสร้างเนื้อหาที่ไม่เหมาะสม ควร:
   - ใช้ manual approval สำหรับเนื้อหาที่สำคัญ
   - Review logs เป็นประจำ

4. **Bot User Security**:
   - ใช้ password ที่แข็งแรง
   - จำกัด permissions ของ bot user

---

## 🔧 Troubleshooting

### ปัญหา: Scheduler ไม่ทำงาน
- ตรวจสอบ `isEnabled = true`
- ตรวจสอบ cron expression ถูกต้อง
- ดู logs ใน console: `🤖 Running auto-post processor...`

### ปัญหา: OpenAI API Error
- ตรวจสอบ API key ถูกต้อง
- ตรวจสอบ credit balance ใน OpenAI account
- ดู error message ใน `auto_post_logs.error_message`

### ปัญหา: โพสต์ไม่ถูกสร้าง
- ตรวจสอบ bot user มี permissions
- ตรวจสอบ topics ไม่ว่างเปล่า
- ดู logs ว่ามี error อะไร

---

## 📚 Resources

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Cron Expression Generator](https://crontab.guru/)
- [Go Cron Library](https://github.com/go-co-op/gocron)

---

## 🎉 สรุป

ระบบ AI Auto-Post พร้อมใช้งานแล้ว! คุณสามารถ:
1. ตั้งค่า OpenAI API Key ใน `.env`
2. สร้าง Bot User
3. สร้าง Auto-Post Setting ด้วย topics ที่ต้องการ
4. เปิดใช้งานและรอ scheduler ทำงาน
5. Monitor ผลลัพธ์ผ่าน logs

ระบบจะทำงานอัตโนมัติตาม schedule ที่กำหนด และสร้างโพสต์คุณภาพสูงจาก AI! 🚀
