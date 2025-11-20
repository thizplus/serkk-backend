# 📝 คู่มือเตรียม 720 หัวข้อสำหรับ AI Auto-Post

## 🎯 Overview

720 หัวข้อ = 24 โพสต์/วัน × 30 วัน = เพียงพอสำหรับ 1 เดือน

---

## 🚀 วิธีที่ 1: Import ผ่าน CSV + Python Script (แนะนำ!)

### ขั้นตอนที่ 1: สร้างไฟล์ Topics

**ตัวอย่าง topics.csv:**
```csv
category,topic,tone
platform_issues,ค่า fee แพงเกินไป - ร้านค้าลำบาก,controversial
platform_issues,Delivery ช้า ลูกค้ารอนาน,controversial
platform_issues,App crash บ่อย ใช้งานลำบาก,casual
rider_issues,Rider ได้เงินน้อย แต่เหนื่อยมาก,controversial
rider_issues,ไม่มี insurance ไม่มีสวัสดิการ,professional
restaurant_tips,วิธีเพิ่มยอดขาย 5 เทคนิค,professional
restaurant_tips,ทำ menu ให้ขายดี,casual
customer_tips,สั่งอาหารให้คุ้ม 10 วิธี,casual
```

### ขั้นตอนที่ 2: Generate Sample Topics (720 หัวข้อ)

```bash
cd scripts

# Generate sample CSV with 720 topics
python convert_topics_to_json.py --generate-sample

# Output: sample_topics.csv (720 rows)
```

### ขั้นตอนที่ 3: แก้ไข Topics ให้ตรงกับธุรกิจ

```bash
# เปิดไฟล์ด้วย Excel หรือ Text Editor
# แก้ไข topics ให้เป็นเนื้อหาที่ต้องการ
```

**💡 Tips:**
- แบ่งเป็น categories (platform_issues, tips, news, etc.)
- ใช้ tone ที่หลากหลาย (ไม่ controversial ทั้งหมด)
- เพิ่ม variations เพื่อความหลากหลาย

### ขั้นตอนที่ 4: Convert CSV → JSON

```bash
python convert_topics_to_json.py sample_topics.csv

# Output: topics.json
```

**ตัวอย่าง topics.json:**
```json
{
  "total_topics": 720,
  "total_settings": 15,
  "settings": [
    {
      "name": "platform_issues_controversial_1",
      "category": "platform_issues",
      "tone": "controversial",
      "topics": [
        "ค่า fee แพงเกินไป - ร้านค้าลำบาก",
        "Delivery ช้า ลูกค้ารอนาน",
        ...
      ],
      "topics_count": 50
    },
    ...
  ]
}
```

### ขั้นตอนที่ 5: Import เข้าระบบ

```bash
# 1. สร้าง Bot User ก่อน (ถ้ายังไม่มี)
curl -X POST http://localhost:3000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "ai_bot",
    "email": "aibot@example.com",
    "password": "SecurePass123!"
  }'

# บันทึก user ID ที่ได้

# 2. Get JWT Token
curl -X POST http://localhost:3000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-admin@example.com",
    "password": "your-password"
  }'

# บันทึก JWT token

# 3. Import topics
python import_topics.py topics.json \
  --api-url http://localhost:3000 \
  --token YOUR_JWT_TOKEN \
  --bot-user-id BOT_USER_UUID

# 4. (Optional) Enable ทั้งหมดเลย
python import_topics.py topics.json \
  --api-url http://localhost:3000 \
  --token YOUR_JWT_TOKEN \
  --bot-user-id BOT_USER_UUID \
  --enable-all
```

---

## 🔧 วิธีที่ 2: Manual Create ผ่าน API (เหมาะกับจำนวนน้อย)

### สร้างทีละ Setting

```bash
# Setting 1: Controversial Topics (50 topics)
curl -X POST http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "botUserId": "BOT_USER_UUID",
    "isEnabled": false,
    "cronSchedule": "0 * * * *",
    "model": "gpt-4o-mini",
    "tone": "controversial",
    "topics": [
      "ค่า fee แพงเกินไป",
      "Delivery ช้า",
      "App crash บ่อย",
      ... (50 topics total)
    ],
    "maxTokens": 1500,
    "enableVariations": true
  }'

# Setting 2: Casual Topics (50 topics)
# Setting 3: Professional Topics (50 topics)
# ...
# ทำต่อจนครบ 720 topics
```

---

## 💾 วิธีที่ 3: Direct Database Insert (Advanced)

### สร้าง SQL Script

```sql
-- Insert Bot User (ถ้ายังไม่มี)
INSERT INTO users (id, username, email, password_hash)
VALUES (
  gen_random_uuid(),
  'ai_bot',
  'aibot@example.com',
  '$2a$10$...'  -- hashed password
);

-- Get bot_user_id
-- SET bot_user_id = 'uuid-here';

-- Insert Settings
INSERT INTO auto_post_settings (
  id, bot_user_id, is_enabled, cron_schedule, model, tone,
  topics, max_tokens, enable_variations, created_at, updated_at
) VALUES
(
  gen_random_uuid(),
  'bot-user-uuid-here',
  false,
  '0 * * * *',
  'gpt-4o-mini',
  'controversial',
  '["topic1", "topic2", "topic3", ...]'::jsonb,
  1500,
  true,
  NOW(),
  NOW()
),
-- Repeat for all settings...
;
```

### Run SQL Script

```bash
psql -U postgres -d gofiber_template -f insert_topics.sql
```

---

## 🎨 วิธีที่ 4: ใช้ ChatGPT/Claude Generate Topics

### Prompt Template

```
Generate 720 unique topics for a food delivery platform social media posts.

Requirements:
- Topics in Thai language
- Covering these categories:
  * Platform issues (ค่า fee, delivery, app)
  * Rider issues (เงินเดือน, สวัสดิการ)
  * Restaurant tips (เพิ่มยอดขาย, marketing)
  * Customer tips (สั่งอาหาร, ประหยัดเงิน)
  * Industry news (สถิติ, แนวโน้ม)

- Mix of tones:
  * 30% Controversial (provocative, challenging)
  * 25% Casual (friendly, conversational)
  * 20% Professional (informative, data-driven)
  * 15% Humorous (funny, entertaining)
  * 10% Neutral (balanced)

Format as CSV:
category,topic,tone
platform_issues,topic text here,controversial
...

Generate all 720 rows.
```

### Copy & Paste ลงไฟล์

```bash
# บันทึกเป็น ai_generated_topics.csv
# แล้วใช้ script ประมวลผลต่อ
python convert_topics_to_json.py ai_generated_topics.csv
```

---

## 📊 วิธีที่ 5: Hybrid - Multiple Settings Strategy

### แนวคิด: แบ่งตาม Time Slots

```bash
# Setting 1: Morning (6:00-12:00) - Professional tone
- Topics: Industry news, tips, statistics
- 6 hours × 30 days = 180 topics

# Setting 2: Afternoon (12:00-18:00) - Casual tone
- Topics: Customer tips, light content
- 6 hours × 30 days = 180 topics

# Setting 3: Evening (18:00-22:00) - Controversial tone
- Topics: Hot topics, debates
- 4 hours × 30 days = 120 topics

# Setting 4: Night (22:00-06:00) - Humorous tone
- Topics: Fun, entertaining content
- 8 hours × 30 days = 240 topics

Total: 720 topics
```

### สร้าง Settings

```python
# Python script to create time-based settings
import requests

settings_config = [
    {
        'name': 'Morning Professional',
        'cron': '0 6-11 * * *',  # 6:00-11:00
        'tone': 'professional',
        'topics': professional_topics[:180]
    },
    {
        'name': 'Afternoon Casual',
        'cron': '0 12-17 * * *',  # 12:00-17:00
        'tone': 'casual',
        'topics': casual_topics[:180]
    },
    {
        'name': 'Evening Controversial',
        'cron': '0 18-21 * * *',  # 18:00-21:00
        'tone': 'controversial',
        'topics': controversial_topics[:120]
    },
    {
        'name': 'Night Humorous',
        'cron': '0 22-23,0-5 * * *',  # 22:00-05:00
        'tone': 'humorous',
        'topics': humorous_topics[:240]
    }
]

for config in settings_config:
    create_setting(config)
```

---

## 🎯 Best Practices

### 1. แบ่ง Topics ตาม Categories

```
📁 Categories (ควรมี 5-10 categories):
├── platform_issues (150 topics)
├── rider_welfare (100 topics)
├── restaurant_tips (150 topics)
├── customer_guide (100 topics)
├── industry_news (100 topics)
├── success_stories (60 topics)
├── trending_topics (60 topics)
└── seasonal (80 topics)
```

### 2. กระจาย Tones

```
🎨 Tone Distribution:
- Controversial: 30% (216 topics)
- Casual: 25% (180 topics)
- Professional: 20% (144 topics)
- Humorous: 15% (108 topics)
- Neutral: 10% (72 topics)
```

### 3. เพิ่ม Variations

```python
# Base topics: 144 unique topics
# Variations: 5 per topic
# Total: 144 × 5 = 720 topics

base_topics = [
    "ค่า fee แพงเกินไป",
    "Delivery ช้า",
    ...
]

variations = [
    "{topic}",
    "{topic} - ประสบการณ์จริง",
    "{topic} - วิธีแก้ปัญหา",
    "{topic} - มุมมองใหม่",
    "{topic} - สิ่งที่ควรรู้",
]

all_topics = []
for topic in base_topics:
    for var in variations:
        all_topics.append(var.format(topic=topic))
```

### 4. ตรวจสอบ Duplicates

```python
# Check for duplicates
def check_duplicates(topics):
    seen = set()
    duplicates = []

    for topic in topics:
        normalized = topic.lower().strip()
        if normalized in seen:
            duplicates.append(topic)
        seen.add(normalized)

    if duplicates:
        print(f"⚠️  Found {len(duplicates)} duplicates:")
        for dup in duplicates[:10]:
            print(f"  - {dup}")
    else:
        print("✅ No duplicates found!")

    return len(duplicates) == 0
```

---

## 🔍 Quality Check Checklist

ก่อน import ให้ check:

- [ ] ✅ Topics ครบ 720 หัวข้อ
- [ ] ✅ แบ่ง categories สมดุล
- [ ] ✅ Mix tones หลากหลาย
- [ ] ✅ ไม่มี duplicates
- [ ] ✅ ความยาวเหมาะสม (10-100 chars)
- [ ] ✅ ภาษาไทยถูกต้อง encoding (UTF-8)
- [ ] ✅ ไม่มี sensitive content ที่ไม่เหมาะสม
- [ ] ✅ Topics เกี่ยวข้องกับธุรกิจ
- [ ] ✅ ครอบคลุม target audience ทุกกลุ่ม

---

## 📈 Post-Import Management

### ตรวจสอบหลัง Import

```bash
# 1. Count total topics
curl http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer TOKEN" \
  | jq '[.settings[].topics | length] | add'

# ควรได้ 720

# 2. Check by tone
curl http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer TOKEN" \
  | jq '.settings | group_by(.tone) | map({tone: .[0].tone, count: ([.[].topics | length] | add)})'

# 3. List all settings
curl http://localhost:3000/api/v1/auto-post/settings \
  -H "Authorization: Bearer TOKEN" \
  | jq '.settings[] | {id, tone, topics_count: (.topics | length), enabled: .isEnabled}'
```

### Enable Strategy

```bash
# แนะนำ: Enable ทีละ setting และ monitor

# Day 1-3: Enable 1 setting (test)
curl -X POST http://localhost:3000/api/v1/auto-post/settings/SETTING_1/enable

# Day 4-7: Enable 2-3 more settings
curl -X POST http://localhost:3000/api/v1/auto-post/settings/SETTING_2/enable

# Week 2: Enable all if everything works fine
```

---

## 💡 Pro Tips

### 1. Use Batch Mode for Efficiency

```json
{
  "useBatchMode": true,
  "batchSize": 6,
  "cronSchedule": "0 */6 * * *"
}
```

**Benefits:**
- 720 topics ÷ 6 posts/batch = 120 batches
- 120 batches ÷ 30 days = 4 batches/day
- 4 API calls/day (instead of 24)
- **Save ~$0.05/month**

### 2. Seasonal Topics Rotation

```python
# เตรียม topics พิเศษสำหรับ events
seasonal_topics = {
    'new_year': [...],  # ใช้ช่วงปีใหม่
    'songkran': [...],  # ใช้ช่วงสงกรานต์
    'loy_krathong': [...],
    'christmas': [...],
}

# Switch topics based on date
```

### 3. A/B Testing

```python
# สร้าง 2 settings สำหรับ topic เดียวกันแต่ tone ต่างกัน
# Monitor which performs better

Setting A: "ค่า fee แพง" (controversial tone)
Setting B: "ค่า fee แพง" (professional tone)

# หลัง 1 สัปดาห์ ดู engagement แล้วเลือกที่ดีกว่า
```

### 4. Topic Recycling

```sql
-- After 30 days, recycle topics with different variations
-- ใช้ topics เดิม แต่ generate title variations ใหม่
```

---

## 🎉 สรุป

**แนะนำ: วิธีที่ 1 (CSV + Python Script)**

เพราะ:
- ✅ ง่าย แก้ไขได้ใน Excel
- ✅ มี script ช่วย automate
- ✅ ตรวจสอบได้ง่าย
- ✅ Scale ได้ดี (รองรับ 10,000+ topics)

**ขั้นตอนโดยสรุป:**
1. Generate sample: `python convert_topics_to_json.py --generate-sample`
2. แก้ไข topics ใน Excel
3. Convert to JSON: `python convert_topics_to_json.py topics.csv`
4. Import: `python import_topics.py topics.json --token TOKEN --bot-user-id UUID`
5. Test & Enable!

---

มีคำถามหรือต้องการความช่วยเหลือเพิ่มเติมไหมครับ? 😊
