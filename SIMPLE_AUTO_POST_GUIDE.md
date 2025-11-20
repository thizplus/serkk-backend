# 🎯 Simple Auto-Post - คู่มือง่ายๆ

ระบบโพสต์อัตโนมัติแบบง่ายที่สุด - แค่ 3 ขั้นตอน!

---

## 🚀 วิธีใช้งาน (3 ขั้นตอน)

### ขั้นตอนที่ 1: สมัคร Bot User

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"ai_bot\",
    \"email\": \"aibot@example.com\",
    \"password\": \"SecurePass123!\"
  }"
```

**บันทึก UUID** ที่ได้จาก response:
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",  ← copy UUID นี้
    "username": "ai_bot",
    ...
  }
}
```

---

### ขั้นตอนที่ 2: Config .env

แก้ไขไฟล์ `.env`:

```bash
# Bot User ID (จากขั้นตอนที่ 1)
AUTO_POST_BOT_USER_ID=123e4567-e89b-12d3-a456-426614174000

# OpenAI API Key
OPENAI_API_KEY=sk-your-openai-api-key-here
OPENAI_MODEL=gpt-4o-mini

# Database (ถ้ายังไม่มี)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gofiber_template
```

---

### ขั้นตอนที่ 3: Import CSV

วางไฟล์ `suekk_720_posts.csv` ใน root directory แล้วรัน:

```bash
go run scripts/simple_import_csv.go suekk_720_posts.csv
```

**เสร็จแล้ว!** 🎉

---

## 📊 รูปแบบ CSV

**ตัวอย่าง `suekk_720_posts.csv`:**

```csv
category,topic,tone
platform_issues,ค่า fee แพงเกินไป - ร้านค้าลำบาก,controversial
rider_issues,Rider ได้เงินน้อย แต่เหนื่อยมาก,controversial
restaurant_tips,วิธีเพิ่มยอดขาย 5 เทคนิค,professional
customer_tips,สั่งอาหารให้คุ้ม 10 วิธี,casual
delivery_news,สถิติการสั่งอาหารปี 2024,neutral
funny_content,เมื่อ rider เจอฝนตก,humorous
```

**คอลัมน์:**
- `category` - หมวดหมู่ (ไม่ใช้ในการ gen, แค่เพื่อ organize)
- `topic` - หัวข้อที่ต้องการให้ AI สร้างโพสต์
- `tone` - โทนเสียง: `neutral`, `casual`, `professional`, `humorous`, `controversial`

---

## 🎮 Start Server

```bash
# Build
go build -o bin/api cmd/api/main.go

# Run
./bin/api

# หรือ Docker
docker-compose up -d
```

**ระบบจะทำงานเอง!** Scheduler จะรันทุกชั่วโมง:
- หา topic ถัดไปที่ status = pending
- ใช้ AI สร้างโพสต์
- เปลี่ยน status = completed

---

## 📈 Monitor

### ดู Queue

```sql
-- ดู topics ทั้งหมด
SELECT
  id,
  topic,
  tone,
  status,
  created_at
FROM auto_post_queue
ORDER BY created_at;

-- ดู topics ที่รอทำ
SELECT COUNT(*) FROM auto_post_queue WHERE status = 'pending';

-- ดู topics ที่เสร็จแล้ว
SELECT COUNT(*) FROM auto_post_queue WHERE status = 'completed';

-- ดู topic ล่าสุดที่ทำ
SELECT * FROM auto_post_queue
WHERE status = 'completed'
ORDER BY completed_at DESC
LIMIT 10;
```

### ดูโพสต์ที่สร้าง

```sql
-- ดูโพสต์จาก bot
SELECT
  p.id,
  p.title,
  p.content,
  p.created_at
FROM posts p
WHERE p.user_id = 'YOUR_BOT_USER_ID'::uuid
ORDER BY p.created_at DESC
LIMIT 10;

-- นับจำนวนโพสต์
SELECT COUNT(*) FROM posts
WHERE user_id = 'YOUR_BOT_USER_ID'::uuid;
```

---

## 🔄 ระบบทำงานยังไง?

**Timeline:**

```
Hour 1:  เช็ค queue → หา topic แรกที่ pending → gen โพสต์ → status = completed
Hour 2:  เช็ค queue → หา topic ถัดไปที่ pending → gen โพสต์ → status = completed
Hour 3:  เช็ค queue → หา topic ถัดไปที่ pending → gen โพสต์ → status = completed
...
Hour 720: เช็ค queue → หา topic สุดท้ายที่ pending → gen โพสต์ → status = completed
Hour 721: เช็ค queue → ไม่มี pending topics → ไม่ทำอะไร
```

**ผลลัพธ์:**
- 720 topics → 720 โพสต์
- 1 โพสต์ต่อชั่วโมง
- ใช้เวลา 30 วัน (24 × 30 = 720)

---

## 🛠️ การปรับแต่ง

### เปลี่ยนความถี่

แก้ไขใน `pkg/di/container.go`:

```go
// ทุกชั่วโมง (default)
c.EventScheduler.AddJob("simple-auto-post-processor", "0 * * * *", ...)

// ทุก 30 นาที (โพสต์เร็วขึ้น 2 เท่า)
c.EventScheduler.AddJob("simple-auto-post-processor", "*/30 * * * *", ...)

// ทุก 2 ชั่วโมง (โพสต์ช้าลง 2 เท่า)
c.EventScheduler.AddJob("simple-auto-post-processor", "0 */2 * * *", ...)

// ทุก 15 นาที (โพสต์เร็วขึ้น 4 เท่า)
c.EventScheduler.AddJob("simple-auto-post-processor", "*/15 * * * *", ...)
```

**rebuild และ restart server:**
```bash
go build -o bin/api cmd/api/main.go
./bin/api
```

---

## 🔧 Operations

### เพิ่ม Topics เพิ่ม

```bash
# เตรียม CSV ใหม่
# แล้วรัน import อีกครั้ง
go run scripts/simple_import_csv.go more_topics.csv

# topics ใหม่จะถูกเพิ่มเข้า queue
```

### Reset Queue

```sql
-- เปลี่ยน status กลับเป็น pending ทั้งหมด
UPDATE auto_post_queue SET status = 'pending', completed_at = NULL;

-- เปลี่ยนเฉพาะบาง topics
UPDATE auto_post_queue
SET status = 'pending', completed_at = NULL
WHERE topic LIKE '%keyword%';
```

### ลบ Topics

```sql
-- ลบ topics ที่ไม่ต้องการ
DELETE FROM auto_post_queue WHERE topic LIKE '%spam%';

-- ลบ topics ที่ทำเสร็จแล้ว
DELETE FROM auto_post_queue WHERE status = 'completed';
```

### Manual Trigger (ทดสอบ)

สร้าง test script `test_auto_post.go`:

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gofiber-template/application/serviceimpl"
	"gofiber-template/infrastructure/postgres"
	"gofiber-template/pkg/ai"
	// import other dependencies...
)

func main() {
	godotenv.Load()

	// Connect DB
	db, _ := postgres.NewDatabase(...)

	// Create services
	openAI := ai.NewOpenAIService(os.Getenv("OPENAI_API_KEY"), "gpt-4o-mini")
	// ... create postService
	service := serviceimpl.NewSimpleAutoPostService(db, openAI, postService)

	// Run
	ctx := context.Background()
	if err := service.ProcessNextTopic(ctx); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Success!")
}
```

```bash
go run test_auto_post.go
```

---

## ⚠️ Troubleshooting

### โพสต์ไม่ถูกสร้าง

**ตรวจสอบ:**

1. **มี pending topics หรือไม่?**
   ```sql
   SELECT COUNT(*) FROM auto_post_queue WHERE status = 'pending';
   ```

2. **Scheduler ทำงานหรือไม่?**
   - ดู server logs หา: `"📝 Running simple auto-post processor..."`

3. **มี errors หรือไม่?**
   ```sql
   SELECT * FROM auto_post_queue WHERE status = 'failed' LIMIT 10;
   ```

4. **OpenAI API Key ถูกต้องหรือไม่?**
   ```bash
   # Test
   curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer $OPENAI_API_KEY"
   ```

### CSV Import Failed

**ตรวจสอบ:**

1. **File encoding = UTF-8?**
   - บันทึกไฟล์ใหม่เป็น UTF-8
   - Excel: Save As → CSV UTF-8

2. **Format ถูกต้อง?**
   - ต้องมี header row: `category,topic,tone`
   - ต้องมีอย่างน้อย 2 คอลัมน์ (topic + tone)

3. **BOT_USER_ID ถูกต้อง?**
   ```bash
   # Check .env
   cat .env | grep AUTO_POST_BOT_USER_ID
   ```

### OpenAI Error

**ปัญหาที่พบบ่อย:**

- **Rate limit exceeded** → ลดความถี่ (ใช้ cron ห่างกว่า 1 ชั่วโมง)
- **Invalid API key** → ตรวจสอบ .env
- **Insufficient credits** → เติมเงินที่ platform.openai.com

---

## 💰 ประมาณการค่าใช้จ่าย

**gpt-4o-mini:**
- Input: $0.15 / 1M tokens
- Output: $0.60 / 1M tokens

**1 โพสต์:**
- ~500 input tokens (~$0.000075)
- ~800 output tokens (~$0.00048)
- รวม ~$0.00055 / โพสต์

**720 โพสต์:**
- ~$0.396 / 720 โพสต์
- หรือ ~$0.40 / เดือน

**ถูกมาก!** 🎉

---

## 📊 สรุป

| ขั้นตอน | สิ่งที่ทำ |
|---------|----------|
| 1. สมัคร user | สร้าง bot user 1 ครั้ง |
| 2. Config .env | ใส่ user id + API key |
| 3. Import CSV | รัน script 1 ครั้ง |
| 4. Start server | ระบบทำงานเองทุกชั่วโมง |

**เท่านี้แหละ!** ไม่ต้องจัดการอะไรเพิ่ม 😊

---

## 🎉 ข้อดี

- ✅ **ง่ายมาก** - แค่ 3 ขั้นตอน
- ✅ **ไม่ซับซ้อน** - ไม่มี settings, batch mode, variations
- ✅ **เข้าใจง่าย** - แค่ queue ธรรมดา
- ✅ **Monitor ง่าย** - ดู table เดียว
- ✅ **ถูก** - ~$0.40 / เดือน
- ✅ **ไม่ต้องดูแล** - ทำงานเอง 100%

---

## 📞 ต้องการความช่วยเหลือ?

- **ดู logs:** ดูว่า scheduler รันหรือไม่
- **ดู queue:** ตรวจสอบ auto_post_queue table
- **ดูโพสต์:** ตรวจสอบ posts table

มีปัญหาติดต่อได้เลยครับ! 😊
