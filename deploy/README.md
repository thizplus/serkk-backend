# 📦 Deployment Files

ไฟล์และ scripts สำหรับ deploy GoFiber backend ไปยัง production server

---

## 📁 Directory Structure

```
deploy/
├── nginx/
│   └── gofiber-api.conf          # Nginx reverse proxy configuration
├── scripts/
│   ├── deploy.sh                 # Automated deployment script
│   └── backup-db.sh              # Database backup script
└── systemd/
    └── gofiber-api.service       # Systemd service configuration
```

---

## 🔧 Usage

### Nginx Configuration

```bash
# Copy to nginx
sudo cp deploy/nginx/gofiber-api.conf /etc/nginx/sites-available/

# Enable site
sudo ln -s /etc/nginx/sites-available/gofiber-api.conf /etc/nginx/sites-enabled/

# Update your domain name
sudo sed -i 's/api.yourdomain.com/your-actual-domain.com/g' /etc/nginx/sites-available/gofiber-api.conf

# Test and reload
sudo nginx -t
sudo systemctl reload nginx
```

### Systemd Service

```bash
# Copy service file
sudo cp deploy/systemd/gofiber-api.service /etc/systemd/system/

# Update paths if needed (default: /home/gofiber/serkk-backend)
sudo vim /etc/systemd/system/gofiber-api.service

# Create log directory
sudo mkdir -p /var/log/gofiber-api
sudo chown gofiber:gofiber /var/log/gofiber-api

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable gofiber-api
sudo systemctl start gofiber-api
```

### Deployment Script

```bash
# Make executable
chmod +x deploy/scripts/deploy.sh

# Run deployment
./deploy/scripts/deploy.sh
```

**What it does:**
1. Pulls latest code from GitHub
2. Downloads dependencies
3. Runs tests (optional)
4. Builds application
5. Restarts service
6. Performs health check
7. Shows recent logs

### Database Backup Script

```bash
# Edit configuration
vim deploy/scripts/backup-db.sh
# Update: DB_PASSWORD="your_actual_password"

# Make executable
chmod +x deploy/scripts/backup-db.sh

# Run backup
./deploy/scripts/backup-db.sh

# Setup automated backup (cron)
crontab -e
# Add: 0 2 * * * /home/gofiber/serkk-backend/deploy/scripts/backup-db.sh
```

---

## 📝 Configuration Notes

### Nginx

- Default port: `8080` (backend)
- Rate limiting: `10 req/s` with burst `20`
- Max body size: `100MB` (for file uploads)
- WebSocket support: Enabled
- SSL: Add after running Certbot

### Systemd Service

- User: `gofiber`
- Working directory: `/home/gofiber/serkk-backend`
- Restart policy: `always` (10s delay)
- Logs: `/var/log/gofiber-api/`
- Environment: Loaded from `.env` file

### Backup Script

- Retention: `7 days`
- Backup location: `/home/gofiber/backups/database/`
- Format: `gofiber_db_YYYYMMDD_HHMMSS.sql.gz`

---

## 🔐 Security Checklist

Before deploying:

- [ ] Update `DB_PASSWORD` in `.env`
- [ ] Update `JWT_SECRET` in `.env`
- [ ] Update `CORS_ALLOWED_ORIGINS` in `.env`
- [ ] Update domain name in Nginx config
- [ ] Update database password in `backup-db.sh`
- [ ] Setup firewall rules
- [ ] Enable HTTPS (Certbot)
- [ ] Review systemd service permissions

---

## 📚 Related Documentation

- **DEPLOYMENT.md** - Full deployment guide
- **QUICK_DEPLOY.md** - Quick 10-minute deployment
- **.env.production.example** - Production environment variables example

---

## 🆘 Troubleshooting

**Nginx won't start:**
```bash
sudo nginx -t  # Test configuration
sudo journalctl -u nginx -n 50
```

**Service won't start:**
```bash
sudo journalctl -u gofiber-api -n 50
ls -la /home/gofiber/serkk-backend/bin/api
```

**Deployment script fails:**
```bash
# Check permissions
ls -la deploy/scripts/deploy.sh
# Should be: -rwxr-xr-x

# Run manually
cd /home/gofiber/serkk-backend
git pull origin main
go build -o bin/api cmd/api/main.go
```

---

## 💡 Tips

1. **Test locally first:**
   ```bash
   ./bin/api  # Test run before deploying
   ```

2. **Use deploy script:**
   ```bash
   ./deploy/scripts/deploy.sh  # Automated deployment
   ```

3. **Monitor logs:**
   ```bash
   sudo journalctl -u gofiber-api -f
   ```

4. **Regular backups:**
   ```bash
   # Setup daily backup at 2 AM
   crontab -e
   0 2 * * * /home/gofiber/serkk-backend/deploy/scripts/backup-db.sh
   ```

---

**Need help?** Check the full guide: **../DEPLOYMENT.md**
