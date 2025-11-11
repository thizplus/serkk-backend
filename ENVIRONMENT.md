# 🔧 Environment Configuration Guide

This guide covers all environment variables and configuration options for the GoFiber Social Media API.

## 📋 Quick Setup

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your values

3. Start the application:
   ```bash
   go run ./cmd/api
   ```

## 🔐 Environment Variables

### Application Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `APP_NAME` | No | `GoFiber Template` | Application name for logging |
| `APP_PORT` | No | `3000` | HTTP server port |
| `APP_ENV` | No | `development` | Environment: `development`, `staging`, `production` |
| `APP_VERSION` | No | `1.0.0` | Application version for health checks |

**Example:**
```env
APP_NAME=My Social Media API
APP_PORT=8080
APP_ENV=production
APP_VERSION=1.2.3
```

---

### Database Configuration (PostgreSQL)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_HOST` | Yes | `localhost` | PostgreSQL host |
| `DB_PORT` | Yes | `5432` | PostgreSQL port |
| `DB_USER` | Yes | `postgres` | Database user |
| `DB_PASSWORD` | Yes | - | Database password |
| `DB_NAME` | Yes | `gofiber_template` | Database name |
| `DB_SSL_MODE` | No | `disable` | SSL mode: `disable`, `require`, `verify-ca`, `verify-full` |

**Example:**
```env
DB_HOST=db.example.com
DB_PORT=5432
DB_USER=myapp
DB_PASSWORD=super-secret-password
DB_NAME=social_media_prod
DB_SSL_MODE=require
```

**Connection Pool Settings** (configured in code):
- Development: 10 idle, 50 max connections
- Production: 25 idle, 100 max connections

---

### Redis Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_HOST` | Yes | `localhost` | Redis host |
| `REDIS_PORT` | Yes | `6379` | Redis port |
| `REDIS_PASSWORD` | No | - | Redis password (if auth enabled) |
| `REDIS_DB` | No | `0` | Redis database number (0-15) |

**Example:**
```env
REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=redis-secret
REDIS_DB=0
```

**Use Cases:**
- Session management
- Cache storage
- Rate limiting
- Real-time features

---

### JWT Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `JWT_SECRET` | Yes | - | Secret key for JWT signing (min 32 characters) |

**Example:**
```env
JWT_SECRET=my-super-secret-jwt-key-min-32-characters-long-please
```

**Security Notes:**
- ⚠️ Never commit this to version control
- Use a strong random string (64+ characters recommended)
- Rotate regularly in production
- Use different secrets for dev/staging/prod

**Generate a secure key:**
```bash
# Using OpenSSL
openssl rand -base64 64

# Using Go
go run -c 'import "crypto/rand"; import "encoding/base64"; b := make([]byte, 64); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b))'
```

---

### Bunny CDN Storage (Images & Files)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BUNNY_STORAGE_ZONE` | No | - | Storage zone name |
| `BUNNY_ACCESS_KEY` | No | - | Storage access key |
| `BUNNY_BASE_URL` | No | `https://storage.bunnycdn.com` | Storage API base URL |
| `BUNNY_CDN_URL` | No | - | CDN pull zone URL |

**Example:**
```env
BUNNY_STORAGE_ZONE=my-storage-zone
BUNNY_ACCESS_KEY=abc123-def456-ghi789
BUNNY_BASE_URL=https://storage.bunnycdn.com
BUNNY_CDN_URL=https://mycdn.b-cdn.net
```

**Setup Guide:**
1. Create account at https://bunny.net
2. Create Storage Zone
3. Create Pull Zone (CDN)
4. Get Access Key from Storage Zone settings
5. Configure CORS if needed

---

### Bunny Stream (Video Hosting)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BUNNY_STREAM_API_KEY` | No | - | Stream API key |
| `BUNNY_STREAM_LIBRARY_ID` | No | - | Stream library ID |
| `BUNNY_STREAM_CDN_URL` | No | - | Stream CDN URL |

**Example:**
```env
BUNNY_STREAM_API_KEY=stream-api-key-here
BUNNY_STREAM_LIBRARY_ID=12345
BUNNY_STREAM_CDN_URL=https://vz-abcd1234.b-cdn.net
```

---

### Cloudflare R2 Storage (Alternative)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `R2_ACCOUNT_ID` | No | - | Cloudflare account ID |
| `R2_ACCESS_KEY_ID` | No | - | R2 access key ID |
| `R2_SECRET_ACCESS_KEY` | No | - | R2 secret access key |
| `R2_BUCKET_NAME` | No | - | R2 bucket name |
| `R2_PUBLIC_URL` | No | - | Public URL for R2 bucket |

**Example:**
```env
R2_ACCOUNT_ID=abc123def456
R2_ACCESS_KEY_ID=r2-access-key-id
R2_SECRET_ACCESS_KEY=r2-secret-key-very-long-and-secure
R2_BUCKET_NAME=my-media-bucket
R2_PUBLIC_URL=https://media.mydomain.com
```

**Setup Guide:**
1. Enable R2 in Cloudflare Dashboard
2. Create bucket
3. Generate API tokens
4. Configure custom domain (optional)
5. Set CORS policy if needed

---

### Google OAuth

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_CLIENT_ID` | No | - | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | - | Google OAuth client secret |
| `GOOGLE_REDIRECT_URL` | No | - | OAuth redirect URL |

**Example:**
```env
GOOGLE_CLIENT_ID=123456789-abcdef.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GOCSPX-abc123def456
GOOGLE_REDIRECT_URL=https://api.example.com/api/v1/auth/google/callback
```

**Setup Guide:**
1. Go to Google Cloud Console
2. Create project
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add authorized redirect URIs
6. Copy Client ID and Secret

---

### VAPID (Push Notifications)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VAPID_PUBLIC_KEY` | No | - | VAPID public key (base64url) |
| `VAPID_PRIVATE_KEY` | No | - | VAPID private key (base64url) |
| `VAPID_SUBJECT` | No | - | Contact email for push service |

**Example:**
```env
VAPID_PUBLIC_KEY=BKxJxxxxxxxxxxxxxxxxxxxxxxxxxxx
VAPID_PRIVATE_KEY=Yxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
VAPID_SUBJECT=mailto:admin@example.com
```

**Generate Keys:**
```bash
# Using web-push (Node.js)
npx web-push generate-vapid-keys

# Or create a Go script (scripts/generate_vapid_keys.go)
```

---

### Frontend URL

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FRONTEND_URL` | Yes | `http://localhost:3000` | Frontend application URL for CORS and redirects |

**Example:**
```env
FRONTEND_URL=https://app.example.com
```

**Usage:**
- CORS allowed origins
- OAuth redirect after authentication
- Email verification links
- Password reset links

---

## 🌍 Environment-Specific Configurations

### Development (.env)
```env
APP_ENV=development
DB_HOST=localhost
REDIS_HOST=localhost
FRONTEND_URL=http://localhost:3000
# Optional: Leave storage configs empty to use mock/local storage
```

### Staging (.env.staging)
```env
APP_ENV=staging
DB_HOST=staging-db.example.com
DB_SSL_MODE=require
REDIS_HOST=staging-redis.example.com
FRONTEND_URL=https://staging.example.com
# All storage services configured
```

### Production (.env.production)
```env
APP_ENV=production
DB_HOST=prod-db.example.com
DB_SSL_MODE=require
REDIS_HOST=prod-redis.example.com
REDIS_PASSWORD=strong-redis-password
FRONTEND_URL=https://app.example.com
# All services configured with production credentials
# Use secrets management in production!
```

---

## 🔒 Security Best Practices

### 1. Never Commit Secrets
```bash
# Add to .gitignore
.env
.env.local
.env.production
.env.staging
```

### 2. Use Different Keys per Environment
- Development: Simple keys for testing
- Staging: Real keys but isolated
- Production: Strong keys, rotated regularly

### 3. Use Secrets Management
**Options:**
- AWS Secrets Manager
- HashiCorp Vault
- Kubernetes Secrets
- Docker Secrets
- Environment variables from CI/CD

**Example with Docker Secrets:**
```yaml
services:
  app:
    secrets:
      - db_password
      - jwt_secret

secrets:
  db_password:
    external: true
  jwt_secret:
    external: true
```

### 4. Validate Configuration
Application validates required variables on startup and fails fast if missing.

### 5. Rotate Secrets
- JWT Secret: Every 90 days
- API Keys: When compromised
- Database Password: Quarterly

---

## 🧪 Testing Configuration

### Unit Tests
```env
APP_ENV=test
DB_NAME=gofiber_test
REDIS_DB=1
JWT_SECRET=test-jwt-secret-not-for-production
```

### Integration Tests
```bash
# Use docker-compose for test database
docker-compose -f docker-compose.test.yml up -d
```

---

## 🔍 Validation & Debugging

### Check Configuration
```bash
# View loaded config (without secrets)
go run ./cmd/api --check-config

# Test database connection
go run ./cmd/api --test-db

# Test Redis connection
go run ./cmd/api --test-redis
```

### Common Issues

**Database Connection Failed:**
```bash
# Check if PostgreSQL is running
psql -h localhost -U postgres -d gofiber_template

# Check firewall rules
# Check SSL mode settings
```

**Redis Connection Failed:**
```bash
# Check if Redis is running
redis-cli ping

# Test with password
redis-cli -a your-password ping
```

**JWT Invalid:**
- Check JWT_SECRET is set
- Ensure minimum 32 characters
- Check for trailing spaces/newlines

---

## 📚 Resources

- [Go Environment Variables](https://golang.org/pkg/os/#Getenv)
- [PostgreSQL SSL Modes](https://www.postgresql.org/docs/current/libpq-ssl.html)
- [Redis Configuration](https://redis.io/topics/config)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [VAPID Protocol](https://datatracker.ietf.org/doc/html/rfc8292)

---

## 🆘 Getting Help

For configuration issues:

1. Check environment variable names (case-sensitive)
2. Verify no trailing spaces or quotes
3. Check application logs for validation errors
4. Use `.env.example` as reference
5. Create GitHub issue with error details (redact secrets!)
