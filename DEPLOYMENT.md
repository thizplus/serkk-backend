# 🚀 GoFiber Backend Deployment Guide

คู่มือการ deploy GoFiber backend ขึ้น production server แบบละเอียด

---

## 📋 สารบัญ

1. [ข้อกำหนดเบื้องต้น](#ข้อกำหนดเบื้องต้น)
2. [เตรียม Server](#เตรียม-server)
3. [ติดตั้ง Dependencies](#ติดตั้ง-dependencies)
4. [Setup Database](#setup-database)
5. [Deploy Application](#deploy-application)
6. [Setup Systemd Service](#setup-systemd-service)
7. [Setup Nginx Reverse Proxy](#setup-nginx-reverse-proxy)
8. [SSL Certificate (HTTPS)](#ssl-certificate-https)
9. [Environment Variables](#environment-variables)
10. [Database Migration](#database-migration)
11. [Monitoring & Logging](#monitoring--logging)
12. [Backup Strategy](#backup-strategy)
13. [CI/CD (Optional)](#cicd-optional)

---

## ข้อกำหนดเบื้องต้น

### Server Requirements
- **OS**: Ubuntu 22.04 LTS (แนะนำ)
- **CPU**: 2 cores ขึ้นไป
- **RAM**: 2GB ขึ้นไป (4GB แนะนำ)
- **Storage**: 20GB ขึ้นไป
- **Bandwidth**: ตามการใช้งาน

### Services Required
- Go 1.24+ (สำหรับ build application)
- PostgreSQL 15+
- Redis 7+
- Nginx (สำหรับ reverse proxy)

### Domain & DNS
- Domain name (เช่น `api.yourdomain.com`)
- DNS A record ชี้ไปที่ server IP

---

## เตรียม Server

### 1. เชื่อมต่อ Server

```bash
ssh root@your-server-ip
```

### 2. Update System

```bash
apt update && apt upgrade -y
apt install -y build-essential curl wget git vim
```

### 3. สร้าง User สำหรับ Application

```bash
# สร้าง user ใหม่
adduser gofiber

# เพิ่ม sudo privileges
usermod -aG sudo gofiber

# Switch to new user
su - gofiber
```

### 4. Setup Firewall

```bash
# อนุญาต SSH, HTTP, HTTPS
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

---

## ติดตั้ง Dependencies

### 1. ติดตั้ง Go

```bash
# Download Go 1.24
cd /tmp
wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz

# Extract และติดตั้ง
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz

# เพิ่ม Go ใน PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# ตรวจสอบ
go version
# ควรแสดง: go version go1.24.3 linux/amd64
```

### 2. ติดตั้ง PostgreSQL 15

```bash
# เพิ่ม PostgreSQL repository
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo tee /etc/apt/trusted.gpg.d/pgdg.asc &>/dev/null

# ติดตั้ง PostgreSQL
sudo apt update
sudo apt install -y postgresql-15 postgresql-contrib-15

# Start และ enable service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# ตรวจสอบสถานะ
sudo systemctl status postgresql
```

### 3. ติดตั้ง Redis

```bash
# ติดตั้ง Redis
sudo apt install -y redis-server

# แก้ไข config ให้ run เป็น systemd service
sudo sed -i 's/supervised no/supervised systemd/g' /etc/redis/redis.conf

# Restart และ enable
sudo systemctl restart redis
sudo systemctl enable redis

# ตรวจสอบ
redis-cli ping
# ควรแสดง: PONG
```

---

## Setup Database

### 1. สร้าง PostgreSQL Database และ User

```bash
# Switch to postgres user
sudo -u postgres psql

# ใน PostgreSQL prompt:
CREATE DATABASE gofiber_db;
CREATE USER gofiber_user WITH ENCRYPTED PASSWORD 'your_secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE gofiber_db TO gofiber_user;

# Enable UUID extension
\c gofiber_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

# ออกจาก psql
\q
```

### 2. Configure PostgreSQL สำหรับ Remote Access (ถ้าต้องการ)

```bash
# แก้ไข postgresql.conf
sudo vim /etc/postgresql/15/main/postgresql.conf

# แก้ไขบรรทัดนี้:
listen_addresses = 'localhost'  # หรือ '*' ถ้าต้องการเปิด remote access

# แก้ไข pg_hba.conf
sudo vim /etc/postgresql/15/main/pg_hba.conf

# เพิ่มบรรทัดนี้ (ถ้าต้องการเปิด remote):
# host    all             all             0.0.0.0/0               md5

# Restart PostgreSQL
sudo systemctl restart postgresql
```

### 3. Test Database Connection

```bash
psql -h localhost -U gofiber_user -d gofiber_db
# ใส่ password ที่สร้างไว้
# ถ้าเข้าได้แสดงว่าสำเร็จ
```

---

## Deploy Application

### 1. Clone Repository

```bash
# ไปที่ home directory
cd ~

# Clone repo
git clone https://github.com/thizplus/serkk-backend.git
cd serkk-backend

# Checkout production branch (ถ้ามี)
# git checkout production
```

### 2. สร้าง Environment File

```bash
# Copy example env
cp .env.example .env

# แก้ไข .env
vim .env
```

**ตัวอย่าง `.env` สำหรับ Production:**

```env
# Server Configuration
APP_ENV=production
APP_PORT=8080
APP_HOST=0.0.0.0

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=gofiber_user
DB_PASSWORD=your_secure_password_here
DB_NAME=gofiber_db
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration (สร้าง secret ใหม่ที่ปลอดภัย)
JWT_SECRET=your_super_secure_jwt_secret_change_this_in_production
JWT_EXPIRE_HOURS=720

# OAuth Configuration (Google)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=https://api.yourdomain.com/api/v1/auth/google/callback

# Frontend URL
FRONTEND_URL=https://yourdomain.com

# Bunny CDN (ถ้าใช้)
BUNNY_STORAGE_API_KEY=your_bunny_storage_api_key
BUNNY_STORAGE_ZONE=your_storage_zone
BUNNY_STORAGE_HOSTNAME=storage.bunnycdn.com
BUNNY_STREAM_API_KEY=your_bunny_stream_api_key
BUNNY_STREAM_LIBRARY_ID=your_library_id

# Cloudflare R2 (ถ้าใช้)
R2_ACCOUNT_ID=your_r2_account_id
R2_ACCESS_KEY_ID=your_r2_access_key
R2_SECRET_ACCESS_KEY=your_r2_secret_key
R2_BUCKET_NAME=your_bucket_name
R2_PUBLIC_URL=https://your-r2-public-url.com

# CORS
CORS_ALLOWED_ORIGINS=https://yourdomain.com

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW_MINUTES=1
```

### 3. Build Application

```bash
# ติดตั้ง dependencies
go mod download
go mod verify

# Build binary
go build -o bin/api cmd/api/main.go

# ตรวจสอบ binary
ls -lh bin/api

# Test run
./bin/api
# ควรเห็น server start ขึ้นมา, กด Ctrl+C เพื่อหยุด
```

### 4. Run Database Migrations

```bash
# Migrations จะรันอัตโนมัติเมื่อ start application
# แต่ถ้าต้องการ test migration:
./bin/api
# ดู log ว่า "✓ Database migrated" ปรากฏหรือไม่
```

---

## Setup Systemd Service

### 1. สร้าง Systemd Service File

```bash
sudo vim /etc/systemd/system/gofiber-api.service
```

**เนื้อหาไฟล์:**

```ini
[Unit]
Description=GoFiber API Service
After=network.target postgresql.service redis.service
Requires=postgresql.service redis.service

[Service]
Type=simple
User=gofiber
Group=gofiber
WorkingDirectory=/home/gofiber/serkk-backend
ExecStart=/home/gofiber/serkk-backend/bin/api
Restart=always
RestartSec=10

# Environment file
EnvironmentFile=/home/gofiber/serkk-backend/.env

# Logging
StandardOutput=append:/var/log/gofiber-api/access.log
StandardError=append:/var/log/gofiber-api/error.log

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/home/gofiber/serkk-backend

# Resource limits
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### 2. สร้าง Log Directory

```bash
sudo mkdir -p /var/log/gofiber-api
sudo chown gofiber:gofiber /var/log/gofiber-api
```

### 3. Enable และ Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service (start on boot)
sudo systemctl enable gofiber-api

# Start service
sudo systemctl start gofiber-api

# ตรวจสอบสถานะ
sudo systemctl status gofiber-api

# ดู logs
sudo journalctl -u gofiber-api -f
```

### 4. Service Management Commands

```bash
# Start service
sudo systemctl start gofiber-api

# Stop service
sudo systemctl stop gofiber-api

# Restart service
sudo systemctl restart gofiber-api

# Check status
sudo systemctl status gofiber-api

# View logs (real-time)
sudo journalctl -u gofiber-api -f

# View logs (last 100 lines)
sudo journalctl -u gofiber-api -n 100

# View error logs only
sudo journalctl -u gofiber-api -p err
```

---

## Setup Nginx Reverse Proxy

### 1. ติดตั้ง Nginx

```bash
sudo apt install -y nginx
```

### 2. สร้าง Nginx Configuration

```bash
sudo vim /etc/nginx/sites-available/gofiber-api
```

**เนื้อหาไฟล์:**

```nginx
# Rate limiting zone
limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;

# Upstream backend
upstream gofiber_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    listen [::]:80;
    server_name api.yourdomain.com;

    # Redirect to HTTPS (จะเปิดใช้หลังติดตั้ง SSL)
    # return 301 https://$server_name$request_uri;

    # Client body size (สำหรับ upload file)
    client_max_body_size 100M;

    # Logging
    access_log /var/log/nginx/gofiber-api-access.log;
    error_log /var/log/nginx/gofiber-api-error.log;

    # Proxy headers
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;

    # WebSocket support
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";

    # Timeouts
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;

    # API endpoints
    location / {
        # Apply rate limiting
        limit_req zone=api_limit burst=20 nodelay;

        proxy_pass http://gofiber_backend;
    }

    # WebSocket endpoints (no rate limit)
    location /ws/ {
        proxy_pass http://gofiber_backend;
    }

    # Health check (no rate limit)
    location /health {
        proxy_pass http://gofiber_backend;
        access_log off;
    }

    # Static files (if any)
    location /static/ {
        alias /home/gofiber/serkk-backend/static/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 3. Enable Configuration

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/gofiber-api /etc/nginx/sites-enabled/

# Remove default site
sudo rm /etc/nginx/sites-enabled/default

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx

# ตรวจสอบสถานะ
sudo systemctl status nginx
```

### 4. Test API

```bash
# Test HTTP
curl http://api.yourdomain.com/health

# ควรได้ response:
# {"status":"ok"}
```

---

## SSL Certificate (HTTPS)

### 1. ติดตั้ง Certbot

```bash
sudo apt install -y certbot python3-certbot-nginx
```

### 2. ขอ SSL Certificate

```bash
# ขอ certificate (Certbot จะแก้ไข Nginx config อัตโนมัติ)
sudo certbot --nginx -d api.yourdomain.com

# ตอบคำถาม:
# - Email: your-email@example.com
# - Agree to terms: Y
# - Share email: N (optional)
# - Redirect HTTP to HTTPS: 2 (Yes)
```

### 3. Test Auto-renewal

```bash
# Test renewal
sudo certbot renew --dry-run

# Certbot จะ auto-renew ทุก 60 วัน
# ตรวจสอบ systemd timer
sudo systemctl status certbot.timer
```

### 4. Test HTTPS

```bash
# Test HTTPS
curl https://api.yourdomain.com/health

# ตรวจสอบ SSL grade
# ไปที่: https://www.ssllabs.com/ssltest/analyze.html?d=api.yourdomain.com
```

---

## Environment Variables

### การจัดการ Environment Variables อย่างปลอดภัย

**1. ใช้ `.env` file (วิธีพื้นฐาน)**

```bash
# แก้ไข .env
vim /home/gofiber/serkk-backend/.env

# Restart service หลังแก้ไข
sudo systemctl restart gofiber-api
```

**2. ใช้ systemd EnvironmentFile (แนะนำ)**

ใน systemd service file มีบรรทัดนี้อยู่แล้ว:
```ini
EnvironmentFile=/home/gofiber/serkk-backend/.env
```

**3. ใช้ Secret Management (Advanced)**

สำหรับ production ขนาดใหญ่ ควรใช้:
- HashiCorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Google Cloud Secret Manager

---

## Database Migration

### วิธีการ Migration

Application นี้ใช้ SQL migration files ใน folder `migrations/`

**1. Migration จะรันอัตโนมัติเมื่อ start:**

```bash
# Migration รันเมื่อ start service
sudo systemctl start gofiber-api

# ดู log
sudo journalctl -u gofiber-api | grep "Database migrated"
```

**2. Manual Migration (ถ้าต้องการ):**

```bash
# เข้าไปใน directory
cd /home/gofiber/serkk-backend

# Run application (จะ migrate อัตโนมัติ)
./bin/api
```

**3. ตรวจสอบ Tables:**

```bash
psql -h localhost -U gofiber_user -d gofiber_db -c "\dt"
```

### การเพิ่ม Migration ใหม่

เมื่อมีการเปลี่ยนแปลง database schema:

1. สร้างไฟล์ migration ใหม่:
```bash
# ตัวอย่าง: 002_add_new_feature.sql
vim migrations/002_add_new_feature.sql
```

2. แก้ไข `infrastructure/postgres/database.go` ให้รัน migration ใหม่

3. Deploy:
```bash
# Pull code ใหม่
git pull origin main

# Rebuild
go build -o bin/api cmd/api/main.go

# Restart service (จะ run migration อัตโนมัติ)
sudo systemctl restart gofiber-api
```

---

## Monitoring & Logging

### 1. Application Logs

```bash
# Real-time logs
sudo journalctl -u gofiber-api -f

# Last 100 lines
sudo journalctl -u gofiber-api -n 100

# Today's logs
sudo journalctl -u gofiber-api --since today

# Errors only
sudo journalctl -u gofiber-api -p err

# Custom log files
tail -f /var/log/gofiber-api/access.log
tail -f /var/log/gofiber-api/error.log
```

### 2. Nginx Logs

```bash
# Access logs
tail -f /var/log/nginx/gofiber-api-access.log

# Error logs
tail -f /var/log/nginx/gofiber-api-error.log
```

### 3. Database Logs

```bash
# PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

### 4. Redis Logs

```bash
# Redis logs
sudo tail -f /var/log/redis/redis-server.log
```

### 5. System Monitoring

```bash
# ติดตั้ง monitoring tools
sudo apt install -y htop iotop nethogs

# CPU, Memory usage
htop

# Disk I/O
sudo iotop

# Network usage
sudo nethogs

# Disk space
df -h

# Memory usage
free -h
```

### 6. Application Metrics (Built-in)

Application มี `/metrics` endpoint สำหรับ Prometheus:

```bash
curl http://localhost:8080/metrics
```

**Setup Prometheus + Grafana (Optional):**

```bash
# ติดตั้ง Prometheus
sudo apt install -y prometheus

# ติดตั้ง Grafana
sudo apt-get install -y software-properties-common
sudo add-apt-repository "deb https://packages.grafana.com/oss/deb stable main"
wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
sudo apt-get update
sudo apt-get install -y grafana

# Start services
sudo systemctl start prometheus
sudo systemctl start grafana-server
sudo systemctl enable prometheus
sudo systemctl enable grafana-server
```

---

## Backup Strategy

### 1. Database Backup

**Script สำหรับ backup อัตโนมัติ:**

```bash
sudo vim /usr/local/bin/backup-gofiber-db.sh
```

**เนื้อหา script:**

```bash
#!/bin/bash

# Configuration
DB_NAME="gofiber_db"
DB_USER="gofiber_user"
BACKUP_DIR="/home/gofiber/backups/database"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/gofiber_db_$DATE.sql.gz"
RETENTION_DAYS=7

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup database
PGPASSWORD='your_secure_password_here' pg_dump -h localhost -U $DB_USER $DB_NAME | gzip > $BACKUP_FILE

# Delete old backups
find $BACKUP_DIR -name "gofiber_db_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $BACKUP_FILE"
```

**ทำให้ script executable:**

```bash
sudo chmod +x /usr/local/bin/backup-gofiber-db.sh
```

**Setup cron job:**

```bash
# Edit crontab
crontab -e

# เพิ่มบรรทัดนี้ (backup ทุกวัน เวลา 2:00 AM)
0 2 * * * /usr/local/bin/backup-gofiber-db.sh >> /var/log/gofiber-backup.log 2>&1
```

### 2. Restore Database

```bash
# Uncompress และ restore
gunzip -c /home/gofiber/backups/database/gofiber_db_20250112_020000.sql.gz | \
    PGPASSWORD='your_secure_password_here' psql -h localhost -U gofiber_user -d gofiber_db
```

### 3. Application Files Backup

```bash
# Backup application และ config
tar -czf /home/gofiber/backups/app_backup_$(date +%Y%m%d).tar.gz \
    /home/gofiber/serkk-backend/.env \
    /home/gofiber/serkk-backend/bin/api

# Or sync to remote storage (S3, etc.)
aws s3 sync /home/gofiber/backups/ s3://your-bucket/backups/
```

---

## CI/CD (Optional)

### GitHub Actions Example

สร้างไฟล์ `.github/workflows/deploy.yml`:

```yaml
name: Deploy to Production

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Run tests
      run: go test ./...

    - name: Deploy to server
      uses: appleboy/ssh-action@master
      with:
        host: ${{ secrets.SERVER_HOST }}
        username: ${{ secrets.SERVER_USER }}
        key: ${{ secrets.SSH_PRIVATE_KEY }}
        script: |
          cd /home/gofiber/serkk-backend
          git pull origin main
          go build -o bin/api cmd/api/main.go
          sudo systemctl restart gofiber-api
          sleep 5
          sudo systemctl status gofiber-api
```

**Setup GitHub Secrets:**

1. ไปที่ GitHub repo → Settings → Secrets → Actions
2. เพิ่ม secrets:
   - `SERVER_HOST`: IP ของ server
   - `SERVER_USER`: username (gofiber)
   - `SSH_PRIVATE_KEY`: private key สำหรับ SSH

---

## 🔐 Security Checklist

### Pre-deployment Security

- [ ] เปลี่ยน JWT secret ใหม่
- [ ] เปลี่ยน database password ที่ปลอดภัย
- [ ] ตั้งค่า CORS_ALLOWED_ORIGINS ให้ถูกต้อง
- [ ] Enable rate limiting
- [ ] ตรวจสอบ sensitive data ใน code
- [ ] ลบ debug endpoints ใน production
- [ ] Enable HTTPS only
- [ ] Setup firewall rules

### Post-deployment Security

- [ ] ตรวจสอบ SSL/TLS configuration
- [ ] Enable automatic security updates
- [ ] Setup fail2ban สำหรับป้องกัน brute force
- [ ] Regular security audits
- [ ] Monitor logs สำหรับ suspicious activities
- [ ] Backup database เป็นประจำ

---

## 🚨 Troubleshooting

### Service ไม่ start

```bash
# ดู detailed error
sudo journalctl -u gofiber-api -n 50

# ตรวจสอบ permissions
ls -la /home/gofiber/serkk-backend/bin/api

# ตรวจสอบ .env file
cat /home/gofiber/serkk-backend/.env
```

### Database connection error

```bash
# ตรวจสอบ PostgreSQL running
sudo systemctl status postgresql

# Test connection
psql -h localhost -U gofiber_user -d gofiber_db

# ดู PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-15-main.log
```

### Nginx error

```bash
# Test config
sudo nginx -t

# ดู error logs
sudo tail -f /var/log/nginx/gofiber-api-error.log

# Restart Nginx
sudo systemctl restart nginx
```

### High memory usage

```bash
# ดู process memory
ps aux | grep api

# Restart service
sudo systemctl restart gofiber-api
```

---

## 📞 Support

หากมีปัญหาในการ deploy:

1. ตรวจสอบ logs: `sudo journalctl -u gofiber-api -f`
2. ดู GitHub Issues: https://github.com/thizplus/serkk-backend/issues
3. อ่านเอกสารเพิ่มเติม: `GETTING_STARTED_TH.md`

---

## 📝 Maintenance Tasks

### Daily
- ตรวจสอบ logs สำหรับ errors
- ดู metrics และ performance

### Weekly
- ตรวจสอบ disk space
- Review security logs
- Test backups

### Monthly
- Update dependencies
- Security patches
- Review and rotate logs
- Database optimization

---

**🎉 Deployment สำเร็จ!**

API ของคุณพร้อมใช้งานแล้วที่: `https://api.yourdomain.com`

