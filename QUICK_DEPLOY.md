# ⚡ Quick Deploy Guide

คู่มือ deploy แบบเร็วสำหรับคนที่รู้จัก Linux แล้ว (10-15 นาที)

---

## 📦 Prerequisites

- Ubuntu 22.04 LTS server
- Domain name ที่ชี้ไปที่ server IP
- SSH access

---

## 🚀 Quick Setup

### 1. Initial Server Setup (5 min)

```bash
# SSH to server
ssh root@your-server-ip

# Update system
apt update && apt upgrade -y

# Create user
adduser gofiber
usermod -aG sudo gofiber
su - gofiber
```

### 2. Install Dependencies (3 min)

```bash
# Install Go 1.24
wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install PostgreSQL 15
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo tee /etc/apt/trusted.gpg.d/pgdg.asc
sudo apt update && sudo apt install -y postgresql-15 redis-server nginx

# Start services
sudo systemctl start postgresql redis nginx
```

### 3. Setup Database (2 min)

```bash
# Create database
sudo -u postgres psql << EOF
CREATE DATABASE gofiber_db;
CREATE USER gofiber_user WITH ENCRYPTED PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE gofiber_db TO gofiber_user;
\c gofiber_db
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
EOF
```

### 4. Deploy Application (3 min)

```bash
# Clone repo
cd ~
git clone https://github.com/thizplus/serkk-backend.git
cd serkk-backend

# Create .env
cp .env.production.example .env
vim .env  # แก้ไขค่าที่จำเป็น

# Build
go mod download
go build -o bin/api cmd/api/main.go
```

### 5. Setup Systemd Service (1 min)

```bash
# Copy service file
sudo cp deploy/systemd/gofiber-api.service /etc/systemd/system/

# Create log directory
sudo mkdir -p /var/log/gofiber-api
sudo chown gofiber:gofiber /var/log/gofiber-api

# Start service
sudo systemctl daemon-reload
sudo systemctl enable gofiber-api
sudo systemctl start gofiber-api
sudo systemctl status gofiber-api
```

### 6. Setup Nginx (2 min)

```bash
# Copy nginx config
sudo cp deploy/nginx/gofiber-api.conf /etc/nginx/sites-available/
sudo ln -s /etc/nginx/sites-available/gofiber-api.conf /etc/nginx/sites-enabled/
sudo rm /etc/nginx/sites-enabled/default

# Update domain
sudo sed -i 's/api.yourdomain.com/your-actual-domain.com/g' /etc/nginx/sites-available/gofiber-api.conf

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

### 7. Setup SSL (1 min)

```bash
# Install Certbot
sudo apt install -y certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d api.yourdomain.com
```

---

## ✅ Verify Deployment

```bash
# Check service
sudo systemctl status gofiber-api

# Check health
curl http://localhost:8080/health
curl https://api.yourdomain.com/health

# View logs
sudo journalctl -u gofiber-api -f
```

---

## 🔧 Quick Commands

```bash
# Restart service
sudo systemctl restart gofiber-api

# View logs
sudo journalctl -u gofiber-api -f

# Update application
cd ~/serkk-backend
git pull origin main
go build -o bin/api cmd/api/main.go
sudo systemctl restart gofiber-api

# Backup database
./deploy/scripts/backup-db.sh
```

---

## 📝 Important Files

```
/home/gofiber/serkk-backend/          # Application directory
/home/gofiber/serkk-backend/.env      # Environment variables
/etc/systemd/system/gofiber-api.service  # Systemd service
/etc/nginx/sites-available/gofiber-api.conf  # Nginx config
/var/log/gofiber-api/                 # Application logs
```

---

## 🆘 Quick Troubleshooting

**Service won't start:**
```bash
sudo journalctl -u gofiber-api -n 50
```

**Database connection error:**
```bash
psql -h localhost -U gofiber_user -d gofiber_db
```

**Nginx error:**
```bash
sudo nginx -t
sudo tail -f /var/log/nginx/error.log
```

---

## 📚 Full Documentation

สำหรับรายละเอียดเพิ่มเติม อ่าน: **DEPLOYMENT.md**

---

**🎉 Done! Your API is live at `https://api.yourdomain.com`**
