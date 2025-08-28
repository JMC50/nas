# 💾 Backup & Restore Guide

Comprehensive backup and disaster recovery procedures for the NAS File Manager.

## 📋 Table of Contents

- [Backup Overview](#backup-overview)
- [Backup Types](#backup-types)
- [Automated Backup Setup](#automated-backup-setup)
- [Manual Backup Procedures](#manual-backup-procedures)
- [Restore Procedures](#restore-procedures)
- [Disaster Recovery](#disaster-recovery)
- [Backup Verification](#backup-verification)
- [Offsite Backup](#offsite-backup)

## Backup Overview

### What to Backup

| Component | Priority | Frequency | Size | Recovery Time |
|-----------|----------|-----------|------|---------------|
| **User Data** | Critical | Daily | Large | 1-2 hours |
| **Database** | Critical | Hourly | Small | 5-10 minutes |
| **Configuration** | High | Weekly | Tiny | 1-2 minutes |
| **Application Code** | Medium | On changes | Small | 5-10 minutes |
| **System Config** | Medium | Monthly | Small | 10-30 minutes |

### Backup Strategy

**3-2-1 Rule Implementation:**
- **3 copies** of important data
- **2 different storage types** (local + cloud/remote)
- **1 offsite backup** for disaster recovery

### Retention Policy

| Backup Type | Keep Daily | Keep Weekly | Keep Monthly | Keep Yearly |
|-------------|------------|-------------|--------------|-------------|
| **User Data** | 7 days | 4 weeks | 12 months | 3 years |
| **Database** | 30 days | 8 weeks | 12 months | 3 years |
| **Configuration** | 30 days | 12 weeks | 24 months | 5 years |

## Backup Types

### Full Backup

Complete backup of all NAS data and configuration.

```bash
#!/bin/bash
# full-backup.sh

BACKUP_ROOT="/backup"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="$BACKUP_ROOT/full-backup-$TIMESTAMP"
SOURCE_DATA="/mnt/nas-storage"
CONFIG_DIR="/home/heesung/NAS"

echo "🔄 Starting full backup to $BACKUP_DIR"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Stop services for consistency
echo "Stopping NAS service..."
systemctl stop nas-app.service

# Backup user data
echo "Backing up user data..."
rsync -av --progress "$SOURCE_DATA/data/" "$BACKUP_DIR/data/"

# Backup admin data
echo "Backing up admin data..."
rsync -av --progress "$SOURCE_DATA/admin-data/" "$BACKUP_DIR/admin-data/"

# Backup database
echo "Backing up database..."
cp -r "$SOURCE_DATA/database" "$BACKUP_DIR/"

# Backup configuration
echo "Backing up configuration..."
cp "$CONFIG_DIR/.env" "$BACKUP_DIR/"
cp "$CONFIG_DIR/nas-app.service" "$BACKUP_DIR/"
cp -r "$CONFIG_DIR/scripts" "$BACKUP_DIR/"

# Create backup manifest
echo "Creating backup manifest..."
cat > "$BACKUP_DIR/backup-manifest.txt" << EOF
NAS Full Backup Manifest
========================
Backup Date: $(date)
Backup Type: Full
Source: $SOURCE_DATA
Destination: $BACKUP_DIR

Components:
- User Data: $(du -sh "$BACKUP_DIR/data" | cut -f1)
- Admin Data: $(du -sh "$BACKUP_DIR/admin-data" | cut -f1)
- Database: $(du -sh "$BACKUP_DIR/database" | cut -f1)
- Configuration: $(ls -la "$BACKUP_DIR"/.env "$BACKUP_DIR"/nas-app.service 2>/dev/null | wc -l) files

Total Size: $(du -sh "$BACKUP_DIR" | cut -f1)
EOF

# Start services
echo "Starting NAS service..."
systemctl start nas-app.service

# Compress backup
echo "Compressing backup..."
cd "$BACKUP_ROOT"
tar -czf "full-backup-$TIMESTAMP.tar.gz" "$(basename "$BACKUP_DIR")"
rm -rf "$BACKUP_DIR"

# Verify compressed backup
if [ -f "full-backup-$TIMESTAMP.tar.gz" ]; then
    BACKUP_SIZE=$(du -sh "full-backup-$TIMESTAMP.tar.gz" | cut -f1)
    echo "✅ Full backup completed: full-backup-$TIMESTAMP.tar.gz ($BACKUP_SIZE)"
else
    echo "❌ Backup failed"
    exit 1
fi
```

### Incremental Backup

Only backs up changes since last backup.

```bash
#!/bin/bash
# incremental-backup.sh

BACKUP_ROOT="/backup"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
INCREMENTAL_DIR="$BACKUP_ROOT/incremental"
SOURCE_DATA="/mnt/nas-storage"
SNAPSHOT_FILE="$INCREMENTAL_DIR/snapshot.file"

echo "🔄 Starting incremental backup..."

# Create incremental directory
mkdir -p "$INCREMENTAL_DIR"

# Find last full backup for reference
LAST_FULL=$(find "$BACKUP_ROOT" -name "full-backup-*.tar.gz" -type f -printf "%T@ %p\n" | sort -n | tail -1 | cut -d' ' -f2-)

if [ -z "$LAST_FULL" ]; then
    echo "❌ No full backup found. Run full backup first."
    exit 1
fi

# Create incremental backup using rsync
rsync -av --compare-dest="$(dirname "$LAST_FULL")" \
    --link-dest="$INCREMENTAL_DIR/latest" \
    "$SOURCE_DATA/" \
    "$INCREMENTAL_DIR/incremental-$TIMESTAMP/"

# Update latest link
rm -f "$INCREMENTAL_DIR/latest"
ln -sf "incremental-$TIMESTAMP" "$INCREMENTAL_DIR/latest"

# Create manifest
cat > "$INCREMENTAL_DIR/incremental-$TIMESTAMP/manifest.txt" << EOF
NAS Incremental Backup Manifest
===============================
Backup Date: $(date)
Backup Type: Incremental
Base Backup: $(basename "$LAST_FULL")
Source: $SOURCE_DATA
Destination: $INCREMENTAL_DIR/incremental-$TIMESTAMP

Size: $(du -sh "$INCREMENTAL_DIR/incremental-$TIMESTAMP" | cut -f1)
EOF

echo "✅ Incremental backup completed: incremental-$TIMESTAMP"
```

### Database-Only Backup

Quick database backup for frequent snapshots.

```bash
#!/bin/bash
# db-backup.sh

DB_SOURCE="/mnt/nas-storage/database"
BACKUP_DIR="/backup/database"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RETENTION_DAYS=30

echo "🗄️ Starting database backup..."

mkdir -p "$BACKUP_DIR"

# Hot backup using SQLite backup API
sqlite3 "$DB_SOURCE/nas.sqlite" ".backup $BACKUP_DIR/nas-backup-$TIMESTAMP.sqlite"

# Verify backup
if sqlite3 "$BACKUP_DIR/nas-backup-$TIMESTAMP.sqlite" "PRAGMA integrity_check;" | grep -q "ok"; then
    echo "✅ Database backup completed: nas-backup-$TIMESTAMP.sqlite"
    
    # Compress backup
    gzip "$BACKUP_DIR/nas-backup-$TIMESTAMP.sqlite"
    
    # Clean old backups
    find "$BACKUP_DIR" -name "nas-backup-*.sqlite.gz" -mtime +$RETENTION_DAYS -delete
    
    echo "📊 Database backup statistics:"
    echo "  - Backup size: $(du -sh "$BACKUP_DIR/nas-backup-$TIMESTAMP.sqlite.gz" | cut -f1)"
    echo "  - Total backups: $(ls "$BACKUP_DIR"/nas-backup-*.sqlite.gz | wc -l)"
else
    echo "❌ Database backup verification failed"
    rm -f "$BACKUP_DIR/nas-backup-$TIMESTAMP.sqlite"
    exit 1
fi
```

## Automated Backup Setup

### Backup Scheduling

```bash
#!/bin/bash
# setup-backup-schedule.sh

echo "⏰ Setting up automated backup schedule..."

# Create backup directories
mkdir -p /backup/{full,incremental,database,config}

# Setup cron jobs for automated backups
(crontab -l 2>/dev/null; cat << 'EOF'
# NAS Backup Schedule

# Database backup every hour
0 * * * * /opt/backup-scripts/db-backup.sh >> /var/log/backup.log 2>&1

# Incremental backup every 6 hours
0 */6 * * * /opt/backup-scripts/incremental-backup.sh >> /var/log/backup.log 2>&1

# Full backup weekly on Sunday at 2 AM
0 2 * * 0 /opt/backup-scripts/full-backup.sh >> /var/log/backup.log 2>&1

# Configuration backup daily at 3 AM
0 3 * * * /opt/backup-scripts/config-backup.sh >> /var/log/backup.log 2>&1

# Backup cleanup weekly on Saturday at 4 AM
0 4 * * 6 /opt/backup-scripts/cleanup-old-backups.sh >> /var/log/backup.log 2>&1

# Backup verification daily at 5 AM
0 5 * * * /opt/backup-scripts/verify-backups.sh >> /var/log/backup.log 2>&1
EOF
) | crontab -

echo "✅ Backup schedule configured"
```

### Configuration Backup

```bash
#!/bin/bash
# config-backup.sh

CONFIG_BACKUP_DIR="/backup/config"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
NAS_DIR="/home/heesung/NAS"

echo "⚙️ Backing up configuration..."

mkdir -p "$CONFIG_BACKUP_DIR"

# Create configuration backup
tar -czf "$CONFIG_BACKUP_DIR/config-backup-$TIMESTAMP.tar.gz" \
    -C "$NAS_DIR" \
    .env \
    nas-app.service \
    scripts/ \
    package.json \
    docker-compose.yml

# System configuration
tar -czf "$CONFIG_BACKUP_DIR/system-config-$TIMESTAMP.tar.gz" \
    /etc/systemd/system/nas-app.service \
    /etc/nginx/sites-available/nas-app \
    /etc/logrotate.d/nas-app

# Backup verification
if [ -f "$CONFIG_BACKUP_DIR/config-backup-$TIMESTAMP.tar.gz" ]; then
    echo "✅ Configuration backup completed"
else
    echo "❌ Configuration backup failed"
    exit 1
fi

# Clean old config backups (keep 30 days)
find "$CONFIG_BACKUP_DIR" -name "config-backup-*.tar.gz" -mtime +30 -delete
find "$CONFIG_BACKUP_DIR" -name "system-config-*.tar.gz" -mtime +30 -delete
```

### Backup Cleanup

```bash
#!/bin/bash
# cleanup-old-backups.sh

BACKUP_ROOT="/backup"
LOG_FILE="/var/log/backup-cleanup.log"

echo "$(date): Starting backup cleanup..." >> "$LOG_FILE"

# Full backups - keep 4 weekly backups
FULL_BACKUPS_TO_KEEP=4
find "$BACKUP_ROOT" -name "full-backup-*.tar.gz" -type f -printf "%T@ %p\n" | \
    sort -rn | tail -n +$((FULL_BACKUPS_TO_KEEP + 1)) | cut -d' ' -f2- | \
    while read backup_file; do
        echo "$(date): Removing old full backup: $(basename "$backup_file")" >> "$LOG_FILE"
        rm -f "$backup_file"
    done

# Database backups - keep 30 days
find "$BACKUP_ROOT/database" -name "nas-backup-*.sqlite.gz" -mtime +30 -exec rm -f {} \;
REMOVED_DB=$(find "$BACKUP_ROOT/database" -name "nas-backup-*.sqlite.gz" -mtime +30 | wc -l)
echo "$(date): Removed $REMOVED_DB old database backups" >> "$LOG_FILE"

# Incremental backups - keep 7 days
find "$BACKUP_ROOT/incremental" -name "incremental-*" -type d -mtime +7 -exec rm -rf {} \;

# Configuration backups - keep 90 days
find "$BACKUP_ROOT/config" -name "*-backup-*.tar.gz" -mtime +90 -exec rm -f {} \;

# Report cleanup results
CURRENT_USAGE=$(du -sh "$BACKUP_ROOT" | cut -f1)
echo "$(date): Backup cleanup completed. Current usage: $CURRENT_USAGE" >> "$LOG_FILE"
```

## Manual Backup Procedures

### Emergency Backup

```bash
#!/bin/bash
# emergency-backup.sh

echo "🚨 Emergency backup procedure starting..."

EMERGENCY_BACKUP="/backup/emergency-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$EMERGENCY_BACKUP"

# Stop service immediately
systemctl stop nas-app.service

# Backup critical data only
echo "Backing up critical data..."

# Database (highest priority)
cp -r /mnt/nas-storage/database "$EMERGENCY_BACKUP/"

# Configuration
cp /home/heesung/NAS/.env "$EMERGENCY_BACKUP/"

# Recent user data (last 7 days)
find /mnt/nas-storage/data -mtime -7 -type f -exec cp --parents {} "$EMERGENCY_BACKUP/" \;

# Create emergency manifest
echo "Emergency backup created: $(date)" > "$EMERGENCY_BACKUP/emergency-manifest.txt"
echo "Database: $(du -sh "$EMERGENCY_BACKUP/database" | cut -f1)" >> "$EMERGENCY_BACKUP/emergency-manifest.txt"
echo "Recent files: $(find "$EMERGENCY_BACKUP" -type f | wc -l)" >> "$EMERGENCY_BACKUP/emergency-manifest.txt"

# Compress for transport
tar -czf "${EMERGENCY_BACKUP}.tar.gz" -C "$(dirname "$EMERGENCY_BACKUP")" "$(basename "$EMERGENCY_BACKUP")"
rm -rf "$EMERGENCY_BACKUP"

# Restart service
systemctl start nas-app.service

echo "✅ Emergency backup completed: ${EMERGENCY_BACKUP}.tar.gz"
echo "💾 Size: $(du -sh "${EMERGENCY_BACKUP}.tar.gz" | cut -f1)"
```

### Selective Backup

```bash
#!/bin/bash
# selective-backup.sh

echo "🎯 Selective backup utility"
echo "Usage: selective-backup.sh [user|admin|database|config] [destination]"

BACKUP_TYPE="$1"
DESTINATION="${2:-/backup/selective}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

case "$BACKUP_TYPE" in
    "user")
        echo "Backing up user data..."
        rsync -av --progress /mnt/nas-storage/data/ "$DESTINATION/user-data-$TIMESTAMP/"
        ;;
    "admin")
        echo "Backing up admin data..."
        rsync -av --progress /mnt/nas-storage/admin-data/ "$DESTINATION/admin-data-$TIMESTAMP/"
        ;;
    "database")
        echo "Backing up database..."
        sqlite3 /mnt/nas-storage/database/nas.sqlite ".backup $DESTINATION/database-$TIMESTAMP.sqlite"
        ;;
    "config")
        echo "Backing up configuration..."
        tar -czf "$DESTINATION/config-$TIMESTAMP.tar.gz" -C /home/heesung/NAS .env nas-app.service
        ;;
    *)
        echo "❌ Invalid backup type. Use: user|admin|database|config"
        exit 1
        ;;
esac

echo "✅ Selective backup completed"
```

## Restore Procedures

### Full System Restore

```bash
#!/bin/bash
# full-restore.sh

echo "🔄 Full system restore procedure"
echo "⚠️  WARNING: This will overwrite all current data!"
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Restore cancelled"
    exit 1
fi

BACKUP_FILE="$1"
RESTORE_DIR="/tmp/nas-restore-$$"

if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Stop all services
echo "Stopping services..."
systemctl stop nas-app.service

# Extract backup
echo "Extracting backup..."
mkdir -p "$RESTORE_DIR"
tar -xzf "$BACKUP_FILE" -C "$RESTORE_DIR"

# Find backup directory
BACKUP_CONTENT=$(find "$RESTORE_DIR" -type d -name "full-backup-*" | head -1)
if [ -z "$BACKUP_CONTENT" ]; then
    echo "❌ Invalid backup file structure"
    exit 1
fi

# Restore data
echo "Restoring user data..."
rm -rf /mnt/nas-storage/data
cp -r "$BACKUP_CONTENT/data" /mnt/nas-storage/

echo "Restoring admin data..."
rm -rf /mnt/nas-storage/admin-data  
cp -r "$BACKUP_CONTENT/admin-data" /mnt/nas-storage/

echo "Restoring database..."
rm -rf /mnt/nas-storage/database
cp -r "$BACKUP_CONTENT/database" /mnt/nas-storage/

# Restore configuration
echo "Restoring configuration..."
cp "$BACKUP_CONTENT/.env" /home/heesung/NAS/
cp "$BACKUP_CONTENT/nas-app.service" /home/heesung/NAS/

# Fix permissions
echo "Fixing permissions..."
chown -R heesung:heesung /home/heesung/NAS
chown -R heesung:heesung /mnt/nas-storage

# Start services
echo "Starting services..."
systemctl start nas-app.service

# Cleanup
rm -rf "$RESTORE_DIR"

# Verify restore
sleep 10
if curl -f -s http://localhost:7777/ > /dev/null; then
    echo "✅ Full restore completed successfully"
else
    echo "❌ Restore completed but application not responding"
fi
```

### Database Restore

```bash
#!/bin/bash
# database-restore.sh

BACKUP_FILE="$1"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Database backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "🗄️ Database restore procedure"
echo "Backup file: $BACKUP_FILE"

# Stop application
systemctl stop nas-app.service

# Backup current database
echo "Creating safety backup of current database..."
cp /mnt/nas-storage/database/nas.sqlite "/backup/pre-restore-backup-$TIMESTAMP.sqlite"

# Restore database
echo "Restoring database..."
if [[ "$BACKUP_FILE" == *.gz ]]; then
    gunzip -c "$BACKUP_FILE" > /mnt/nas-storage/database/nas.sqlite
else
    cp "$BACKUP_FILE" /mnt/nas-storage/database/nas.sqlite
fi

# Verify restored database
echo "Verifying restored database..."
if sqlite3 /mnt/nas-storage/database/nas.sqlite "PRAGMA integrity_check;" | grep -q "ok"; then
    echo "✅ Database integrity verified"
else
    echo "❌ Database restore failed - restoring original"
    cp "/backup/pre-restore-backup-$TIMESTAMP.sqlite" /mnt/nas-storage/database/nas.sqlite
    exit 1
fi

# Fix permissions
chown heesung:heesung /mnt/nas-storage/database/nas.sqlite

# Start application
systemctl start nas-app.service

echo "✅ Database restore completed"
```

## Disaster Recovery

### Disaster Recovery Plan

```bash
#!/bin/bash
# disaster-recovery.sh

echo "🆘 NAS Disaster Recovery Procedure"
echo "=================================="

# Step 1: Assessment
echo "Step 1: Damage Assessment"
echo "- Check hardware status"
echo "- Assess data corruption level"
echo "- Verify backup availability"
read -p "Press Enter to continue..."

# Step 2: Environment preparation
echo "Step 2: Preparing Recovery Environment"

# Install required software
echo "Installing required packages..."
apt-get update
apt-get install -y nodejs npm sqlite3 nginx

# Create user and directories
echo "Creating NAS user and directories..."
useradd -r -s /bin/bash heesung
mkdir -p /home/heesung/NAS
mkdir -p /mnt/nas-storage/{data,admin-data,database}
chown -R heesung:heesung /home/heesung/NAS /mnt/nas-storage

# Step 3: Application recovery
echo "Step 3: Application Recovery"

# Restore application code
if [ -f "/backup/config/config-backup-latest.tar.gz" ]; then
    tar -xzf "/backup/config/config-backup-latest.tar.gz" -C /home/heesung/NAS
    echo "✅ Application configuration restored"
else
    echo "❌ No configuration backup found - manual setup required"
fi

# Step 4: Data recovery
echo "Step 4: Data Recovery"

# Find latest full backup
LATEST_FULL_BACKUP=$(find /backup -name "full-backup-*.tar.gz" -type f -printf "%T@ %p\n" | sort -n | tail -1 | cut -d' ' -f2-)

if [ -n "$LATEST_FULL_BACKUP" ]; then
    echo "Restoring from: $(basename "$LATEST_FULL_BACKUP")"
    /opt/backup-scripts/full-restore.sh "$LATEST_FULL_BACKUP"
else
    echo "❌ No full backup found"
    exit 1
fi

# Step 5: Service setup
echo "Step 5: Service Configuration"

# Install systemd service
cp /home/heesung/NAS/nas-app.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable nas-app.service

# Start services
systemctl start nas-app.service

# Step 6: Verification
echo "Step 6: Recovery Verification"
sleep 15

if systemctl is-active --quiet nas-app.service && curl -f -s http://localhost:7777/ > /dev/null; then
    echo "✅ Disaster recovery completed successfully!"
    echo "🌐 Application available at: http://$(hostname):7777"
else
    echo "❌ Recovery verification failed"
    echo "Check logs: journalctl -u nas-app.service"
fi
```

### Recovery Time Objectives (RTO)

| Component | RTO Target | Steps Required |
|-----------|------------|----------------|
| Database | 15 minutes | Stop service → Restore DB → Start service |
| Configuration | 5 minutes | Copy config files → Restart |
| User Data | 2-4 hours | Full data restore from backup |
| Complete System | 4-6 hours | Full disaster recovery procedure |

### Recovery Point Objectives (RPO)

| Data Type | RPO Target | Backup Frequency |
|-----------|------------|------------------|
| Database | 1 hour | Hourly backups |
| User Files | 6 hours | Every 6 hours |
| Configuration | 24 hours | Daily backups |

## Backup Verification

### Automated Verification

```bash
#!/bin/bash
# verify-backups.sh

BACKUP_ROOT="/backup"
VERIFICATION_LOG="/var/log/backup-verification.log"
FAILED_BACKUPS=0

echo "$(date): Starting backup verification..." >> "$VERIFICATION_LOG"

# Verify full backups
echo "Verifying full backups..."
for backup in $(find "$BACKUP_ROOT" -name "full-backup-*.tar.gz" -mtime -7); do
    echo "Testing: $(basename "$backup")"
    
    if tar -tzf "$backup" > /dev/null 2>&1; then
        echo "$(date): PASS - $(basename "$backup")" >> "$VERIFICATION_LOG"
    else
        echo "$(date): FAIL - $(basename "$backup") - Archive corrupted" >> "$VERIFICATION_LOG"
        FAILED_BACKUPS=$((FAILED_BACKUPS + 1))
    fi
done

# Verify database backups
echo "Verifying database backups..."
for db_backup in $(find "$BACKUP_ROOT/database" -name "nas-backup-*.sqlite.gz" -mtime -1); do
    echo "Testing: $(basename "$db_backup")"
    
    # Extract and test
    temp_db="/tmp/test-$(basename "$db_backup" .gz)"
    gunzip -c "$db_backup" > "$temp_db"
    
    if sqlite3 "$temp_db" "PRAGMA integrity_check;" | grep -q "ok"; then
        echo "$(date): PASS - $(basename "$db_backup")" >> "$VERIFICATION_LOG"
    else
        echo "$(date): FAIL - $(basename "$db_backup") - Database corrupted" >> "$VERIFICATION_LOG"
        FAILED_BACKUPS=$((FAILED_BACKUPS + 1))
    fi
    
    rm -f "$temp_db"
done

# Report results
if [ $FAILED_BACKUPS -eq 0 ]; then
    echo "$(date): ✅ All backups verified successfully" >> "$VERIFICATION_LOG"
else
    echo "$(date): ❌ $FAILED_BACKUPS backup(s) failed verification" >> "$VERIFICATION_LOG"
    # Send alert
    echo "Backup verification failures detected" | mail -s "NAS Backup Alert" admin@example.com
fi

echo "Backup verification completed. Check $VERIFICATION_LOG for details."
```

## Offsite Backup

### Cloud Backup Setup

```bash
#!/bin/bash
# setup-cloud-backup.sh

echo "☁️  Setting up cloud backup..."

# Install rclone for cloud storage
curl https://rclone.org/install.sh | bash

# Configure cloud provider (example for AWS S3)
cat > /root/.config/rclone/rclone.conf << 'EOF'
[s3-backup]
type = s3
provider = AWS
access_key_id = YOUR_ACCESS_KEY
secret_access_key = YOUR_SECRET_KEY
region = us-east-1
EOF

chmod 600 /root/.config/rclone/rclone.conf

# Create cloud backup script
cat > /opt/backup-scripts/cloud-backup.sh << 'EOF'
#!/bin/bash
# cloud-backup.sh

CLOUD_REMOTE="s3-backup:nas-backup-bucket"
LOCAL_BACKUP="/backup"
LOG_FILE="/var/log/cloud-backup.log"

echo "$(date): Starting cloud backup..." >> "$LOG_FILE"

# Sync recent backups to cloud
rclone sync "$LOCAL_BACKUP" "$CLOUD_REMOTE" \
    --include "full-backup-*.tar.gz" \
    --include "database/nas-backup-*.sqlite.gz" \
    --max-age 30d \
    --progress >> "$LOG_FILE" 2>&1

if [ $? -eq 0 ]; then
    echo "$(date): ✅ Cloud backup completed" >> "$LOG_FILE"
else
    echo "$(date): ❌ Cloud backup failed" >> "$LOG_FILE"
    exit 1
fi
EOF

chmod +x /opt/backup-scripts/cloud-backup.sh

# Add to cron for weekly cloud sync
(crontab -l; echo "0 3 * * 0 /opt/backup-scripts/cloud-backup.sh") | crontab -

echo "✅ Cloud backup configured"
```

### Offsite Backup Verification

```bash
#!/bin/bash
# verify-cloud-backups.sh

CLOUD_REMOTE="s3-backup:nas-backup-bucket"
LOG_FILE="/var/log/cloud-verification.log"

echo "$(date): Verifying cloud backups..." >> "$LOG_FILE"

# List cloud backups
CLOUD_BACKUPS=$(rclone ls "$CLOUD_REMOTE" | grep -c "full-backup-.*\.tar\.gz")
LOCAL_BACKUPS=$(find /backup -name "full-backup-*.tar.gz" -mtime -30 | wc -l)

echo "$(date): Cloud backups: $CLOUD_BACKUPS, Local backups: $LOCAL_BACKUPS" >> "$LOG_FILE"

# Check for recent backups
RECENT_CLOUD=$(rclone ls "$CLOUD_REMOTE" --max-age 7d | wc -l)

if [ "$RECENT_CLOUD" -eq 0 ]; then
    echo "$(date): ❌ No recent cloud backups found" >> "$LOG_FILE"
    echo "No recent cloud backups" | mail -s "Cloud Backup Alert" admin@example.com
else
    echo "$(date): ✅ Recent cloud backups verified" >> "$LOG_FILE"
fi
```

---

*For monitoring backup systems, see [Monitoring Guide](monitoring.md). For troubleshooting backup issues, see [Troubleshooting Guide](troubleshooting.md).*