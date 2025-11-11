# 🐳 Docker Deployment Guide

This guide covers how to deploy the GoFiber Social Media API using Docker and Docker Compose.

## 📋 Prerequisites

- Docker Engine 20.10+
- Docker Compose 2.0+
- (Optional) Make for running commands

## 🚀 Quick Start

### Development

```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f app

# Stop all services
docker-compose -f docker-compose.dev.yml down
```

### Production

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down
```

## 📦 Docker Images

### Application Image

**Base Image:** `golang:1.21-alpine`
**Final Image:** `alpine:latest` (with non-root user)

**Features:**
- Multi-stage build for minimal size
- Non-root user for security
- Health checks configured
- Static binary compilation

## 🎯 Services

### 1. Application (app)

**Image:** Built from Dockerfile
**Port:** 3000 (internal), 8080 (external in prod)
**Dependencies:** PostgreSQL, Redis

**Health Check:**
- Endpoint: `/health`
- Interval: 30s
- Timeout: 10s
- Retries: 3

### 2. PostgreSQL (postgres)

**Image:** `postgres:15-alpine`
**Port:** 5432
**Data:** Persisted in volume `postgres_data`

**Health Check:**
- Command: `pg_isready`
- Interval: 10s
- Timeout: 5s
- Retries: 5

### 3. Redis (redis)

**Image:** `redis:7-alpine`
**Port:** 6379
**Data:** Persisted in volume `redis_data`

**Health Check:**
- Command: `redis-cli ping`
- Interval: 10s
- Timeout: 5s
- Retries: 5

### 4. pgAdmin (optional, dev only)

**Image:** `dpage/pgadmin4:latest`
**Port:** 5050
**Profile:** tools

Access at: http://localhost:5050
- Email: admin@example.com
- Password: admin

### 5. RedisInsight (optional, dev only)

**Image:** `redislabs/redisinsight:latest`
**Port:** 8001
**Profile:** tools

Access at: http://localhost:8001

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the project root:

```env
# Application
APP_NAME=GoFiber Social Media API
APP_PORT=3000
APP_ENV=production
APP_VERSION=1.0.0

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-secure-password
DB_NAME=gofiber_prod
DB_SSL_MODE=disable

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0

# JWT
JWT_SECRET=your-super-secret-jwt-key-min-32-characters

# Storage (Bunny CDN)
BUNNY_STORAGE_ZONE=your-storage-zone
BUNNY_ACCESS_KEY=your-access-key
BUNNY_BASE_URL=https://storage.bunnycdn.com
BUNNY_CDN_URL=https://your-cdn-url.b-cdn.net

# R2 Storage (Cloudflare)
R2_ACCOUNT_ID=your-account-id
R2_ACCESS_KEY_ID=your-access-key
R2_SECRET_ACCESS_KEY=your-secret-key
R2_BUCKET_NAME=your-bucket
R2_PUBLIC_URL=https://your-r2-url.com

# OAuth (Google)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=https://your-domain.com/auth/google/callback

# Push Notifications (VAPID)
VAPID_PUBLIC_KEY=your-public-key
VAPID_PRIVATE_KEY=your-private-key
VAPID_SUBJECT=mailto:your-email@example.com

# Frontend
FRONTEND_URL=https://your-frontend-domain.com
```

### Using .env File

```bash
# Load .env file
docker-compose --env-file .env up -d
```

## 📝 Common Commands

### Build

```bash
# Build without cache
docker-compose build --no-cache

# Build specific service
docker-compose build app
```

### Run

```bash
# Start all services
docker-compose up -d

# Start with specific profile (dev tools)
docker-compose -f docker-compose.dev.yml --profile tools up -d

# Start specific service
docker-compose up -d app

# View logs
docker-compose logs -f app
docker-compose logs --tail=100 app

# Follow logs of all services
docker-compose logs -f
```

### Execute Commands

```bash
# Execute command in app container
docker-compose exec app sh

# Run migrations
docker-compose exec app ./main migrate

# Check application health
docker-compose exec app wget -O- http://localhost:3000/health
```

### Stop and Clean

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Stop and remove images
docker-compose down --rmi all

# Complete cleanup
docker-compose down -v --rmi all --remove-orphans
```

### Database Operations

```bash
# PostgreSQL shell
docker-compose exec postgres psql -U postgres -d gofiber_template

# Run SQL file
docker-compose exec -T postgres psql -U postgres -d gofiber_template < backup.sql

# Create database backup
docker-compose exec postgres pg_dump -U postgres gofiber_template > backup.sql

# Restore database
docker-compose exec -T postgres psql -U postgres -d gofiber_template < backup.sql
```

### Redis Operations

```bash
# Redis CLI
docker-compose exec redis redis-cli

# Flush all data
docker-compose exec redis redis-cli FLUSHALL

# Monitor commands
docker-compose exec redis redis-cli MONITOR
```

## 🔍 Monitoring

### View Container Stats

```bash
# Real-time stats
docker stats

# Specific service stats
docker stats gofiber-app
```

### Inspect Containers

```bash
# Container details
docker-compose ps

# Detailed info
docker inspect gofiber-app

# Check logs
docker-compose logs --tail=50 app
```

### Health Checks

```bash
# Check health status
docker-compose ps

# Manual health check
curl http://localhost:8080/health
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
```

## 🐛 Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs app

# Check if ports are already in use
netstat -an | grep 3000

# Restart services
docker-compose restart app
```

### Database Connection Issues

```bash
# Check if PostgreSQL is healthy
docker-compose ps postgres

# Test connection
docker-compose exec postgres pg_isready -U postgres

# Check environment variables
docker-compose exec app env | grep DB_
```

### Cannot Connect to Redis

```bash
# Check if Redis is running
docker-compose ps redis

# Test connection
docker-compose exec redis redis-cli ping

# Check logs
docker-compose logs redis
```

### Out of Memory

```bash
# Check memory usage
docker stats

# Increase Docker memory limit
# Docker Desktop > Settings > Resources > Memory
```

### Permission Denied

```bash
# Rebuild with --no-cache
docker-compose build --no-cache

# Check file permissions
ls -la Dockerfile docker-compose.yml
```

## 🔒 Security Best Practices

1. **Never commit secrets** - Use `.env` file or secrets management
2. **Use non-root user** - Already configured in Dockerfile
3. **Keep images updated** - Regularly update base images
4. **Scan for vulnerabilities**
   ```bash
   docker scan gofiber-app:latest
   ```
5. **Limit resource usage** - Configure memory and CPU limits
6. **Use secrets** - For sensitive data in production
   ```yaml
   secrets:
     db_password:
       external: true
   ```

## 📊 Performance Optimization

### Build Performance

```bash
# Use BuildKit
DOCKER_BUILDKIT=1 docker-compose build

# Cache Go modules
# Already configured in Dockerfile with separate RUN for go mod download
```

### Runtime Performance

```bash
# Limit resources in docker-compose.yml
services:
  app:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '1'
          memory: 512M
```

## 🚀 Production Deployment

### Using Docker Swarm

```bash
# Initialize swarm
docker swarm init

# Deploy stack
docker stack deploy -c docker-compose.yml gofiber

# List services
docker service ls

# Scale service
docker service scale gofiber_app=3

# Remove stack
docker stack rm gofiber
```

### Using Kubernetes

Convert docker-compose to Kubernetes manifests:

```bash
# Using kompose
kompose convert -f docker-compose.yml
```

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)

## 🆘 Getting Help

If you encounter issues:

1. Check logs: `docker-compose logs -f`
2. Verify environment variables: `docker-compose config`
3. Check service health: `docker-compose ps`
4. Restart services: `docker-compose restart`
5. Clean rebuild: `docker-compose down -v && docker-compose up --build`

For more help, create an issue in the GitHub repository.
