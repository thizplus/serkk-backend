# 🛑 Graceful Shutdown Guide

This document explains how the application handles graceful shutdown to ensure no data loss and clean resource cleanup.

## 📋 Overview

Graceful shutdown ensures that:
1. Server stops accepting new connections
2. Existing connections are allowed to finish
3. WebSocket connections are closed properly
4. Background workers complete their tasks
5. Database connections are closed cleanly
6. Redis connections are terminated
7. All resources are released

## 🔧 How It Works

### Shutdown Sequence

```
SIGTERM/SIGINT received
        ↓
Stop accepting new requests
        ↓
Wait for active requests to complete
        ↓
Shutdown Fiber server
        ↓
Stop WebSocket hubs (Chat & Notifications)
        ↓
Stop background workers (VideoEncoder)
        ↓
Stop event scheduler
        ↓
Close Redis connection
        ↓
Close database connection
        ↓
Exit application
```

### Implementation

The shutdown process is implemented in `cmd/api/main.go`:

```go
// Start server in goroutine
go func() {
    if err := app.Listen(":" + port); err != nil {
        log.Fatalf("❌ Server error: %v", err)
    }
}()

// Wait for interrupt signal
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
<-quit

// Graceful shutdown
log.Println("\n🛑 Gracefully shutting down...")
if err := app.Shutdown(); err != nil {
    log.Printf("❌ Error shutting down server: %v", err)
}

// Cleanup resources
if err := container.Cleanup(); err != nil {
    log.Printf("❌ Error during cleanup: %v", err)
}

log.Println("👋 Shutdown complete")
```

### Cleanup Process

The `container.Cleanup()` method handles resource cleanup:

```go
func (c *Container) Cleanup() error {
    log.Println("Starting cleanup...")

    // 1. Stop VideoEncoderWorker
    if c.VideoEncoderWorker != nil {
        c.VideoEncoderWorker.Stop()
        log.Println("✓ VideoEncoderWorker stopped")
    }

    // 2. Stop ChatHub
    if c.ChatHub != nil {
        c.ChatHub.Stop()
        log.Println("✓ ChatHub stopped")
    }

    // 3. Stop NotificationHub
    if c.NotificationHub != nil {
        c.NotificationHub.Stop()
        log.Println("✓ NotificationHub stopped")
    }

    // 4. Stop EventScheduler
    if c.EventScheduler != nil {
        if c.EventScheduler.IsRunning() {
            c.EventScheduler.Stop()
            log.Println("✓ Event scheduler stopped")
        }
    }

    // 5. Close Redis connection
    if c.RedisClient != nil {
        if err := c.RedisClient.Close(); err != nil {
            log.Printf("Warning: Failed to close Redis: %v", err)
        } else {
            log.Println("✓ Redis connection closed")
        }
    }

    // 6. Close database connection
    if c.DB != nil {
        sqlDB, err := c.DB.DB()
        if err == nil {
            if err := sqlDB.Close(); err != nil {
                log.Printf("Warning: Failed to close database: %v", err)
            } else {
                log.Println("✓ Database connection closed")
            }
        }
    }

    log.Println("✓ Cleanup completed")
    return nil
}
```

## 🚀 Triggering Shutdown

### Manual Shutdown

```bash
# Using Ctrl+C
# Sends SIGINT signal

# Using kill command
kill -TERM <pid>
```

### Docker/Kubernetes

```bash
# Docker stop (sends SIGTERM, waits 10s, then SIGKILL)
docker stop <container>

# Docker stop with custom timeout
docker stop -t 30 <container>

# Kubernetes pod termination
kubectl delete pod <pod-name>
# Sends SIGTERM, waits 30s (default terminationGracePeriodSeconds)
```

### Systemd

```ini
[Service]
Type=simple
ExecStart=/app/api
ExecStop=/bin/kill -TERM $MAINPID
TimeoutStopSec=30s
KillMode=mixed
```

## ⏱️ Timeouts

### Server Shutdown Timeout

Fiber's `Shutdown()` method waits indefinitely by default. For production, you may want to add a timeout:

```go
import (
    "context"
    "time"
)

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Shutdown with timeout
if err := app.ShutdownWithContext(ctx); err != nil {
    log.Printf("❌ Forced shutdown after timeout: %v", err)
}
```

### Docker Timeout

```yaml
# docker-compose.yml
services:
  app:
    stop_grace_period: 30s  # Wait 30s before force killing
```

### Kubernetes Timeout

```yaml
# pod.yaml
spec:
  containers:
  - name: app
    terminationGracePeriodSeconds: 30  # Wait 30s before force killing
```

## 📊 Monitoring Shutdown

### Log Output

Normal shutdown logs:
```
🛑 Gracefully shutting down...
Starting cleanup...
✓ VideoEncoderWorker stopped
✓ ChatHub stopped
✓ NotificationHub stopped
✓ Event scheduler stopped
✓ Redis connection closed
✓ Database connection closed
✓ Cleanup completed
👋 Shutdown complete
```

### Health Check During Shutdown

The `/health/ready` endpoint can be used to check if the application is ready for shutdown:

```bash
# Check if app is ready
curl http://localhost:3000/health/ready

# Response when shutting down
{
  "status": "not_ready",
  "checks": {
    "database": {
      "status": "unhealthy",
      "message": "Connection closed"
    }
  }
}
```

## 🐛 Troubleshooting

### Shutdown Hangs

**Possible Causes:**
1. Long-running requests not completing
2. WebSocket connections not closing
3. Database transactions not committed
4. Redis operations stuck

**Solutions:**
```bash
# Force kill after timeout (last resort)
kill -KILL <pid>

# Check active connections
lsof -i :3000

# Check database connections
SELECT * FROM pg_stat_activity WHERE application_name = 'gofiber-template';
```

### Resources Not Released

**Check for leaks:**
```bash
# Memory leaks
pprof http://localhost:3000/debug/pprof/heap

# Goroutine leaks
pprof http://localhost:3000/debug/pprof/goroutine

# Database connections
SELECT count(*) FROM pg_stat_activity;
```

### Incomplete Cleanup

**Check logs for errors:**
```bash
# Filter cleanup errors
docker logs <container> 2>&1 | grep "cleanup"

# Kubernetes logs
kubectl logs <pod> | grep "cleanup"
```

## 🔒 Best Practices

### 1. Set Appropriate Timeouts

```go
// Application-level timeout
const shutdownTimeout = 30 * time.Second

ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
defer cancel()

app.ShutdownWithContext(ctx)
```

### 2. Handle In-Flight Requests

```go
// Use context in long-running operations
func (s *Service) ProcessLongTask(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()  // Return if context cancelled
    default:
        // Continue processing
    }
}
```

### 3. Graceful WebSocket Closure

```go
// Send close message before shutdown
hub.BroadcastToAll(&Message{
    Type: "server_shutdown",
    Data: "Server shutting down, please reconnect",
})

time.Sleep(1 * time.Second)  // Allow clients to receive
hub.Stop()
```

### 4. Database Transaction Handling

```go
// Always use transactions with proper cleanup
err := database.WithTransaction(db, func(tx *gorm.DB) error {
    // Your database operations
    return nil
})
```

### 5. Testing Shutdown

```bash
# Start application
go run ./cmd/api &
APP_PID=$!

# Wait for startup
sleep 2

# Trigger shutdown
kill -TERM $APP_PID

# Check if shutdown completed
wait $APP_PID
echo "Exit code: $?"  # Should be 0
```

## 📝 Shutdown Checklist

Before deploying to production, verify:

- [ ] Graceful shutdown implemented
- [ ] All resources properly cleaned up
- [ ] Timeouts configured appropriately
- [ ] WebSocket connections close gracefully
- [ ] Background workers stop cleanly
- [ ] Database connections close without errors
- [ ] Redis connections close properly
- [ ] No goroutine leaks
- [ ] No memory leaks
- [ ] Shutdown logs are clear
- [ ] Docker/K8s grace period configured
- [ ] Health checks work during shutdown
- [ ] Tested with real traffic

## 🚀 Production Configuration

### Docker

```dockerfile
# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Set stop signal
STOPSIGNAL SIGTERM
```

### Kubernetes

```yaml
apiVersion: v1
kind: Pod
spec:
  terminationGracePeriodSeconds: 30
  containers:
  - name: app
    lifecycle:
      preStop:
        exec:
          command: ["/bin/sh", "-c", "sleep 5"]
    livenessProbe:
      httpGet:
        path: /health/live
        port: 3000
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 3000
```

## 📚 Resources

- [Fiber Graceful Shutdown](https://docs.gofiber.io/api/app/#shutdown)
- [Go Signal Handling](https://gobyexample.com/signals)
- [Docker Stop Signal](https://docs.docker.com/engine/reference/commandline/stop/)
- [Kubernetes Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/)
- [Systemd Service Management](https://www.freedesktop.org/software/systemd/man/systemd.service.html)

## 🆘 Getting Help

If graceful shutdown is not working:

1. Check logs for error messages
2. Verify signal handling is working: `kill -TERM <pid>`
3. Check for blocking operations
4. Monitor resource cleanup
5. Test with short timeout to identify issues
6. Use pprof to detect goroutine leaks
7. Create GitHub issue with logs and environment details
