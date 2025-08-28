# 🔧 Maintenance Guide

Comprehensive maintenance procedures for the NAS File Manager system.

## 📋 Table of Contents

- [Routine Maintenance](#routine-maintenance)
- [System Updates](#system-updates)
- [Database Maintenance](#database-maintenance)
- [Storage Maintenance](#storage-maintenance)
- [Security Maintenance](#security-maintenance)
- [Performance Optimization](#performance-optimization)
- [Backup Verification](#backup-verification)
- [Scheduled Maintenance](#scheduled-maintenance)

## Routine Maintenance

### Daily Tasks

#### Health Check Script
```bash
#!/bin/bash
# daily-health-check.sh

LOG_FILE="/var/log/nas-maintenance.log"
echo "$(date): Starting daily health check" >> "$LOG_FILE"

# Check service status
if ! systemctl is-active --quiet nas-app.service; then
    echo "$(date): ⚠️ Service not running - attempting restart" >> "$LOG_FILE"
    systemctl start nas-app.service
fi

# Check disk space
USAGE=$(df /mnt/nas-storage | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$USAGE" -gt 85 ]; then
    echo "$(date): ⚠️ Disk usage high: ${USAGE}%" >> "$LOG_FILE"
fi

# Check application response
if ! curl -f -s http://localhost:7777/ > /dev/null; then
    echo "$(date): ❌ Application not responding" >> "$LOG_FILE"
fi

echo "$(date): Daily health check complete" >> "$LOG_FILE"
```

### Weekly Tasks

#### System Cleanup Script
```bash
#!/bin/bash
# weekly-cleanup.sh

echo "🧹 Starting weekly system cleanup..."

# Clean temporary files
find /tmp/nas -type f -mtime +7 -delete 2>/dev/null
echo "✅ Cleaned temporary files"

# Clean old logs
find /var/log -name "*.log" -mtime +30 -delete 2>/dev/null
echo "✅ Cleaned old log files"

# Clean package cache
apt-get autoremove -y
apt-get autoclean
echo "✅ Cleaned package cache"

# Optimize database
sqlite3 /mnt/nas-storage/database/nas.sqlite "VACUUM; ANALYZE;"
echo "✅ Optimized database"

echo "🧹 Weekly cleanup complete"
```

### Monthly Tasks

#### Security Update Check
```bash
#!/bin/bash
# monthly-security-check.sh

echo "🔒 Monthly security maintenance..."

# Update package lists
apt-get update

# Check for security updates
SECURITY_UPDATES=$(apt list --upgradable 2>/dev/null | grep -i security | wc -l)
if [ "$SECURITY_UPDATES" -gt 0 ]; then
    echo "⚠️ $SECURITY_UPDATES security updates available"
    # Apply security updates
    DEBIAN_FRONTEND=noninteractive apt-get upgrade -y
fi

# Check SSL certificate expiration
if [ -f /etc/letsencrypt/live/*/cert.pem ]; then
    openssl x509 -in /etc/letsencrypt/live/*/cert.pem -checkend 2592000 >/dev/null
    if [ $? != 0 ]; then
        echo "⚠️ SSL certificate expires within 30 days"
        certbot renew --quiet
    fi
fi

echo "🔒 Security maintenance complete"
```

## System Updates

### Application Updates

#### Update Procedure
```bash
#!/bin/bash
# update-application.sh

echo "📦 Updating NAS application..."

# Create backup
BACKUP_DIR="/backup/app-backup-$(date +%Y%m%d-%H%M%S)"
cp -r /home/heesung/NAS "$BACKUP_DIR"
echo "✅ Application backed up to $BACKUP_DIR"

# Stop service
systemctl stop nas-app.service

# Update code
cd /home/heesung/NAS
git stash  # Preserve local changes
git pull origin main
git stash pop  # Restore local changes (if any)

# Update dependencies
npm install

# Rebuild application
npm run build

# Start service
systemctl start nas-app.service

# Verify update
sleep 10
if curl -f -s http://localhost:7777/ > /dev/null; then
    echo "✅ Application update successful"
else
    echo "❌ Application update failed - restoring backup"
    systemctl stop nas-app.service
    rm -rf /home/heesung/NAS
    mv "$BACKUP_DIR" /home/heesung/NAS
    systemctl start nas-app.service
fi
```

### System Updates

#### Operating System Updates
```bash
#!/bin/bash
# system-update.sh

echo "💿 Updating operating system..."

# Update package lists
apt-get update

# Upgrade packages
DEBIAN_FRONTEND=noninteractive apt-get upgrade -y

# Update kernel and critical packages
DEBIAN_FRONTEND=noninteractive apt-get dist-upgrade -y

# Clean up
apt-get autoremove -y
apt-get autoclean

# Check if reboot required
if [ -f /var/run/reboot-required ]; then
    echo "⚠️ System reboot required"
    echo "Run: sudo reboot"
fi

echo "💿 System update complete"
```

### Node.js Updates

#### Update Node.js
```bash
#!/bin/bash
# update-nodejs.sh

echo "📦 Updating Node.js..."

# Check current version
CURRENT_NODE=$(node --version)
echo "Current Node.js version: $CURRENT_NODE"

# Install latest LTS version
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
apt-get install -y nodejs

# Verify new version
NEW_NODE=$(node --version)
echo "New Node.js version: $NEW_NODE"

# Rebuild native modules
cd /home/heesung/NAS
npm rebuild

# Restart application
systemctl restart nas-app.service

echo "📦 Node.js update complete"
```

## Database Maintenance

### SQLite Optimization

#### Regular Database Maintenance
```bash
#!/bin/bash
# database-maintenance.sh

DB_PATH="/mnt/nas-storage/database/nas.sqlite"
BACKUP_PATH="/backup/db-backup-$(date +%Y%m%d-%H%M%S).sqlite"

echo "🗄️ Starting database maintenance..."

# Stop application
systemctl stop nas-app.service

# Backup database
cp "$DB_PATH" "$BACKUP_PATH"
echo "✅ Database backed up to $BACKUP_PATH"

# Check database integrity
INTEGRITY_CHECK=$(sqlite3 "$DB_PATH" "PRAGMA integrity_check;")
if [ "$INTEGRITY_CHECK" != "ok" ]; then
    echo "❌ Database integrity check failed: $INTEGRITY_CHECK"
    exit 1
fi

# Optimize database
sqlite3 "$DB_PATH" << 'EOF'
PRAGMA optimize;
VACUUM;
ANALYZE;
REINDEX;
EOF

# Update statistics
NEW_SIZE=$(du -h "$DB_PATH" | cut -f1)
OLD_SIZE=$(du -h "$BACKUP_PATH" | cut -f1)
echo "Database optimized: $OLD_SIZE -> $NEW_SIZE"

# Start application
systemctl start nas-app.service

echo "🗄️ Database maintenance complete"
```

### Database Migration

#### Migration Script Template
```bash
#!/bin/bash
# database-migration.sh

DB_PATH="/mnt/nas-storage/database/nas.sqlite"
MIGRATION_LOG="/var/log/database-migration.log"

echo "$(date): Starting database migration" >> "$MIGRATION_LOG"

# Stop application
systemctl stop nas-app.service

# Backup database
cp "$DB_PATH" "$DB_PATH.pre-migration-$(date +%Y%m%d)"

# Run migration
sqlite3 "$DB_PATH" << 'EOF'
-- Example migration
ALTER TABLE users ADD COLUMN last_activity DATETIME;
UPDATE users SET last_activity = datetime('now') WHERE last_activity IS NULL;
EOF

# Verify migration
RESULT=$(sqlite3 "$DB_PATH" "PRAGMA table_info(users);" | grep last_activity)
if [ -n "$RESULT" ]; then
    echo "$(date): Migration successful" >> "$MIGRATION_LOG"
else
    echo "$(date): Migration failed" >> "$MIGRATION_LOG"
    # Restore backup
    cp "$DB_PATH.pre-migration-$(date +%Y%m%d)" "$DB_PATH"
fi

# Start application
systemctl start nas-app.service
```

## Storage Maintenance

### Disk Space Management

#### Cleanup Script
```bash
#!/bin/bash
# storage-cleanup.sh

STORAGE_PATH="/mnt/nas-storage"
TEMP_PATH="/tmp/nas"

echo "💾 Starting storage cleanup..."

# Clean temporary files older than 7 days
find "$TEMP_PATH" -type f -mtime +7 -delete 2>/dev/null
echo "✅ Cleaned temporary files"

# Remove empty directories
find "$STORAGE_PATH/data" -type d -empty -delete 2>/dev/null
echo "✅ Removed empty directories"

# Clean old upload temporary files
find "$STORAGE_PATH/temp" -name "upload_*" -mtime +1 -delete 2>/dev/null
echo "✅ Cleaned old upload files"

# Report space saved
CURRENT_USAGE=$(df "$STORAGE_PATH" | awk 'NR==2 {print $5}')
echo "Current storage usage: $CURRENT_USAGE"

echo "💾 Storage cleanup complete"
```

### File System Check

#### RAID Maintenance (if applicable)
```bash
#!/bin/bash
# raid-maintenance.sh

if [ -f /proc/mdstat ]; then
    echo "⚙️ RAID maintenance..."
    
    # Check RAID status
    RAID_STATUS=$(cat /proc/mdstat)
    echo "RAID Status: $RAID_STATUS"
    
    # Check for degraded arrays
    if echo "$RAID_STATUS" | grep -q "_"; then
        echo "⚠️ RAID array degraded!"
        # Send alert
        mail -s "RAID Degraded" admin@example.com < /dev/null
    fi
    
    # Scrub array monthly
    DAY_OF_MONTH=$(date +%d)
    if [ "$DAY_OF_MONTH" = "01" ]; then
        echo "Starting RAID scrub..."
        echo check > /sys/block/md0/md/sync_action
    fi
fi
```

### File System Health Check
```bash
#!/bin/bash
# filesystem-check.sh

DEVICE="/dev/sdb1"
MOUNT_POINT="/mnt/nas-storage"

echo "🔍 File system health check..."

# Unmount for read-only check
if umount "$MOUNT_POINT"; then
    # Run file system check
    fsck -f -y "$DEVICE"
    FSCK_EXIT=$?
    
    # Remount
    mount "$DEVICE" "$MOUNT_POINT"
    
    if [ $FSCK_EXIT -eq 0 ]; then
        echo "✅ File system healthy"
    else
        echo "⚠️ File system issues found and corrected"
    fi
else
    echo "⚠️ Cannot unmount - running online check"
    fsck -f -n "$DEVICE"
fi
```

## Security Maintenance

### SSL Certificate Management

#### Certificate Renewal
```bash
#!/bin/bash
# ssl-maintenance.sh

echo "🔒 SSL certificate maintenance..."

# Check certificate expiration
if [ -d /etc/letsencrypt/live ]; then
    for cert_dir in /etc/letsencrypt/live/*/; do
        domain=$(basename "$cert_dir")
        if [ "$domain" != "README" ]; then
            expiry=$(openssl x509 -in "$cert_dir/cert.pem" -noout -enddate | cut -d= -f2)
            echo "Certificate for $domain expires: $expiry"
            
            # Check if expires within 30 days
            if ! openssl x509 -in "$cert_dir/cert.pem" -checkend 2592000 >/dev/null; then
                echo "⚠️ Certificate for $domain expires soon - renewing..."
                certbot renew --cert-name "$domain"
            fi
        fi
    done
fi

# Reload services after renewal
systemctl reload nginx 2>/dev/null || systemctl reload apache2 2>/dev/null
```

### Security Auditing

#### Security Audit Script
```bash
#!/bin/bash
# security-audit.sh

echo "🔒 Security audit..."

# Check file permissions
echo "Checking file permissions..."
find /mnt/nas-storage -type f -perm /o+w -exec ls -la {} \;

# Check for world-writable directories
echo "Checking directory permissions..."
find /mnt/nas-storage -type d -perm /o+w -exec ls -la {} \;

# Check for SUID/SGID files
echo "Checking for SUID/SGID files..."
find /home/heesung/NAS -type f \( -perm -4000 -o -perm -2000 \) -exec ls -la {} \;

# Check service configuration
echo "Checking service security..."
systemctl show nas-app.service | grep -E "(User|Group|PrivateNetwork|ProtectSystem)"

# Check open ports
echo "Checking open ports..."
ss -tulpn | grep :7777

echo "🔒 Security audit complete"
```

## Performance Optimization

### System Performance Tuning

#### Performance Optimization Script
```bash
#!/bin/bash
# performance-optimization.sh

echo "⚡ Performance optimization..."

# Optimize system swappiness
echo 'vm.swappiness=10' >> /etc/sysctl.conf

# Optimize file system
echo 'vm.vfs_cache_pressure=50' >> /etc/sysctl.conf

# Optimize network
echo 'net.core.rmem_max = 16777216' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 16777216' >> /etc/sysctl.conf

# Apply settings
sysctl -p

# Optimize database
sqlite3 /mnt/nas-storage/database/nas.sqlite << 'EOF'
PRAGMA cache_size = 10000;
PRAGMA temp_store = memory;
PRAGMA mmap_size = 268435456;
EOF

echo "⚡ Performance optimization complete"
```

### Application Performance

#### Node.js Optimization
```bash
#!/bin/bash
# nodejs-optimization.sh

# Set Node.js environment variables for production
cat >> /etc/systemd/system/nas-app.service.d/override.conf << 'EOF'
[Service]
Environment=NODE_OPTIONS="--max-old-space-size=2048 --optimize-for-size"
Environment=UV_THREADPOOL_SIZE=16
EOF

# Reload systemd and restart service
systemctl daemon-reload
systemctl restart nas-app.service
```

## Backup Verification

### Backup Testing

#### Backup Verification Script
```bash
#!/bin/bash
# verify-backups.sh

BACKUP_DIR="/backup"
TEST_RESTORE_DIR="/tmp/backup-test"

echo "🧪 Testing backup integrity..."

# Find latest backup
LATEST_BACKUP=$(ls -t "$BACKUP_DIR"/nas-*.tar.gz | head -1)

if [ -z "$LATEST_BACKUP" ]; then
    echo "❌ No backups found"
    exit 1
fi

# Create test directory
mkdir -p "$TEST_RESTORE_DIR"
cd "$TEST_RESTORE_DIR"

# Extract backup
if tar -xzf "$LATEST_BACKUP"; then
    echo "✅ Backup extraction successful"
    
    # Verify critical files
    if [ -f "database/nas.sqlite" ]; then
        echo "✅ Database file present"
        
        # Test database integrity
        if sqlite3 "database/nas.sqlite" "PRAGMA integrity_check;" | grep -q "ok"; then
            echo "✅ Database integrity verified"
        else
            echo "❌ Database corrupted in backup"
        fi
    else
        echo "❌ Database file missing from backup"
    fi
    
    # Check data directory
    if [ -d "data" ]; then
        echo "✅ Data directory present"
        DATA_SIZE=$(du -sh data | cut -f1)
        echo "Data size in backup: $DATA_SIZE"
    else
        echo "❌ Data directory missing from backup"
    fi
else
    echo "❌ Backup extraction failed"
fi

# Cleanup
rm -rf "$TEST_RESTORE_DIR"

echo "🧪 Backup verification complete"
```

## Scheduled Maintenance

### Cron Job Setup

#### Install Maintenance Cron Jobs
```bash
#!/bin/bash
# setup-maintenance-crons.sh

echo "⏰ Setting up maintenance cron jobs..."

# Create maintenance script directory
mkdir -p /opt/nas-maintenance
chmod +x /opt/nas-maintenance/*.sh

# Setup cron jobs
(crontab -l 2>/dev/null; cat << 'EOF'
# NAS Maintenance Jobs

# Daily health check at 6 AM
0 6 * * * /opt/nas-maintenance/daily-health-check.sh

# Weekly cleanup on Sunday at 2 AM  
0 2 * * 0 /opt/nas-maintenance/weekly-cleanup.sh

# Monthly security check on 1st at 3 AM
0 3 1 * * /opt/nas-maintenance/monthly-security-check.sh

# Database maintenance monthly on 15th at 1 AM
0 1 15 * * /opt/nas-maintenance/database-maintenance.sh

# Backup verification weekly on Saturday at 4 AM
0 4 * * 6 /opt/nas-maintenance/verify-backups.sh

# SSL certificate check monthly on 1st at 4 AM
0 4 1 * * /opt/nas-maintenance/ssl-maintenance.sh
EOF
) | crontab -

echo "⏰ Cron jobs installed successfully"
```

### Maintenance Calendar

#### Monthly Maintenance Schedule

| Week | Monday | Tuesday | Wednesday | Thursday | Friday | Saturday | Sunday |
|------|---------|---------|------------|-----------|---------|-----------|---------|
| 1st | Security Update | - | - | - | - | - | Weekly Cleanup |
| 2nd | - | - | - | - | - | Backup Verify | Weekly Cleanup |
| 3rd | DB Maintenance | - | - | - | - | - | Weekly Cleanup |
| 4th | Performance Review | - | - | - | - | Backup Verify | Weekly Cleanup |

#### Quarterly Tasks

- **Q1**: Full system backup and disaster recovery test
- **Q2**: Security audit and penetration testing  
- **Q3**: Performance benchmarking and optimization
- **Q4**: Documentation review and update

---

*For specific troubleshooting during maintenance, see [Troubleshooting Guide](troubleshooting.md). For monitoring maintenance tasks, see [Monitoring Guide](monitoring.md).*