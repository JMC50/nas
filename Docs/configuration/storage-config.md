# 💾 Storage Configuration

Complete guide for configuring storage, file systems, and data management in the NAS File Manager.

## 📋 Table of Contents

- [Storage Overview](#storage-overview)
- [Path Configuration](#path-configuration)
- [File System Setup](#file-system-setup)
- [Storage Backends](#storage-backends)
- [Performance Optimization](#performance-optimization)
- [Security Configuration](#security-configuration)
- [Backup Configuration](#backup-configuration)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Storage Overview

The NAS File Manager provides flexible storage configuration supporting various deployment scenarios and storage backends.

### Storage Architecture

```
NAS Application
├── User Data Storage      (DATA_PATH/data/)
├── Admin Data Storage     (DATA_PATH/admin-data/)
├── Database Storage       (DATA_PATH/database/)
├── Temporary Storage      (TEMP_DIR)
└── System Logs           (LOG_DIR - optional)
```

### Key Features

- **Unified Path System**: Single `DATA_PATH` configuration
- **Cross-Platform Support**: Windows, Linux, Docker
- **Multiple Storage Backends**: Local, NFS, RAID, Cloud mounts
- **Automatic Directory Creation**: Creates required directories on startup
- **Permission Management**: Configurable file system permissions
- **Large File Support**: Handles files up to configured size limits

## Path Configuration

### Environment Configuration

#### Basic Path Settings
```env
# Primary storage path (all data stored under this path)
DATA_PATH=/mnt/nas-storage

# Temporary files (can be separate fast storage)
TEMP_DIR=/tmp/nas

# Optional: Custom database path (overrides DATA_PATH/database)
# DB_PATH=/custom/db/location

# Optional: Custom log directory
# LOG_DIR=/var/log/nas-app
```

### Platform-Specific Configuration

#### Development (Windows)
```env
# Windows development with relative paths
DATA_PATH=../../nas-data
TEMP_DIR=C:/temp/nas

# Alternative: Absolute Windows paths
# DATA_PATH=C:/Users/Developer/nas-storage
# TEMP_DIR=C:/temp/nas
```

#### Production (Linux)
```env
# Standard Linux production paths
DATA_PATH=/mnt/nas-storage
TEMP_DIR=/tmp/nas

# Alternative: Custom mount points
# DATA_PATH=/srv/nas-storage
# DATA_PATH=/home/nas/storage
```

#### Docker Container
```env
# Docker container paths (mapped to volumes)
DATA_PATH=/app/data
TEMP_DIR=/tmp/nas

# Container volumes handle actual storage location
```

### Directory Structure

The application automatically creates this structure under `DATA_PATH`:

```
DATA_PATH/
├── data/                  # User file storage
│   ├── users/            # User-specific folders (if implemented)
│   └── shared/           # Shared file storage
├── admin-data/           # Administrative files
│   ├── system/           # System configuration backups
│   └── reports/          # Generated reports
├── database/             # SQLite database files
│   ├── nas.sqlite        # Main database
│   └── nas.sqlite-wal    # Write-ahead log (if WAL enabled)
└── temp/                 # Temporary processing files
    ├── uploads/          # Temporary upload staging
    ├── downloads/        # Temporary download preparation
    └── extracts/         # ZIP extraction workspace
```

### Path Validation

The application validates paths on startup:

```typescript
// Example path validation logic
const validateStoragePaths = () => {
  const requiredPaths = [
    { path: PATHS.dataDir, name: 'user data' },
    { path: PATHS.adminDataDir, name: 'admin data' },
    { path: PATHS.dbDir, name: 'database' },
    { path: PATHS.tempDir, name: 'temporary' }
  ];

  for (const { path, name } of requiredPaths) {
    if (!fs.existsSync(path)) {
      fs.mkdirSync(path, { recursive: true });
      console.log(`✅ Created ${name} directory: ${path}`);
    }
  }
};
```

## File System Setup

### Local File System

#### Recommended File Systems

| OS | File System | Pros | Cons |
|----|-------------|------|------|
| Linux | ext4 | Mature, reliable | No built-in compression |
| Linux | XFS | Large file performance | No built-in snapshots |
| Linux | Btrfs | Snapshots, compression | Less mature |
| Windows | NTFS | Windows native | Linux compatibility issues |
| macOS | APFS | macOS optimized | Limited cross-platform |

#### File System Preparation (Linux)

```bash
# Create partition (example: /dev/sdb1)
sudo fdisk /dev/sdb

# Format with ext4
sudo mkfs.ext4 /dev/sdb1

# Create mount point
sudo mkdir /mnt/nas-storage

# Mount filesystem
sudo mount /dev/sdb1 /mnt/nas-storage

# Add to /etc/fstab for permanent mounting
echo "/dev/sdb1 /mnt/nas-storage ext4 defaults 0 2" | sudo tee -a /etc/fstab

# Set ownership
sudo chown -R nas-user:nas-group /mnt/nas-storage
```

### Network File Systems

#### NFS Setup

**Server Configuration:**
```bash
# Install NFS server
sudo apt install nfs-kernel-server

# Configure exports
echo "/srv/nfs-storage *(rw,sync,no_subtree_check,no_root_squash)" | sudo tee -a /etc/exports

# Restart NFS service
sudo systemctl restart nfs-kernel-server
```

**Client Configuration:**
```bash
# Install NFS client
sudo apt install nfs-common

# Mount NFS share
sudo mount -t nfs server-ip:/srv/nfs-storage /mnt/nas-storage

# Add to /etc/fstab
echo "server-ip:/srv/nfs-storage /mnt/nas-storage nfs defaults 0 0" | sudo tee -a /etc/fstab
```

#### SMB/CIFS Setup

```bash
# Install CIFS utilities
sudo apt install cifs-utils

# Create credentials file
sudo bash -c 'cat > /etc/cifs-credentials << EOF
username=nas-user
password=nas-password
domain=workgroup
EOF'

# Secure credentials file
sudo chmod 600 /etc/cifs-credentials

# Mount SMB share
sudo mount -t cifs //server-ip/share /mnt/nas-storage -o credentials=/etc/cifs-credentials

# Add to /etc/fstab
echo "//server-ip/share /mnt/nas-storage cifs credentials=/etc/cifs-credentials,uid=1000,gid=1000,iocharset=utf8 0 0" | sudo tee -a /etc/fstab
```

## Storage Backends

### RAID Configuration

#### RAID Setup Script
The application includes a RAID setup script at `scripts/setup-raid.sh`:

```bash
#!/bin/bash
# scripts/setup-raid.sh - RAID configuration for NAS

RAID_LEVEL="1"           # RAID 1 (mirror)
DEVICES=("/dev/sdb" "/dev/sdc")
MOUNT_POINT="/mnt/nas-storage"
RAID_DEVICE="/dev/md0"

# Install mdadm
sudo apt update
sudo apt install -y mdadm

# Create RAID array
sudo mdadm --create $RAID_DEVICE --level=$RAID_LEVEL --raid-devices=${#DEVICES[@]} ${DEVICES[@]}

# Format with ext4
sudo mkfs.ext4 $RAID_DEVICE

# Create mount point
sudo mkdir -p $MOUNT_POINT

# Mount RAID array
sudo mount $RAID_DEVICE $MOUNT_POINT

# Add to /etc/fstab
echo "$RAID_DEVICE $MOUNT_POINT ext4 defaults 0 2" | sudo tee -a /etc/fstab

# Save RAID configuration
sudo mdadm --detail --scan | sudo tee -a /etc/mdadm/mdadm.conf

echo "✅ RAID $RAID_LEVEL array created and mounted at $MOUNT_POINT"
```

#### RAID Monitoring

```bash
# Check RAID status
sudo cat /proc/mdstat

# Detailed RAID information
sudo mdadm --detail /dev/md0

# Monitor RAID health
sudo mdadm --monitor /dev/md0 --mail admin@example.com
```

### Cloud Storage Integration

#### AWS S3 with S3FS

```bash
# Install s3fs
sudo apt install s3fs

# Configure credentials
echo "ACCESS_KEY:SECRET_KEY" | sudo tee /etc/passwd-s3fs
sudo chmod 600 /etc/passwd-s3fs

# Mount S3 bucket
s3fs bucket-name /mnt/nas-storage -o passwd_file=/etc/passwd-s3fs

# Add to /etc/fstab
echo "s3fs#bucket-name /mnt/nas-storage fuse _netdev,passwd_file=/etc/passwd-s3fs 0 0" | sudo tee -a /etc/fstab
```

#### Google Drive with rclone

```bash
# Install rclone
curl https://rclone.org/install.sh | sudo bash

# Configure Google Drive
rclone config

# Mount Google Drive
rclone mount googledrive: /mnt/nas-storage --daemon

# Systemd service for persistent mounting
sudo tee /etc/systemd/system/rclone-mount.service << EOF
[Unit]
Description=RClone Google Drive Mount
After=network.target

[Service]
Type=simple
ExecStart=/usr/bin/rclone mount googledrive: /mnt/nas-storage --allow-other --vfs-cache-mode writes
Restart=always
RestartSec=10
User=nas-user

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable rclone-mount
sudo systemctl start rclone-mount
```

## Performance Optimization

### File System Optimization

#### SSD Optimization
```bash
# Enable TRIM for SSD
sudo fstrim -v /mnt/nas-storage

# Add TRIM to cron
echo "0 2 * * 0 /usr/bin/fstrim -v /mnt/nas-storage" | sudo crontab -

# Optimize mount options for SSD
# Add 'noatime,discard' to /etc/fstab
/dev/sdb1 /mnt/nas-storage ext4 defaults,noatime,discard 0 2
```

#### HDD Optimization
```bash
# Optimize for large files
# Add 'noatime' to reduce write operations
/dev/sdb1 /mnt/nas-storage ext4 defaults,noatime 0 2
```

### Application Configuration

#### Large File Handling
```env
# Optimize for large files
MAX_FILE_SIZE=50gb
UPLOAD_TIMEOUT=600000      # 10 minutes
ENABLE_STREAMING=true      # Enable range requests

# Node.js memory optimization
NODE_OPTIONS="--max-old-space-size=4096"
```

#### Database Optimization
```env
# Enable SQLite optimizations
DB_ENABLE_WAL=true         # Write-Ahead Logging
DB_ENABLE_FOREIGN_KEYS=true
DB_TIMEOUT=30000
```

### Temporary Storage Optimization

```bash
# Use tmpfs for temporary files (RAM disk)
echo "tmpfs /tmp/nas tmpfs defaults,noatime,nosuid,nodev,noexec,mode=1777,size=2G 0 0" | sudo tee -a /etc/fstab

# Mount tmpfs
sudo mount -a

# Verify tmpfs mount
df -h /tmp/nas
```

## Security Configuration

### File System Permissions

#### Basic Permission Setup
```bash
# Create NAS user and group
sudo useradd -r -s /bin/false nas-user
sudo groupadd nas-group
sudo usermod -a -G nas-group nas-user

# Set directory ownership
sudo chown -R nas-user:nas-group /mnt/nas-storage

# Set secure permissions
sudo chmod 750 /mnt/nas-storage
sudo chmod -R 750 /mnt/nas-storage/data
sudo chmod -R 700 /mnt/nas-storage/database
```

#### Advanced Permission Setup
```bash
# Set up ACLs for fine-grained control
sudo apt install acl

# Enable ACL on filesystem
sudo tune2fs -o acl /dev/sdb1

# Set default ACLs
sudo setfacl -d -m group:nas-group:rwx /mnt/nas-storage/data
sudo setfacl -d -m other::--- /mnt/nas-storage/data
```

### Encryption

#### Full Disk Encryption (LUKS)
```bash
# Encrypt partition
sudo cryptsetup luksFormat /dev/sdb1

# Open encrypted partition
sudo cryptsetup luksOpen /dev/sdb1 nas-storage-encrypted

# Format encrypted device
sudo mkfs.ext4 /dev/mapper/nas-storage-encrypted

# Mount encrypted device
sudo mount /dev/mapper/nas-storage-encrypted /mnt/nas-storage
```

#### Directory-Level Encryption (EncFS)
```bash
# Install EncFS
sudo apt install encfs

# Create encrypted directory
encfs /mnt/raw-storage/.encrypted /mnt/nas-storage

# Add to startup script
echo "echo 'password' | encfs -S /mnt/raw-storage/.encrypted /mnt/nas-storage" >> /etc/rc.local
```

### Access Control

#### Application-Level Restrictions
```env
# Restrict file types
ALLOWED_EXTENSIONS=jpg,jpeg,png,gif,pdf,doc,docx,txt,mp3,mp4

# File size limits
MAX_FILE_SIZE=10gb

# Disable dangerous features in production
DEBUG_MODE=false
ENABLE_CORS=false  # Or restrict to specific origins
```

## Backup Configuration

### Automated Backup Setup

#### Backup Script
```bash
#!/bin/bash
# /opt/scripts/nas-backup.sh

BACKUP_DIR="/backup/nas-$(date +%Y%m%d-%H%M%S)"
SOURCE_DIR="/mnt/nas-storage"
RETENTION_DAYS=30

echo "Starting NAS backup to $BACKUP_DIR"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Stop application for consistent backup
sudo systemctl stop nas-app

# Backup data with rsync
rsync -av --progress "$SOURCE_DIR/" "$BACKUP_DIR/"

# Start application
sudo systemctl start nas-app

# Compress backup
tar -czf "$BACKUP_DIR.tar.gz" -C "$(dirname $BACKUP_DIR)" "$(basename $BACKUP_DIR)"
rm -rf "$BACKUP_DIR"

# Clean old backups
find /backup -name "nas-*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $BACKUP_DIR.tar.gz"
```

#### Scheduled Backups
```bash
# Add to crontab (daily at 2 AM)
echo "0 2 * * * /opt/scripts/nas-backup.sh >> /var/log/nas-backup.log 2>&1" | crontab -
```

### Backup Verification

#### Backup Verification Script
```bash
#!/bin/bash
# verify-backup.sh

LATEST_BACKUP=$(ls -t /backup/nas-*.tar.gz | head -1)

if [ -z "$LATEST_BACKUP" ]; then
    echo "❌ No backup found"
    exit 1
fi

# Test backup integrity
if tar -tzf "$LATEST_BACKUP" > /dev/null 2>&1; then
    echo "✅ Backup integrity verified: $LATEST_BACKUP"
else
    echo "❌ Backup corrupted: $LATEST_BACKUP"
    exit 1
fi

# Check backup size
BACKUP_SIZE=$(stat -f%z "$LATEST_BACKUP" 2>/dev/null || stat -c%s "$LATEST_BACKUP")
if [ "$BACKUP_SIZE" -gt 1000 ]; then
    echo "✅ Backup size acceptable: $(numfmt --to=iec $BACKUP_SIZE)"
else
    echo "⚠️ Backup suspiciously small: $(numfmt --to=iec $BACKUP_SIZE)"
fi
```

## Monitoring & Maintenance

### Storage Monitoring

#### Disk Usage Monitoring
```bash
#!/bin/bash
# monitor-storage.sh

THRESHOLD=90
MOUNT_POINT="/mnt/nas-storage"

USAGE=$(df "$MOUNT_POINT" | awk 'NR==2 {print $5}' | sed 's/%//')

if [ "$USAGE" -gt "$THRESHOLD" ]; then
    echo "⚠️ Storage usage at ${USAGE}% - exceeds ${THRESHOLD}% threshold"
    # Send alert (email, webhook, etc.)
    # mail -s "NAS Storage Alert" admin@example.com < /dev/null
else
    echo "✅ Storage usage at ${USAGE}% - within acceptable limits"
fi

# Log largest directories
echo "📊 Largest directories:"
du -h "$MOUNT_POINT"/* | sort -hr | head -10
```

#### RAID Health Monitoring
```bash
#!/bin/bash
# monitor-raid.sh

if [ -f /proc/mdstat ]; then
    RAID_STATUS=$(grep -o "\[U*_*\]" /proc/mdstat)
    
    if echo "$RAID_STATUS" | grep -q "_"; then
        echo "❌ RAID degraded: $RAID_STATUS"
        # Send alert
    else
        echo "✅ RAID healthy: $RAID_STATUS"
    fi
fi
```

### Performance Monitoring

#### I/O Performance Check
```bash
#!/bin/bash
# test-io-performance.sh

TEST_DIR="/mnt/nas-storage/performance-test"
mkdir -p "$TEST_DIR"

echo "🧪 Testing storage I/O performance..."

# Write test
WRITE_SPEED=$(dd if=/dev/zero of="$TEST_DIR/test-write" bs=1M count=100 2>&1 | grep -o "[0-9.]\+ MB/s")
echo "📝 Write speed: $WRITE_SPEED"

# Read test
READ_SPEED=$(dd if="$TEST_DIR/test-write" of=/dev/null bs=1M 2>&1 | grep -o "[0-9.]\+ MB/s")
echo "📖 Read speed: $READ_SPEED"

# Cleanup
rm -rf "$TEST_DIR"
```

### Maintenance Tasks

#### Storage Cleanup Script
```bash
#!/bin/bash
# cleanup-storage.sh

TEMP_DIR="/tmp/nas"
LOG_DIR="/var/log/nas-app"
RETENTION_DAYS=7

echo "🧹 Cleaning up storage..."

# Clean temporary files older than 1 day
find "$TEMP_DIR" -type f -mtime +1 -delete 2>/dev/null
echo "✅ Cleaned temporary files"

# Clean old log files
if [ -d "$LOG_DIR" ]; then
    find "$LOG_DIR" -name "*.log" -mtime +$RETENTION_DAYS -delete
    echo "✅ Cleaned old log files"
fi

# Clean empty directories
find "/mnt/nas-storage" -type d -empty -delete 2>/dev/null
echo "✅ Removed empty directories"

echo "🧹 Storage cleanup completed"
```

## Troubleshooting

### Common Storage Issues

#### Disk Space Issues

**Error: "No space left on device"**
```bash
# Check disk usage
df -h /mnt/nas-storage

# Find largest files
find /mnt/nas-storage -type f -exec ls -lh {} \; | sort -nk5 | tail -10

# Clean up temporary files
rm -rf /tmp/nas/*
```

#### Permission Issues

**Error: "Permission denied"**
```bash
# Check ownership
ls -la /mnt/nas-storage

# Fix ownership
sudo chown -R nas-user:nas-group /mnt/nas-storage

# Check file permissions
stat /mnt/nas-storage

# Fix permissions
sudo chmod 750 /mnt/nas-storage
```

#### Mount Issues

**Error: "Transport endpoint is not connected" (NFS)**
```bash
# Remount NFS share
sudo umount /mnt/nas-storage
sudo mount -a

# Check NFS server status
rpcinfo -p nfs-server-ip
```

**Error: "Device is busy"**
```bash
# Find processes using the mount
sudo lsof +D /mnt/nas-storage

# Force unmount (if safe)
sudo umount -f /mnt/nas-storage
```

### Performance Issues

#### Slow File Operations
```bash
# Check I/O wait
iostat -x 1

# Check for filesystem errors
sudo fsck /dev/sdb1

# Check RAID status (if applicable)
cat /proc/mdstat
```

#### High Memory Usage
```bash
# Check for memory leaks
ps aux | grep node | grep nas

# Check filesystem cache usage
free -h

# Clear filesystem cache (if needed)
echo 3 | sudo tee /proc/sys/vm/drop_caches
```

### Database Issues

#### Database Corruption
```bash
# Check SQLite database integrity
sqlite3 /mnt/nas-storage/database/nas.sqlite "PRAGMA integrity_check;"

# Repair database (backup first!)
cp /mnt/nas-storage/database/nas.sqlite /backup/
sqlite3 /mnt/nas-storage/database/nas.sqlite ".recover" | sqlite3 /mnt/nas-storage/database/nas-recovered.sqlite
```

#### Database Locking
```bash
# Check for database locks
lsof /mnt/nas-storage/database/nas.sqlite

# Kill blocking processes (if safe)
sudo kill -9 <PID>
```

### Configuration Validation

#### Storage Configuration Check
```bash
#!/bin/bash
# check-storage-config.sh

echo "🔍 Validating storage configuration..."

source .env

# Check DATA_PATH
if [ -z "$DATA_PATH" ]; then
    echo "❌ DATA_PATH not configured"
    exit 1
fi

if [ ! -d "$DATA_PATH" ]; then
    echo "❌ DATA_PATH directory does not exist: $DATA_PATH"
    exit 1
fi

# Check permissions
if [ ! -w "$DATA_PATH" ]; then
    echo "❌ DATA_PATH is not writable: $DATA_PATH"
    exit 1
fi

# Check disk space
USAGE=$(df "$DATA_PATH" | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$USAGE" -gt 90 ]; then
    echo "⚠️ Storage usage high: ${USAGE}%"
else
    echo "✅ Storage usage acceptable: ${USAGE}%"
fi

echo "✅ Storage configuration validation complete"
```

---

*For additional storage optimization and advanced configurations, see [Performance Guide](../operations/performance-guide.md) and [Backup & Restore Guide](../operations/backup-restore.md).*