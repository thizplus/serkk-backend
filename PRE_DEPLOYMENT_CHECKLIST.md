# ✅ Pre-Deployment Checklist

เช็คลิสต์ตรวจสอบก่อน deploy ระบบ Auto-Post

---

## 📋 ก่อน Import Topics

### 1. ไฟล์และ Scripts

- [ ] มีไฟล์ `suekk_720_posts.csv` ใน root directory
- [ ] CSV encoding เป็น UTF-8
- [ ] CSV format ถูกต้อง (category,topic,tone)
- [ ] มีไฟล์ `scripts/import_csv_to_db.go`
- [ ] มีไฟล์ `scripts/setup_auto_post.bat` (Windows)
- [ ] มีไฟล์ `scripts/setup_auto_post.sh` (Linux/Mac)

### 2. Database

- [ ] PostgreSQL ทำงานอยู่
- [ ] Database `gofiber_template` ถูกสร้างแล้ว
- [ ] Migration 018 ถูก run แล้ว (`auto_post_settings` table exists)
- [ ] Migration 019 ถูก run แล้ว (enhanced columns exists)

ตรวจสอบ:
```sql
-- ตรวจสอบ tables
SELECT tablename FROM pg_tables
WHERE tablename IN ('auto_post_settings', 'auto_post_logs');

-- ตรวจสอบ columns ใหม่ของ migration 019
SELECT column_name FROM information_schema.columns
WHERE table_name = 'auto_post_settings'
  AND column_name IN ('tone', 'enable_variations', 'use_batch_mode', 'batch_size');
```

### 3. Configuration (.env)

- [ ] มีไฟล์ `.env` ใน root directory
- [ ] `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` ถูกต้อง
- [ ] `OPENAI_API_KEY` มีค่าและถูกต้อง (เริ่มต้นด้วย `sk-`)
- [ ] `OPENAI_MODEL` = `gpt-4o-mini` (หรือ model อื่นที่ต้องการ)
- [ ] `AUTO_POST_BOT_USER_ID` มีค่าและเป็น valid UUID

ตรวจสอบ:
```bash
# Windows
type .env | findstr OPENAI
type .env | findstr AUTO_POST_BOT_USER_ID

# Linux/Mac
grep OPENAI .env
grep AUTO_POST_BOT_USER_ID .env
```

### 4. Bot User

- [ ] Bot user ถูกสร้างแล้วใน database
- [ ] Bot user ID ตรงกับ `AUTO_POST_BOT_USER_ID` ใน .env

ตรวจสอบ:
```sql
-- ตรวจสอบ bot user exists
SELECT id, username, email, created_at
FROM users
WHERE id = 'YOUR_BOT_USER_ID'::uuid;

-- ถ้าไม่มี ให้สร้างผ่าน API หรือ SQL
```

---

## 📋 หลัง Import Topics

### 5. Verify Import

- [ ] Script รันสำเร็จ (exit code 0)
- [ ] ไม่มี error messages
- [ ] Settings ถูกสร้างใน database

ตรวจสอบ:
```sql
-- ตรวจสอบจำนวน settings
SELECT COUNT(*) as total_settings FROM auto_post_settings;
-- ควรได้ 10-20 settings

-- ตรวจสอบจำนวน topics ทั้งหมด
SELECT
  SUM(jsonb_array_length(topics)) as total_topics
FROM auto_post_settings;
-- ควรได้ 720 topics

-- ตรวจสอบแยกตาม tone
SELECT
  tone,
  COUNT(*) as settings_count,
  SUM(jsonb_array_length(topics)) as topics_count
FROM auto_post_settings
GROUP BY tone
ORDER BY tone;
```

### 6. Settings Verification

- [ ] ทุก settings มี `bot_user_id` ตรงกับที่ config ไว้
- [ ] ทุก settings มี topics (jsonb array)
- [ ] `is_enabled` = false (ยังไม่ enable)
- [ ] `cron_schedule` มีค่า (default: `0 * * * *`)

ตรวจสอบ:
```sql
-- ตรวจสอบ settings details
SELECT
  id,
  bot_user_id,
  tone,
  jsonb_array_length(topics) as topics_count,
  is_enabled,
  cron_schedule,
  model,
  use_batch_mode,
  batch_size
FROM auto_post_settings
ORDER BY tone;

-- ตรวจสอบว่าทุก setting มี topics
SELECT id, tone
FROM auto_post_settings
WHERE topics IS NULL OR jsonb_array_length(topics) = 0;
-- ควรได้ 0 rows
```

---

## 📋 ก่อน Enable Auto-Post

### 7. OpenAI API Test

- [ ] OpenAI API key ใช้งานได้
- [ ] มี credits เพียงพอ
- [ ] Test generate content สำเร็จ

ตรวจสอบ:
```bash
# Test OpenAI API
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# ถ้าสำเร็จ จะได้ list of models
```

### 8. Manual Trigger Test

- [ ] Login และรับ JWT token สำเร็จ
- [ ] Manual trigger 1 โพสต์สำเร็จ
- [ ] โพสต์ถูกสร้างใน `posts` table
- [ ] Log ถูกสร้างใน `auto_post_logs` table

ตรวจสอบ:
```bash
# 1. Login
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-admin@example.com",
    "password": "your-password"
  }'

# 2. Get settings
curl http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 3. Manual trigger (เลือก setting_id จากขั้นตอนที่ 2)
curl -X POST http://localhost:3000/api/v1/auto-post/settings/SETTING_ID/trigger \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# 4. Check result
curl http://localhost:3000/api/v1/auto-post/settings/SETTING_ID/logs \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

ตรวจสอบ database:
```sql
-- ตรวจสอบว่า log ถูกสร้าง
SELECT * FROM auto_post_logs
ORDER BY created_at DESC
LIMIT 1;

-- ตรวจสอบว่า post ถูกสร้าง
SELECT * FROM posts
WHERE user_id = 'BOT_USER_ID'::uuid
ORDER BY created_at DESC
LIMIT 1;
```

### 9. Scheduler Test

- [ ] Server start สำเร็จ
- [ ] Scheduler ถูก initialized
- [ ] Auto-post processor job registered

ตรวจสอบ server logs:
```
[INFO] Starting cron scheduler...
[INFO] Registered job: auto-post-processor (schedule: 0 * * * *)
[INFO] Cron scheduler started successfully
```

---

## 📋 Production Readiness

### 10. Configuration Review

- [ ] Cron schedule เหมาะสม (ไม่บ่อยเกินไป)
- [ ] Batch mode ถูกพิจารณาแล้ว
- [ ] Title variations enabled (ถ้าต้องการ)
- [ ] Max tokens เหมาะสม (default: 1500)
- [ ] Temperature เหมาะสม (default: 0.8)

### 11. Monitoring Setup

- [ ] มีวิธีดู server logs
- [ ] มีวิธีเข้าถึง database
- [ ] ตั้ง alerts สำหรับ errors (ถ้ามี)
- [ ] มีวิธีติดตาม API costs

### 12. Backup & Recovery

- [ ] มี database backup
- [ ] มี .env backup (ไม่ commit ลง git!)
- [ ] มี CSV file backup
- [ ] รู้วิธี disable auto-post ฉุกเฉิน

Disable ฉุกเฉิน:
```sql
-- Disable ทั้งหมดทันที
UPDATE auto_post_settings SET is_enabled = false;
```

---

## 📋 Final Checks

### 13. Documentation

- [ ] อ่าน `AUTO_POST_README.md` แล้ว
- [ ] อ่าน `AUTO_POST_SETUP_FINAL.md` แล้ว
- [ ] เข้าใจวิธี enable/disable settings
- [ ] เข้าใจวิธี monitor logs
- [ ] เข้าใจวิธี troubleshooting

### 14. Security

- [ ] OpenAI API key ไม่ถูก commit ลง git
- [ ] .env file ไม่ถูก commit ลง git
- [ ] Bot user มี password ที่ปลอดภัย
- [ ] JWT secret key ปลอดภัย (ถ้าใช้ production)

### 15. Cost Management

- [ ] เข้าใจ pricing ของ OpenAI
- [ ] ตั้ง usage limits ที่ OpenAI dashboard (ถ้าต้องการ)
- [ ] Monitor costs ได้
- [ ] มี plan สำรองถ้า costs เกิน

Expected costs:
```
Standard mode:  ~$0.40/month (24 posts/day)
Batch mode:     ~$0.37/month (24 posts/day, 4 API calls)
```

---

## 🚀 Ready to Deploy!

เมื่อเช็คทุกข้อแล้ว:

### Step 1: Enable Settings (Start Small)

```sql
-- Enable 1 setting ก่อน (test)
UPDATE auto_post_settings
SET is_enabled = true
WHERE id = 'ONE_SETTING_ID'::uuid;
```

### Step 2: Restart Server

```bash
# Docker
docker-compose restart app

# Direct
./bin/api
```

### Step 3: Monitor (24-48 hours)

```sql
-- ดู logs
SELECT
  topic,
  generated_title,
  status,
  tokens_used,
  created_at
FROM auto_post_logs
ORDER BY created_at DESC
LIMIT 20;

-- ดูสถิติ
SELECT
  COUNT(*) as total,
  COUNT(CASE WHEN status = 'success' THEN 1 END) as success,
  COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed
FROM auto_post_logs
WHERE created_at > NOW() - INTERVAL '24 hours';
```

### Step 4: Enable More (If Test Successful)

```sql
-- Enable all settings
UPDATE auto_post_settings
SET is_enabled = true
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;
```

### Step 5: Optimize (Optional)

```sql
-- Switch to batch mode for cost savings
UPDATE auto_post_settings
SET
  use_batch_mode = true,
  batch_size = 6,
  cron_schedule = '0 */6 * * *'
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;
```

---

## 📊 Success Criteria

ระบบใช้งานได้ดีถ้า:

- ✅ โพสต์ถูกสร้างตาม schedule
- ✅ Success rate > 95%
- ✅ API costs ตามที่คาดการณ์
- ✅ Content quality ดี
- ✅ ไม่มี error ซ้ำๆ
- ✅ Title ไม่ซ้ำกัน (ถ้าเปิด variations)

Monitor metrics:
```sql
-- Success rate
SELECT
  COUNT(*) as total,
  ROUND(COUNT(CASE WHEN status = 'success' THEN 1 END) * 100.0 / COUNT(*), 2) as success_rate
FROM auto_post_logs;

-- Average tokens per post
SELECT
  AVG(tokens_used) as avg_tokens,
  MAX(tokens_used) as max_tokens,
  MIN(tokens_used) as min_tokens
FROM auto_post_logs
WHERE status = 'success';

-- Posts per day
SELECT
  DATE(created_at) as date,
  COUNT(*) as posts_count
FROM auto_post_logs
WHERE status = 'success'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

## 🎉 เสร็จสมบูรณ์!

ระบบพร้อม deploy แล้ว เมื่อทุกข้อ checklist ผ่าน! 🚀

**Good luck!** 😊
