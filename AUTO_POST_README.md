# 🤖 AI Auto-Post System

ระบบโพสต์อัตโนมัติด้วย AI สำหรับ GoFiber Backend

---

## 🎯 คุณสมบัติหลัก

- ✅ **AI Content Generation** - ใช้ OpenAI GPT สร้างเนื้อหาคุณภาพสูง
- ✅ **Auto Scheduling** - โพสต์อัตโนมัติตาม cron schedule
- ✅ **Topic-based** - รองรับ 720+ หัวข้อ
- ✅ **5 Tone Styles** - neutral, casual, professional, humorous, controversial
- ✅ **Title Variations** - ป้องกันชื่อซ้ำ
- ✅ **Batch Mode** - ประหยัดต้นทุน API 30-40%
- ✅ **Approval Workflow** - ตรวจสอบเนื้อหาก่อนโพสต์

---

## 📚 เอกสารทั้งหมด

### 🚀 เริ่มต้นใช้งาน
- **[AUTO_POST_SETUP_FINAL.md](AUTO_POST_SETUP_FINAL.md)** ⭐ **อ่านตัวนี้ก่อน!**
  - Setup ง่ายที่สุด 3 ขั้นตอน
  - การใช้งานพื้นฐาน
  - Configuration แนะนำ

### 📖 คู่มือหลัก
- **[AI_AUTO_POST_GUIDE.md](AI_AUTO_POST_GUIDE.md)**
  - คำอธิบายระบบแบบละเอียด
  - Flow diagram
  - API endpoints ทั้งหมด
  - Troubleshooting

### 🔧 CSV Import
- **[CSV_IMPORT_QUICKSTART.md](CSV_IMPORT_QUICKSTART.md)**
  - Import topics จาก CSV
  - Step-by-step guide
  - ตรวจสอบและ monitor

### ⚡ Features & Improvements
- **[AI_AUTO_POST_IMPROVEMENTS.md](AI_AUTO_POST_IMPROVEMENTS.md)**
  - Title Variation Generator
  - Batch Mode
  - Approval Workflow
  - Performance comparison
  - Cost analysis

### 📝 Topic Preparation
- **[PREPARE_720_TOPICS_GUIDE.md](PREPARE_720_TOPICS_GUIDE.md)**
  - วิธีเตรียม 720 หัวข้อ
  - 5 วิธีการต่างๆ
  - Best practices
  - Quality check

---

## ⚡ Quick Start (3 Steps)

### 1. สร้าง Bot User

```bash
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ai_bot",
    "email": "aibot@example.com",
    "password": "SecurePass123!"
  }'

# บันทึก UUID ที่ได้
```

### 2. Config .env

```bash
# เพิ่มใน .env
AUTO_POST_BOT_USER_ID=your-bot-user-uuid
OPENAI_API_KEY=sk-your-openai-api-key
OPENAI_MODEL=gpt-4o-mini
```

### 3. Import Topics

```bash
# Windows
scripts\setup_auto_post.bat suekk_720_posts.csv

# Linux/Mac
chmod +x scripts/setup_auto_post.sh
./scripts/setup_auto_post.sh suekk_720_posts.csv
```

### 4. Enable & Start

```sql
-- Enable all settings
psql -U postgres -d gofiber_template -c \
  "UPDATE auto_post_settings SET is_enabled = true WHERE bot_user_id = 'YOUR_UUID'::uuid;"
```

```bash
# Restart server
docker-compose restart app
# หรือ
./bin/api
```

**เสร็จแล้ว!** 🎉 ระบบจะโพสต์อัตโนมัติทุกชั่วโมง

---

## 📂 ไฟล์สำคัญ

### Code
```
pkg/ai/openai_service.go                    # OpenAI integration
domain/models/auto_post_setting.go          # Database models
domain/services/auto_post_service.go        # Service interface
application/serviceimpl/auto_post_service_impl.go  # Service implementation
interfaces/api/handlers/auto_post_handler.go       # API handlers
infrastructure/postgres/auto_post_repository_impl.go  # Repository
```

### Scripts
```
scripts/import_csv_to_db.go        # Import CSV → Database (Go)
scripts/setup_auto_post.bat        # Setup script (Windows)
scripts/setup_auto_post.sh         # Setup script (Linux/Mac)
scripts/convert_topics_to_json.py  # CSV → JSON converter
scripts/import_topics.py           # Import via API
```

### Migrations
```
migrations/018_create_auto_post_tables.sql     # Initial tables
migrations/019_update_auto_post_tables_v2.sql  # Enhanced features
```

---

## 🎮 การใช้งานพื้นฐาน

### View Settings

```bash
curl http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Manual Trigger

```bash
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/trigger \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### View Logs

```bash
curl http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/logs \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Enable/Disable

```bash
# Enable
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/enable \
  -H "Authorization: Bearer YOUR_TOKEN"

# Disable
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/disable \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📊 Monitor

### View Logs (Database)

```sql
-- ดู 10 โพสต์ล่าสุด
SELECT
  topic,
  generated_title,
  status,
  tokens_used,
  created_at
FROM auto_post_logs
ORDER BY created_at DESC
LIMIT 10;
```

### View Statistics

```sql
-- สถิติแยกตาม tone
SELECT
  setting.tone,
  COUNT(log.id) as total_posts,
  COUNT(CASE WHEN log.status = 'success' THEN 1 END) as success,
  SUM(log.tokens_used) as total_tokens
FROM auto_post_settings setting
LEFT JOIN auto_post_logs log ON log.setting_id = setting.id
GROUP BY setting.tone;
```

---

## ⚙️ Configuration

### Cron Schedule Examples

```sql
-- ทุกชั่วโมง (24 โพสต์/วัน)
UPDATE auto_post_settings SET cron_schedule = '0 * * * *';

-- ทุก 30 นาที (48 โพสต์/วัน)
UPDATE auto_post_settings SET cron_schedule = '*/30 * * * *';

-- ทุก 2 ชั่วโมง (12 โพสต์/วัน)
UPDATE auto_post_settings SET cron_schedule = '0 */2 * * *';

-- ช่วงเวลาทำการ 9:00-18:00 (10 โพสต์/วัน)
UPDATE auto_post_settings SET cron_schedule = '0 9-18 * * *';
```

### Batch Mode (แนะนำ!)

```sql
-- สร้าง 6 โพสต์ต่อครั้ง, ทุก 6 ชั่วโมง
UPDATE auto_post_settings
SET
  use_batch_mode = true,
  batch_size = 6,
  cron_schedule = '0 */6 * * *'
WHERE bot_user_id = 'YOUR_UUID'::uuid;

-- ผลลัพธ์: 24 โพสต์/วัน, เพียง 4 API calls (ประหยัด 83%)
```

---

## 💰 Cost Estimation

### Standard Mode (24 posts/day)

```
API Calls:    24 calls/day
Tokens:       ~31,200 tokens/day
Monthly Cost: ~$0.40/month (gpt-4o-mini)
```

### Batch Mode (24 posts/day in 4 batches)

```
API Calls:    4 calls/day
Tokens:       ~24,000 tokens/day
Monthly Cost: ~$0.37/month (gpt-4o-mini)
Save:         ~10%
```

---

## 🚨 Troubleshooting

### โพสต์ไม่ถูกสร้าง

```sql
-- ตรวจสอบ settings
SELECT id, tone, is_enabled FROM auto_post_settings;

-- ตรวจสอบ errors
SELECT topic, status, error_message
FROM auto_post_logs
WHERE status = 'failed'
ORDER BY created_at DESC
LIMIT 5;
```

### Check Scheduler Logs

```bash
# ดู server logs
docker-compose logs -f app | grep "auto-post"

# ควรเห็น:
# [INFO] Running scheduled job: auto-post-processor
```

---

## 🎯 Best Practices

### 1. เริ่มต้นแบบ Safe

```
Day 1-2:  Enable 1 setting, monitor
Day 3-5:  Enable 2-3 settings
Week 2:   Enable all if everything works
```

### 2. ใช้ Batch Mode

```
- ประหยัดต้นทุน 30-40%
- API calls ลดลง 83%
- ยังได้โพสต์เท่าเดิม
```

### 3. เปิด Title Variations

```sql
UPDATE auto_post_settings
SET enable_variations = true;

-- ป้องกันชื่อซ้ำ, ดูน่าสนใจกว่า
```

### 4. Monitor Regularly

```
- ตรวจสอบ logs ทุกวัน
- ดูสถิติ success/failed rate
- ปรับ schedule ตามผล
```

---

## 📞 ต้องการความช่วยเหลือ?

อ่านเอกสารเพิ่มเติม:
1. **[AUTO_POST_SETUP_FINAL.md](AUTO_POST_SETUP_FINAL.md)** - Setup & การใช้งาน
2. **[AI_AUTO_POST_GUIDE.md](AI_AUTO_POST_GUIDE.md)** - คู่มือละเอียด
3. **[CSV_IMPORT_QUICKSTART.md](CSV_IMPORT_QUICKSTART.md)** - CSV Import

---

## 🎉 สรุป

**ระบบพร้อมใช้งาน 100%!**

- ✅ Import CSV → Database
- ✅ AI generate content
- ✅ Auto-post ตาม schedule
- ✅ Monitor & analytics
- ✅ Cost-efficient

**เริ่มใช้งานได้เลย!** 🚀
