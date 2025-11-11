# 🎯 EasyPanel Deployment Guide

คู่มือ deploy GoFiber Backend บน EasyPanel (10-15 นาที)

---

## 🚀 ทำไมต้อง EasyPanel?

**เปรียบเทียบ:**

| Feature | Manual Deploy | EasyPanel |
|---------|--------------|-----------|
| ความยาก | 😰😰😰 ยุ่งยาก | 😊 ง่ายมาก |
| เวลา setup | 1-2 ชั่วโมง | 10-15 นาที |
| GUI Management | ❌ ไม่มี | ✅ มี |
| SSL Certificate | ต้องทำเอง | อัตโนมัติ |
| Monitoring | ต้อง setup | มีในตัว |
| Logs | ต้อง config | มีในตัว |
| Database UI | ต้องติดตั้ง | มีในตัว |
| Backups | ต้องทำเอง | มีในตัว |
| Updates | ต้อง rebuild manual | กดปุ่มเดียว |

**สรุป:** EasyPanel ประหยัดเวลาและลดความซับซ้อนอย่างมาก ✨

---

## 📋 ข้อกำหนดเบื้องต้น

- **VPS/Server:** 2GB RAM, 2 CPU cores, Ubuntu 22.04 (แนะนำ)
- **Domain:** ชี้ DNS A record ไปที่ server IP
- **SSH Access:** สำหรับติดตั้ง EasyPanel

---

## 🎯 Step 1: ติดตั้ง EasyPanel (3 นาที)

### SSH เข้า Server

```bash
ssh root@your-server-ip
```

### Run Installation Script

```bash
curl -sSL https://get.easypanel.io | sh
```

**Installation จะติดตั้ง:**
- Docker & Docker Compose
- EasyPanel Dashboard
- Traefik (reverse proxy)

### เข้า EasyPanel Dashboard

```
URL: http://your-server-ip:3000
```

สร้าง admin account ตามที่ระบบแนะนำ

---

## 🗄️ Step 2: สร้าง Database Services (2 นาที)

### 2.1 สร้าง Project

1. คลิก **"+ New Project"**
2. ตั้งชื่อ: `serkk-backend`

### 2.2 เพิ่ม PostgreSQL

1. ใน project คลิก **"+ Add Service"**
2. เลือก **"PostgreSQL"**
3. ตั้งค่า:
   ```
   Name: postgres
   Version: 15
   Database Name: gofiber_db
   Database User: gofiber_user
   Password: [สร้าง secure password]
   ```
4. คลิก **"Create"**

### 2.3 เพิ่ม Redis

1. คลิก **"+ Add Service"** อีกครั้ง
2. เลือก **"Redis"**
3. ตั้งค่า:
   ```
   Name: redis
   Version: 7
   ```
4. คลิก **"Create"**

### 2.4 เพิ่ม UUID Extension ใน PostgreSQL

1. คลิกที่ PostgreSQL service
2. ไปที่ tab **"Terminal"**
3. Run:
   ```bash
   psql -U gofiber_user -d gofiber_db
   ```
4. ใน psql prompt:
   ```sql
   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
   \q
   ```

---

## 🚀 Step 3: Deploy Backend Application (5 นาที)

### 3.1 เพิ่ม App Service

1. คลิก **"+ Add Service"**
2. เลือก **"App"**

### 3.2 Configure Source

**Option A: Deploy from GitHub (แนะนำ)**

```
Source: GitHub
Repository URL: https://github.com/thizplus/serkk-backend
Branch: main
Build Method: Dockerfile
Dockerfile Path: Dockerfile.easypanel
```

**Option B: Deploy from Docker Hub**

```
Source: Docker Image
Image: your-dockerhub/serkk-backend:latest
```

### 3.3 Configure Port

```
Port: 8080
```

### 3.4 Environment Variables

คลิก **"Environment"** tab และเพิ่ม:

```env
# Server
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0

# Database (ใช้ชื่อ service ที่สร้างไว้)
DB_HOST=postgres
DB_PORT=5432
DB_USER=gofiber_user
DB_PASSWORD=your_secure_password_from_step2
DB_NAME=gofiber_db
DB_SSL_MODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT (Generate new secret: openssl rand -base64 32)
JWT_SECRET=your_super_secure_jwt_secret_min_32_chars
JWT_EXPIRE_HOURS=720

# Frontend
FRONTEND_URL=https://yourdomain.com

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com

# OAuth (Google)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=https://api.yourdomain.com/api/v1/auth/google/callback

# CDN (ถ้าใช้)
BUNNY_STORAGE_API_KEY=your_key
BUNNY_STORAGE_ZONE=your_zone
BUNNY_CDN_URL=https://your-zone.b-cdn.net

# R2 (ถ้าใช้)
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_SECRET_ACCESS_KEY=your_secret_key
R2_BUCKET_NAME=your_bucket
R2_PUBLIC_URL=https://your-r2-url.com
```

### 3.5 Resources (Optional)

```
Memory Limit: 512MB
CPU Limit: 0.5
```

### 3.6 Deploy

1. คลิก **"Create"**
2. รอ build & deploy (2-3 นาที)
3. ดู logs ที่ tab **"Logs"**

---

## 🌐 Step 4: Setup Domain & SSL (2 นาที)

### 4.1 เพิ่ม Domain

1. คลิกที่ App service
2. ไปที่ tab **"Domains"**
3. คลิก **"+ Add Domain"**
4. ใส่: `api.yourdomain.com`
5. เลือก **"Enable SSL"** (Let's Encrypt)
6. คลิก **"Add"**

### 4.2 Update DNS

ไปที่ DNS provider ของคุณ:

```
Type: A
Name: api
Value: your-server-ip
TTL: 3600
```

รอ DNS propagate (5-10 นาที)

### 4.3 Test SSL

```bash
curl https://api.yourdomain.com/health
```

ควรได้ response:
```json
{"status":"ok"}
```

---

## ✅ Step 5: Verify Deployment (2 นาที)

### 5.1 Health Check

```bash
curl https://api.yourdomain.com/health
```

### 5.2 Swagger Docs

เปิดเว็บ:
```
https://api.yourdomain.com/swagger/index.html
```

### 5.3 Test WebSocket

```bash
# Chat WebSocket
wscat -c wss://api.yourdomain.com/ws/chat

# Notifications WebSocket
wscat -c wss://api.yourdomain.com/ws/notifications
```

### 5.4 Check Database

ใน EasyPanel:
1. คลิกที่ PostgreSQL service
2. ไปที่ tab **"Terminal"**
3. Run:
   ```bash
   psql -U gofiber_user -d gofiber_db -c "\dt"
   ```

ควรเห็น 20 tables

---

## 🔧 การจัดการหลัง Deploy

### View Logs

1. คลิกที่ App service
2. ไปที่ tab **"Logs"**
3. ดู real-time logs

### Restart Service

1. คลิกที่ App service
2. คลิก **"⋮"** (3 dots)
3. เลือก **"Restart"**

### Update Application

**ถ้า deploy จาก GitHub:**

1. Push code ใหม่ไป GitHub
2. ใน EasyPanel คลิก **"Rebuild"**
3. รอ build เสร็จ

**ถ้า deploy จาก Docker Image:**

1. Build image ใหม่
2. Push to Docker Hub
3. ใน EasyPanel คลิก **"Rebuild"**

### Scale Resources

1. คลิกที่ App service
2. ไปที่ tab **"Resources"**
3. ปรับ Memory/CPU
4. คลิก **"Update"**

### Database Backup

1. คลิกที่ PostgreSQL service
2. ไปที่ tab **"Backups"**
3. คลิก **"Create Backup"**
4. ตั้ง scheduled backup (ถ้าต้องการ)

### Monitor Resources

1. Dashboard แสดง CPU/Memory/Disk usage
2. ดู metrics ที่ tab **"Metrics"**

---

## 🐛 Troubleshooting

### App ไม่ขึ้น

```bash
# ดู logs
คลิก Logs tab ใน EasyPanel

# ตรวจสอบ environment variables
ดู Environment tab - ตรวจสอบ DB credentials

# Test database connection
ใน PostgreSQL service terminal:
psql -U gofiber_user -d gofiber_db
```

### SSL Certificate Error

```bash
# ตรวจสอบ DNS
dig api.yourdomain.com

# Force renew SSL
ใน Domains tab - คลิก "Renew Certificate"
```

### Database Connection Error

```bash
# ตรวจสอบ postgres service
ใน PostgreSQL service - ดู Logs tab

# Test connection
DB_HOST=postgres  # ต้องเป็นชื่อ service ไม่ใช่ localhost
```

### Out of Memory

```bash
# เพิ่ม memory limit
Resources tab → Memory Limit: 1024MB
```

---

## 🎯 Deploy Frontend (Next.js/React)

### สร้าง Frontend Service

1. ใน project เดียวกัน คลิก **"+ Add Service"**
2. เลือก **"App"**
3. Configure:
   ```
   Source: GitHub
   Repository: your-frontend-repo
   Branch: main
   Build Method: Dockerfile
   Port: 3000
   ```

4. Environment:
   ```
   NEXT_PUBLIC_API_URL=https://api.yourdomain.com
   ```

5. Domain:
   ```
   yourdomain.com
   www.yourdomain.com
   ```

6. Deploy!

---

## 💰 ค่าใช้จ่าย

### VPS (DigitalOcean/Hetzner)

- **Basic:** $6/month (2GB RAM, 1 CPU) - เพียงพอสำหรับ demo
- **Production:** $12/month (4GB RAM, 2 CPU) - แนะนำ
- **High Traffic:** $24/month (8GB RAM, 4 CPU)

### EasyPanel

- **Free!** Open source, self-hosted

### Domain

- **$10-15/year** (Namecheap, Cloudflare)

### SSL Certificate

- **Free** (Let's Encrypt via EasyPanel)

**Total:** ประมาณ $6-12/month + domain

---

## 📚 เปรียบเทียบกับ Manual Deploy

| Task | Manual | EasyPanel | Time Saved |
|------|--------|-----------|------------|
| Install dependencies | 15 min | 0 min | ✅ 15 min |
| Setup database | 10 min | 2 min | ✅ 8 min |
| Configure Nginx | 15 min | 0 min | ✅ 15 min |
| SSL Certificate | 10 min | 1 min | ✅ 9 min |
| Deploy app | 20 min | 5 min | ✅ 15 min |
| Setup monitoring | 30 min | 0 min | ✅ 30 min |
| **Total** | **100 min** | **10 min** | **✅ 90 min** |

---

## 🎉 Conclusion

**EasyPanel ช่วยประหยัดเวลา 90 นาที (90%)** และทำให้:

✅ Deploy ง่ายกว่ามาก
✅ จัดการง่าย (GUI)
✅ Update ง่าย (rebuild)
✅ Monitor ง่าย (built-in)
✅ Backup ง่าย (built-in)
✅ Scale ง่าย (เพิ่ม resources)

**แนะนำมากๆ สำหรับ:**
- Solo developers
- Small teams
- Startups
- Projects ที่ต้องการ deploy เร็ว

---

## 🔗 Links

- **EasyPanel:** https://easypanel.io
- **Docs:** https://easypanel.io/docs
- **Discord:** https://discord.gg/easypanel
- **GitHub:** https://github.com/easypanel-io/easypanel

---

**🚀 Happy Deploying!**

หากมีปัญหา DM ได้เลยครับ 😊
