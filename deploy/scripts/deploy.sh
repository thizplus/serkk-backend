#!/bin/bash

# GoFiber API Deployment Script
# Usage: ./deploy.sh

set -e  # Exit on error

echo "🚀 Starting deployment..."

# Configuration
APP_DIR="/home/gofiber/serkk-backend"
SERVICE_NAME="gofiber-api"
BRANCH="main"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if running as correct user
if [ "$USER" != "gofiber" ]; then
    log_error "This script must be run as 'gofiber' user"
    exit 1
fi

# Navigate to app directory
cd $APP_DIR

# 1. Pull latest code
log_info "Pulling latest code from GitHub..."
git fetch origin
git pull origin $BRANCH

# 2. Check Go version
log_info "Checking Go version..."
go version

# 3. Download dependencies
log_info "Downloading dependencies..."
go mod download
go mod verify

# 4. Run tests (optional, comment out if tests take too long)
log_info "Running tests..."
if go test ./... -timeout 30s; then
    log_info "Tests passed ✓"
else
    log_warning "Tests failed, continuing anyway..."
fi

# 5. Build application
log_info "Building application..."
go build -o bin/api cmd/api/main.go

# Check if build was successful
if [ ! -f "bin/api" ]; then
    log_error "Build failed - binary not found"
    exit 1
fi

log_info "Build successful ✓"

# 6. Restart service
log_info "Restarting service..."
sudo systemctl restart $SERVICE_NAME

# Wait for service to start
sleep 3

# 7. Check service status
if sudo systemctl is-active --quiet $SERVICE_NAME; then
    log_info "Service is running ✓"
else
    log_error "Service failed to start"
    sudo journalctl -u $SERVICE_NAME -n 20
    exit 1
fi

# 8. Health check
log_info "Performing health check..."
sleep 2

if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_info "Health check passed ✓"
else
    log_error "Health check failed"
    exit 1
fi

# 9. Show recent logs
log_info "Recent logs:"
sudo journalctl -u $SERVICE_NAME -n 10 --no-pager

echo ""
log_info "🎉 Deployment completed successfully!"
echo ""
log_info "Service status:"
sudo systemctl status $SERVICE_NAME --no-pager -l

echo ""
log_info "To view logs: sudo journalctl -u $SERVICE_NAME -f"
