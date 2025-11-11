#!/bin/bash

# Database Backup Script for GoFiber API
# Usage: ./backup-db.sh

set -e  # Exit on error

# Configuration
DB_NAME="gofiber_db"
DB_USER="gofiber_user"
DB_PASSWORD="your_secure_password_here"  # CHANGE THIS!
BACKUP_DIR="/home/gofiber/backups/database"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/gofiber_db_$DATE.sql.gz"
RETENTION_DAYS=7

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create backup directory if not exists
mkdir -p $BACKUP_DIR

log_info "Starting database backup..."

# Backup database
if PGPASSWORD=$DB_PASSWORD pg_dump -h localhost -U $DB_USER $DB_NAME | gzip > $BACKUP_FILE; then
    log_info "Backup completed: $BACKUP_FILE"

    # Get backup file size
    SIZE=$(du -h $BACKUP_FILE | cut -f1)
    log_info "Backup size: $SIZE"
else
    log_error "Backup failed!"
    exit 1
fi

# Delete old backups
log_info "Cleaning up old backups (older than $RETENTION_DAYS days)..."
DELETED=$(find $BACKUP_DIR -name "gofiber_db_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete -print | wc -l)
log_info "Deleted $DELETED old backup(s)"

# List recent backups
log_info "Recent backups:"
ls -lht $BACKUP_DIR | head -n 6

log_info "✓ Backup completed successfully!"
