# 🚀 Quick Start: Import Topics จาก CSV

คู่มือฉบับย่อสำหรับ import หัวข้อ 720 หัวข้อจากไฟล์ `suekk_720_posts.csv` เข้าสู่ระบบ Auto-Post

---

## ⚡ ขั้นตอนการใช้งาน (3 Steps)

### Step 1: เตรียม CSV File

วางไฟล์ `suekk_720_posts.csv` ไว้ใน root directory ของโปรเจค

```
gofiber-backend/
├── suekk_720_posts.csv  ← วางไฟล์ตรงนี้
├── bin/
├── scripts/
└── ...
```

**รูปแบบ CSV ที่ต้องการ:**
```csv
category,topic,tone
platform_issues,ค่า fee แพงเกินไป - ร้านค้าลำบาก,controversial
rider_issues,Rider ได้เงินน้อย แต่เหนื่อยมาก,controversial
restaurant_tips,วิธีเพิ่มยอดขาย 5 เทคนิค,professional
customer_tips,สั่งอาหารให้คุ้ม 10 วิธี,casual
```

---

### Step 2: ตั้งค่า Bot User ID

#### 2.1 สร้าง Bot User (ถ้ายังไม่มี)

```bash
# สร้าง bot user
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ai_bot",
    "email": "aibot@example.com",
    "password": "SecurePass123!"
  }'
```

**บันทึก user ID ที่ได้** จากผลลัพธ์:
```json
{
  "user": {
    "id": "123e4567-e89b-12d3-a456-426614174000",  ← copy UUID นี้
    "username": "ai_bot",
    ...
  }
}
```

#### 2.2 เพิ่ม Bot User ID ลงใน .env

แก้ไขไฟล์ `.env`:

```bash
# เพิ่ม/แก้ไขบรรทัดนี้
AUTO_POST_BOT_USER_ID=123e4567-e89b-12d3-a456-426614174000
```

ตรวจสอบว่ามี OpenAI API Key ด้วย:
```bash
OPENAI_API_KEY=sk-your-openai-api-key-here
OPENAI_MODEL=gpt-4o-mini
```

---

### Step 3: Run Import Script

```bash
# เข้าไปที่ root directory ของโปรเจค
cd gofiber-backend

# Run import script
go run scripts/import_csv_to_db.go suekk_720_posts.csv
```

**ผลลัพธ์ที่คาดหวัง:**
```
📁 Reading CSV file: suekk_720_posts.csv
✅ Found 720 topics
✅ Connected to database
🤖 Bot User ID: 123e4567-e89b-12d3-a456-426614174000
📊 Grouped into 15 settings
✅ Created: platform_issues_controversial_1 (50 topics)
✅ Created: platform_issues_controversial_2 (50 topics)
✅ Created: rider_issues_controversial_1 (50 topics)
...
==========================================================
📊 Import Summary:
  ✅ Success: 15 settings
  ❌ Failed: 0 settings
  📝 Total topics: 720
==========================================================

🎯 Next Steps:
  1. Review settings: SELECT * FROM auto_post_settings;
  2. Test one setting: UPDATE auto_post_settings SET is_enabled = true WHERE id = '...';
  3. Restart server to activate scheduler
```

---

## ✅ ตรวจสอบว่า Import สำเร็จ

### ตรวจสอบจาก Database

```bash
# เข้า PostgreSQL
psql -U postgres -d gofiber_template

# ตรวจสอบจำนวน settings
SELECT COUNT(*) FROM auto_post_settings;
-- ควรได้ประมาณ 10-20 settings

# ตรวจสอบจำนวน topics ทั้งหมด
SELECT
  tone,
  COUNT(*) as settings_count,
  SUM(jsonb_array_length(topics)) as total_topics
FROM auto_post_settings
GROUP BY tone;

# ดูรายละเอียด settings
SELECT
  id,
  tone,
  jsonb_array_length(topics) as topics_count,
  is_enabled,
  created_at
FROM auto_post_settings
ORDER BY tone, created_at;
```

---

## 🎮 การใช้งาน

### 1. Enable Settings (ทดสอบทีละอัน)

```bash
# Login เพื่อรับ JWT token
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-admin@example.com",
    "password": "your-password"
  }'

# Copy JWT token ที่ได้

# Enable setting ที่ต้องการ
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/enable \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 2. ทดสอบ Manual Trigger (แนะนำ!)

```bash
# ทดสอบสร้างโพสต์แบบ manual ก่อน
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{SETTING_ID}/trigger \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json"

# ตรวจสอบผลลัพธ์
# ถ้าสำเร็จ จะได้โพสต์ใหม่สร้างขึ้นมา
```

### 3. Enable Scheduler

Restart server เพื่อเปิดใช้งาน scheduler:

```bash
# ถ้าใช้ Docker
docker-compose restart app

# ถ้า run แบบ direct
# หยุด server (Ctrl+C)
# แล้ว start ใหม่
./bin/api
```

---

## 📊 Monitor การทำงาน

### ตรวจสอบ Logs

```bash
# ดู logs จาก auto_post_logs table
SELECT
  id,
  topic,
  generated_title,
  status,
  tokens_used,
  created_at
FROM auto_post_logs
ORDER BY created_at DESC
LIMIT 20;
```

### ตรวจสอบสถิติ

```bash
# ดูสถิติการ generate
SELECT
  setting.tone,
  setting.total_posts_generated,
  setting.last_generated_at,
  COUNT(log.id) as total_logs,
  COUNT(CASE WHEN log.status = 'success' THEN 1 END) as success_count,
  COUNT(CASE WHEN log.status = 'failed' THEN 1 END) as failed_count
FROM auto_post_settings setting
LEFT JOIN auto_post_logs log ON log.setting_id = setting.id
WHERE setting.is_enabled = true
GROUP BY setting.id, setting.tone
ORDER BY setting.last_generated_at DESC;
```

---

## 🛠️ Enable ทั้งหมดพร้อมกัน (Advanced)

### SQL Script

สร้างไฟล์ `enable_all_settings.sql`:

```sql
-- Enable ทุก settings ที่ import มา
UPDATE auto_post_settings
SET is_enabled = true
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;

-- ตรวจสอบ
SELECT
  id,
  tone,
  jsonb_array_length(topics) as topics,
  is_enabled
FROM auto_post_settings
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;
```

Run script:
```bash
psql -U postgres -d gofiber_template -f enable_all_settings.sql
```

---

## ⚙️ การปรับแต่ง Settings

### แก้ไข Cron Schedule

```sql
-- เปลี่ยนจาก ทุกชั่วโมง เป็น ทุก 2 ชั่วโมง
UPDATE auto_post_settings
SET cron_schedule = '0 */2 * * *'
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;

-- เปลี่ยนเป็น ทุก 30 นาที
UPDATE auto_post_settings
SET cron_schedule = '*/30 * * * *';

-- เฉพาะ Controversial tone ให้โพสต์บ่อยขึ้น
UPDATE auto_post_settings
SET cron_schedule = '*/30 * * * *'
WHERE tone = 'controversial';
```

### เปิดใช้งาน Batch Mode (ประหยัด API calls)

```sql
-- เปิด batch mode: สร้าง 6 โพสต์ต่อครั้ง
UPDATE auto_post_settings
SET
  use_batch_mode = true,
  batch_size = 6
WHERE bot_user_id = 'YOUR_BOT_USER_ID'::uuid;

-- เปลี่ยน schedule เป็น ทุก 6 ชั่วโมง
UPDATE auto_post_settings
SET cron_schedule = '0 */6 * * *'
WHERE use_batch_mode = true;
```

**ประโยชน์ของ Batch Mode:**
- 720 topics ÷ 6 posts/batch = 120 batches
- 120 batches ÷ 30 days = 4 batches/day
- **API calls ลดลง 83%** (จาก 24 calls/day → 4 calls/day)
- **ประหยัดค่าใช้จ่าย ~30-40%**

---

## 🔥 Quick Commands Reference

```bash
# Import topics
go run scripts/import_csv_to_db.go suekk_720_posts.csv

# ตรวจสอบ settings
psql -U postgres -d gofiber_template -c "SELECT COUNT(*) FROM auto_post_settings;"

# Enable setting
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{ID}/enable \
  -H "Authorization: Bearer TOKEN"

# Trigger manual
curl -X POST http://localhost:3000/api/v1/auto-post/settings/{ID}/trigger \
  -H "Authorization: Bearer TOKEN"

# ดู logs ล่าสุด
psql -U postgres -d gofiber_template -c "SELECT topic, status, created_at FROM auto_post_logs ORDER BY created_at DESC LIMIT 10;"
```

---

## 🚨 Troubleshooting

### ❌ Error: "BOT_USER_ID not set in .env"

**แก้ไข:** ตรวจสอบว่าไฟล์ `.env` มี:
```bash
AUTO_POST_BOT_USER_ID=your-uuid-here
```

### ❌ Error: "Failed to connect to database"

**แก้ไข:** ตรวจสอบ database config ใน `.env`:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=gofiber_template
```

### ❌ Error: "Invalid BOT_USER_ID"

**แก้ไข:** ตรวจสอบว่า UUID ถูกต้อง (format: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`)

### ⚠️ CSV encoding issues (ภาษาไทยแสดงผิด)

**แก้ไข:** บันทึก CSV เป็น UTF-8:
- Excel: Save As → CSV UTF-8 (Comma delimited)
- VS Code: Click encoding at bottom → Save with Encoding → UTF-8

### ⚠️ โพสต์ไม่ถูกสร้างอัตโนมัติ

**ตรวจสอบ:**
1. Setting enabled หรือยัง?
   ```sql
   SELECT id, is_enabled FROM auto_post_settings;
   ```

2. Scheduler ทำงานหรือไม่? ดู server logs:
   ```
   [INFO] Running scheduled job: auto-post-processor
   ```

3. มี error ใน logs หรือไม่?
   ```sql
   SELECT * FROM auto_post_logs WHERE status = 'failed' ORDER BY created_at DESC LIMIT 5;
   ```

---

## 📈 Expected Timeline

**24 โพสต์/วัน:**
- 720 topics ÷ 24 posts/day = **30 วัน**

**หมุนเวียน topics:**
- หลังจากใช้ครบ 720 หัวข้อ (30 วัน)
- ระบบจะเริ่มต้นใช้หัวข้อใหม่จากเซ็ตแรกอีกครั้ง
- แต่ title จะแตกต่าง เพราะใช้ GenerateTitleVariations()

---

## 🎉 เสร็จสมบูรณ์!

ตอนนี้คุณมีระบบ Auto-Post ที่:
- ✅ มี 720 หัวข้อพร้อมใช้งาน
- ✅ Auto-post ทุกชั่วโมง (24 โพสต์/วัน)
- ✅ AI สร้างเนื้อหาแตกต่างกันทุกครั้ง
- ✅ ไม่ต้อง manual entry topics
- ✅ ทำงานอัตโนมัติ 100%

**ต่อไปทำอะไร?**
1. Monitor performance ใน 2-3 วันแรก
2. ปรับ cron schedule ตามต้องการ
3. เปิด batch mode เพื่อประหยัดค่าใช้จ่าย
4. เพิ่ม topics ใหม่เมื่อต้องการ

มีคำถามหรือต้องการความช่วยเหลือเพิ่มเติมครับ? 😊
