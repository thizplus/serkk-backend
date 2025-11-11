# 🚀 GoFiber Social Media API

A production-ready social media backend API built with Go Fiber, featuring real-time chat, posts, comments, notifications, and comprehensive testing.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Fiber](https://img.shields.io/badge/Fiber-v2.52-00ACD7?style=flat)](https://gofiber.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-93%20passing-success)](.)

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Documentation](#-documentation)
- [API Endpoints](#-api-endpoints)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Contributing](#-contributing)

## ✨ Features

### Core Features
- 🔐 **Authentication & Authorization** - JWT-based auth + OAuth2 (Google)
- 👤 **User Management** - Profiles, follow/unfollow, settings
- 📝 **Posts** - Create, read, update, delete with media support
- 💬 **Comments** - Nested comments with voting
- 👍 **Voting System** - Upvote/downvote on posts and comments
- 🔖 **Saved Posts** - Bookmark favorite content
- 🔍 **Search** - Full-text search with history
- 🏷️ **Tags** - Categorize content
- 💬 **Real-time Chat** - WebSocket-based messaging
- 🔔 **Notifications** - Real-time notifications via WebSocket & Push API
- 📁 **Media Upload** - Images & videos (Bunny CDN + Cloudflare R2)
- 🎥 **Video Streaming** - Bunny Stream integration

### Production Features
- ✅ **93 Passing Tests** - 86 unit/integration tests + 7 compression tests
- 🛡️ **Security Hardened** - Rate limiting, CORS, security headers, input validation
- 📊 **Monitoring** - Health checks, metrics collection, structured logging
- 🗜️ **Performance** - Response compression, database indexing, caching
- 📚 **API Documentation** - Interactive Swagger/OpenAPI docs
- 🐳 **Docker Ready** - Multi-stage builds, docker-compose
- 🔄 **CI/CD** - GitHub Actions with automated testing & security scans
- ♻️ **Graceful Shutdown** - Clean resource cleanup
- 🎯 **Clean Architecture** - Domain-driven design, SOLID principles

## 🛠️ Tech Stack

**Core:**
- [Go](https://golang.org/) 1.21+ - Programming language
- [Fiber v2](https://gofiber.io/) - Web framework
- [GORM v1.30](https://gorm.io/) - ORM
- [PostgreSQL 15](https://www.postgresql.org/) - Primary database
- [Redis 7](https://redis.io/) - Caching & sessions

**Storage:**
- [Bunny CDN](https://bunny.net/) - Image/file storage & video streaming
- [Cloudflare R2](https://www.cloudflare.com/products/r2/) - Alternative object storage

**Authentication:**
- JWT tokens
- Google OAuth2

**Real-time:**
- WebSocket - Chat & notifications
- Server-Sent Events - Live updates

**Testing:**
- [testify](https://github.com/stretchr/testify) - Testing toolkit
- [go-faker](https://github.com/go-faker/faker) - Test data generation

**Tools:**
- [Air](https://github.com/cosmtrek/air) - Live reload
- [golangci-lint](https://golangci-lint.run/) - Linting
- [Swagger/OpenAPI](https://swagger.io/) - API documentation
- [Docker](https://www.docker.com/) - Containerization

## 🏗️ Architecture

Clean Architecture / Hexagonal Architecture pattern:

```
├── cmd/api/              # Application entry point
├── domain/               # Business logic layer
│   ├── models/          # Domain entities
│   ├── dto/             # Data transfer objects
│   ├── repositories/    # Repository interfaces
│   └── services/        # Service interfaces
├── application/          # Application logic layer
│   └── serviceimpl/     # Service implementations
├── infrastructure/       # Infrastructure layer
│   ├── postgres/        # Database implementations
│   ├── redis/           # Cache implementations
│   ├── storage/         # File storage implementations
│   ├── websocket/       # WebSocket hubs
│   └── workers/         # Background workers
├── interfaces/           # Interface adapters layer
│   └── api/
│       ├── handlers/    # HTTP handlers
│       ├── middleware/  # HTTP middlewares
│       ├── routes/      # Route definitions
│       └── websocket/   # WebSocket handlers
└── pkg/                  # Shared packages
    ├── config/          # Configuration
    ├── database/        # Database utilities
    ├── logger/          # Logging
    ├── metrics/         # Metrics collection
    ├── health/          # Health checks
    └── middleware/      # Reusable middlewares
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15+
- Redis 7+
- (Optional) Docker & Docker Compose

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd gofiber-backend
   ```

2. **Setup environment**
   ```bash
   make setup-env
   # Or manually:
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install dependencies**
   ```bash
   go mod download
   make install-tools
   ```

4. **Run database migrations**
   ```bash
   # Migrations run automatically on startup
   # Or manually run migration files in migrations/
   ```

5. **Start the application**
   ```bash
   # Development mode with hot reload
   make dev

   # Or run directly
   make run

   # Or build and run
   make build
   ./bin/api
   ```

6. **Access the application**
   - API: http://localhost:3000
   - Swagger UI: http://localhost:3000/swagger/index.html
   - Health Check: http://localhost:3000/health
   - Metrics: http://localhost:3000/metrics

### Docker Development

```bash
# Start all services (app + postgres + redis)
docker-compose -f docker-compose.dev.yml up -d

# With management tools (pgAdmin + RedisInsight)
docker-compose -f docker-compose.dev.yml --profile tools up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop services
docker-compose -f docker-compose.dev.yml down
```

**Management Tools:**
- pgAdmin: http://localhost:5050 (admin@example.com / admin)
- RedisInsight: http://localhost:8001

### Production Deployment

```bash
# Using Docker Compose
docker-compose up -d

# Or build and run manually
make ci-build
./bin/api
```

See [DOCKER.md](DOCKER.md) for detailed deployment instructions.

## 📚 Documentation

Comprehensive guides available:

- **[ENVIRONMENT.md](ENVIRONMENT.md)** - Environment configuration
- **[DOCKER.md](DOCKER.md)** - Docker deployment guide
- **[SHUTDOWN.md](SHUTDOWN.md)** - Graceful shutdown documentation
- **[docs/README.md](docs/README.md)** - API documentation guide
- **[.github/README.md](.github/README.md)** - CI/CD documentation
- **[CODEBASE_ANALYSIS.md](CODEBASE_ANALYSIS.md)** - Architecture analysis
- **[IMPROVEMENT_ROADMAP.md](IMPROVEMENT_ROADMAP.md)** - Development roadmap

## 📡 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/auth/google` - Google OAuth
- `GET /api/v1/auth/google/callback` - OAuth callback

### Users
- `GET /api/v1/users/profile` - Get current user profile
- `PUT /api/v1/users/profile` - Update profile
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users/:id/follow` - Follow user
- `DELETE /api/v1/users/:id/follow` - Unfollow user

### Posts
- `GET /api/v1/posts` - List posts (paginated)
- `POST /api/v1/posts` - Create post
- `GET /api/v1/posts/:id` - Get post by ID
- `PUT /api/v1/posts/:id` - Update post
- `DELETE /api/v1/posts/:id` - Delete post
- `POST /api/v1/posts/:id/vote` - Vote on post
- `POST /api/v1/posts/:id/save` - Save post

### Comments
- `GET /api/v1/posts/:id/comments` - List comments
- `POST /api/v1/posts/:id/comments` - Create comment
- `PUT /api/v1/comments/:id` - Update comment
- `DELETE /api/v1/comments/:id` - Delete comment
- `POST /api/v1/comments/:id/vote` - Vote on comment

### Real-time
- `WS /ws/chat` - WebSocket chat
- `WS /ws/notifications` - WebSocket notifications

### Monitoring
- `GET /health` - Health check
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe
- `GET /metrics` - Application metrics

**Full API documentation:** http://localhost:3000/swagger/index.html

## 🧪 Testing

### Run Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests
make test-integration

# With coverage
make test-coverage

# With race detection
make test-race

# Specific packages
make test-services      # Service layer
make test-repositories  # Repository layer
make test-handlers      # Handler layer
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
make test-coverage
# Open coverage.html in browser

# Check coverage percentage
make coverage-check
```

### Current Test Status

- ✅ 86 core tests passing
- ✅ 7 compression tests passing
- ✅ **Total: 93 tests**

**Test Breakdown:**
- 23 unit tests (service layer)
- 26 integration tests (repository layer)
- 17 security/validation tests
- 12 cache tests
- 8 monitoring tests
- 7 compression tests

## 🔒 Security Features

- ✅ **Rate Limiting** - IP-based, 3 tiers (global, auth, strict)
- ✅ **Security Headers** - 8 headers (XSS, CSP, HSTS, etc.)
- ✅ **Input Validation** - Custom validators for all inputs
- ✅ **JWT Authentication** - Secure token-based auth
- ✅ **CORS Configuration** - Configurable allowed origins
- ✅ **SQL Injection Prevention** - Parameterized queries
- ✅ **Password Hashing** - bcrypt with salt
- ✅ **Security Scanning** - Automated with Gosec & Trivy

## 📊 Monitoring & Observability

### Health Checks
```bash
# Application health
curl http://localhost:3000/health

# Liveness probe (K8s)
curl http://localhost:3000/health/live

# Readiness probe (K8s)
curl http://localhost:3000/health/ready
```

### Metrics
```bash
# View metrics
curl http://localhost:3000/metrics

# Using make command
make metrics
```

**Available Metrics:**
- Total requests
- Successful/failed requests
- Response times (min/max/avg)
- Status code distribution
- Error count
- Uptime

### Logging

Structured logging with zerolog:
- Development: Pretty console output
- Production: JSON format
- Request/response logging
- Error tracking
- Performance metrics

## 🚀 Performance

### Optimizations Implemented

- ✅ **Response Compression** - Gzip compression (70-80% size reduction)
- ✅ **Database Indexing** - 30+ strategic indexes
- ✅ **Connection Pooling** - Optimized DB connection pool
- ✅ **Caching** - In-memory cache with TTL
- ✅ **Query Optimization** - Preloading, batch operations
- ✅ **Graceful Shutdown** - Clean resource cleanup

### Benchmarks

```bash
# Run benchmarks
make test-benchmark

# Load testing
make load-test

# Stress testing
make stress-test
```

## 🐳 Deployment

### Docker

See [DOCKER.md](DOCKER.md) for comprehensive Docker deployment guide.

**Quick Deploy:**
```bash
# Production
docker-compose up -d

# Development
docker-compose -f docker-compose.dev.yml up -d
```

### Kubernetes

```bash
# Convert docker-compose to k8s manifests
kompose convert -f docker-compose.yml

# Apply manifests
kubectl apply -f .
```

### Manual Deployment

```bash
# Build for production
make ci-build

# Run migrations
# (Automatic on startup)

# Start application
./bin/api
```

## 🔧 Development

### Available Make Commands

```bash
make help           # Show all commands

# Development
make dev            # Run with hot reload
make run            # Run directly
make build          # Build binary

# Testing
make test           # Run all tests
make test-coverage  # Run with coverage
make test-race      # Run with race detection

# Code Quality
make lint           # Run linter
make format         # Format code
make vet            # Run go vet

# Documentation
make swagger        # Generate Swagger docs
make docs           # Start Go docs server

# Database
make db-backup      # Backup database
make db-restore     # Restore database
make db-reset       # Reset database

# Docker
make docker-build   # Build Docker image
make docker-run     # Run in Docker

# Monitoring
make health         # Check health
make metrics        # View metrics

# Utilities
make version        # Show version info
make stats          # Show project statistics
```

## 🤝 Contributing

Contributions are welcome! Please see our [Pull Request Template](.github/PULL_REQUEST_TEMPLATE.md).

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Standards

- Follow Go best practices
- Write tests for new features
- Update documentation
- Run linter before committing (`make lint`)
- Ensure all tests pass (`make test`)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **Your Name** - *Initial work*

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Amazing Go web framework
- [GORM](https://gorm.io/) - Fantastic ORM
- All open-source contributors

## 📞 Support

- 📧 Email: support@example.com
- 🐛 Issues: [GitHub Issues](../../issues)
- 📖 Docs: [Documentation](#-documentation)

---

**Made with ❤️ using Go Fiber**

🤖 Generated with [Claude Code](https://claude.com/claude-code)
