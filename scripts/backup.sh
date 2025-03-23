#!/bin/bash
# Backup script for Wave Capacitor

# Default values
BACKUP_DIR="./backups"
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-5433}
DB_USER=${DB_USER:-"yugabyte"}
DB_NAME=${DB_NAME:-"wave"}
DB_PASSWORD=${DB_PASSWORD:-"yugabyte"}

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate timestamp for the backup file
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/wave_backup_$TIMESTAMP.sql"

# Create backup using pg_dump (works with YugabyteDB too)
echo "Creating database backup to $BACKUP_FILE..."
PGPASSWORD="$DB_PASSWORD" pg_dump \
  -h "$DB_HOST" \
  -p "$DB_PORT" \
  -U "$DB_USER" \
  -d "$DB_NAME" \
  -F c \
  -b \
  -v \
  -f "$BACKUP_FILE"

# Check if backup was successful
if [ $? -eq 0 ]; then
  echo "Backup completed successfully"
  
  # Compress the backup file
  gzip -f "$BACKUP_FILE"
  echo "Backup compressed: ${BACKUP_FILE}.gz"
  
  # List available backups
  echo "Available backups:"
  ls -lh "$BACKUP_DIR"/*.gz
else
  echo "Backup failed"
  exit 1
fi

# Remove backups older than 7 days
echo "Removing backups older than 7 days..."
find "$BACKUP_DIR" -name "wave_backup_*.sql.gz" -type f -mtime +7 -delete

echo "Backup process completed"