# 🔧 Troubleshooting Guide

Comprehensive troubleshooting guide for common issues with the NAS File Manager.

## 📋 Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Service Issues](#service-issues)
- [Authentication Problems](#authentication-problems)
- [File Operation Issues](#file-operation-issues)
- [Database Problems](#database-problems)
- [Network & Connectivity](#network--connectivity)
- [Performance Issues](#performance-issues)
- [Storage Problems](#storage-problems)
- [Configuration Issues](#configuration-issues)
- [Log Analysis](#log-analysis)

## Quick Diagnostics

### System Health Check Script

```bash
#!/bin/bash
# health-check.sh - Quick system diagnosis

echo "🩺 NAS System Health Check"
echo "========================="

# Check service status
echo "📊 Service Status:"
sudo systemctl is-active nas-app.service
sudo systemctl is-enabled nas-app.service

# Check application response
echo "🌐 Application Response:"
if curl -f -s http://localhost:7777/ > /dev/null; then
    echo "✅ Application responding"
else
    echo "❌ Application not responding"
fi

# Check disk space
echo "💾 Storage Status:"
df -h /mnt/nas-storage 2>/dev/null || df -h ../../nas-data 2>/dev/null || echo "⚠️ Storage path not found"

# Check memory usage
echo "🧠 Memory Usage:"
free -h

# Check recent errors
echo "🚨 Recent Errors:"
sudo journalctl -u nas-app.service --since "1 hour ago" | grep -i error | tail -5

echo "========================="
echo "Health check complete"
```

### First Steps Checklist

When encountering issues, check these items first:

1. **Service Running**: `sudo systemctl status nas-app.service`
2. **Application Responding**: `curl http://localhost:7777/`
3. **Port Available**: `sudo lsof -i :7777`
4. **Storage Accessible**: `ls -la /mnt/nas-storage/` or `ls -la ../../nas-data/`
5. **Environment File**: `ls -la .env`
6. **Recent Logs**: `sudo journalctl -u nas-app.service -f`

## Service Issues

### Service Won't Start

#### Check Service Status
```bash
# Detailed service status
sudo systemctl status nas-app.service -l

# Check service file
sudo systemctl cat nas-app.service
```

#### Common Causes

**1. User doesn't exist:**
```bash
# Error: "User 'heesung' could not be found"
# Solution: Create user or update service file
id heesung || sudo useradd -r -s /bin/bash heesung
```

**2. Working directory doesn't exist:**
```bash
# Error: "WorkingDirectory '/home/heesung/NAS' is not a directory"
# Solution: Create directory or fix path
ls -la /home/heesung/NAS
# If doesn't exist, either create it or edit service file
```

**3. npm command not found:**
```bash
# Error: "nas-app.service: Failed to execute command: No such file or directory"
# Solution: Fix npm path in service file
which npm
sudo systemctl edit nas-app.service
# Update ExecStart path
```

**4. Permission denied:**
```bash
# Error: "Permission denied" 
# Solution: Fix file permissions
sudo chown -R heesung:heesung /home/heesung/NAS
sudo chmod +x /home/heesung/NAS/package.json
```

### Service Starts But Fails Immediately

#### Check Application Logs
```bash
# Recent startup logs
sudo journalctl -u nas-app.service -n 50

# Follow logs in real-time
sudo journalctl -u nas-app.service -f
```

#### Common Issues

**1. Environment file missing:**
```bash
# Error: "Cannot find module" or environment variable errors
ls -la /home/heesung/NAS/.env
# If missing, copy from template
cp .env.example .env
```

**2. Dependencies not installed:**
```bash
# Error: "Cannot find module 'express'"
cd /home/heesung/NAS
sudo -u heesung npm install
```

**3. Database permission issues:**
```bash
# Error: "SQLITE_CANTOPEN: unable to open database file"
ls -la /mnt/nas-storage/database/
sudo mkdir -p /mnt/nas-storage/database
sudo chown -R heesung:heesung /mnt/nas-storage
```

### Service Stops Unexpectedly

#### Check for Crashes
```bash
# Look for crash signals
sudo journalctl -u nas-app.service | grep -i "killed\|terminated\|segfault"

# Check system messages
sudo dmesg | grep -i "killed\|oom"
```

#### Memory Issues
```bash
# Check memory usage
ps aux | grep node | grep nas

# Check for out-of-memory kills
grep -i "killed process" /var/log/syslog

# Solution: Add memory limit to service
sudo systemctl edit nas-app.service
# Add: MemoryLimit=2G
```

### Port Conflicts

```bash
# Error: "EADDRINUSE: address already in use :::7777"
# Find process using port 7777
sudo lsof -i :7777
sudo netstat -tulpn | grep 7777

# Kill conflicting process
sudo kill -9 <PID>

# Or change port in .env
echo "PORT=7778" >> .env
```

## Authentication Problems

### OAuth Authentication Issues

#### Discord OAuth Problems

**1. Invalid client credentials:**
```bash
# Check Discord configuration
grep DISCORD .env

# Test Discord API connectivity
curl -X POST "https://discord.com/api/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$DISCORD_CLIENT_ID&client_secret=$DISCORD_CLIENT_SECRET&grant_type=client_credentials"
```

**2. Redirect URI mismatch:**
```bash
# Check redirect URI in .env matches Discord app settings
grep DISCORD_REDIRECT_URI .env
# Must exactly match Discord application configuration
```

#### Kakao OAuth Problems

**1. API key issues:**
```bash
# Check Kakao configuration
grep KAKAO .env

# Verify API key is valid
curl "https://kapi.kakao.com/v1/user/access_token_info" \
  -H "Authorization: KakaoAK $KAKAO_REST_API_KEY"
```

### Local Authentication Issues

#### Password Requirements

```bash
# Error: "Password does not meet requirements"
# Check password policy in .env
grep PASSWORD_ .env

# Test password validation
node -e "
const password = 'TestPass123!';
console.log('Length:', password.length >= parseInt(process.env.PASSWORD_MIN_LENGTH || 8));
console.log('Uppercase:', /[A-Z]/.test(password) || !process.env.PASSWORD_REQUIRE_UPPERCASE);
console.log('Lowercase:', /[a-z]/.test(password) || !process.env.PASSWORD_REQUIRE_LOWERCASE);
console.log('Number:', /[0-9]/.test(password) || !process.env.PASSWORD_REQUIRE_NUMBER);
console.log('Special:', /[^A-Za-z0-9]/.test(password) || !process.env.PASSWORD_REQUIRE_SPECIAL);
"
```

### JWT Token Issues

#### Invalid Token Errors

```bash
# Error: "Invalid token"
# Check JWT configuration
grep -E "(PRIVATE_KEY|JWT_EXPIRY)" .env

# Verify token format (should be three base64 segments separated by dots)
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." | grep -E '^[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+$'
```

#### Token Expiration

```bash
# Check token expiration
node -e "
const jwt = require('jsonwebtoken');
const token = 'your-jwt-token-here';
try {
  const decoded = jwt.decode(token);
  console.log('Expires:', new Date(decoded.exp * 1000));
  console.log('Current:', new Date());
  console.log('Expired:', decoded.exp * 1000 < Date.now());
} catch (e) {
  console.log('Invalid token:', e.message);
}
"
```

## File Operation Issues

### Upload Problems

#### File Size Limits

```bash
# Error: "File too large"
# Check file size limit
grep MAX_FILE_SIZE .env

# Check actual file size
ls -lh /path/to/file

# Convert size limit to bytes for comparison
node -e "console.log('50gb =', 50 * 1024 * 1024 * 1024, 'bytes')"
```

#### Upload Timeout

```bash
# Error: "Upload timeout"
# Check timeout settings
grep UPLOAD_TIMEOUT .env

# Increase timeout for large files
echo "UPLOAD_TIMEOUT=600000" >> .env  # 10 minutes
```

#### Permission Issues

```bash
# Error: "Permission denied" during upload
# Check storage directory permissions
ls -la /mnt/nas-storage/

# Fix permissions
sudo chown -R nas-user:nas-group /mnt/nas-storage/
sudo chmod -R 755 /mnt/nas-storage/
```

### Download Problems

#### File Not Found

```bash
# Error: "File not found"
# Check file exists
ls -la "/mnt/nas-storage/data/path/to/file"

# Check file permissions
stat "/mnt/nas-storage/data/path/to/file"
```

#### Streaming Issues

```bash
# Error: Media streaming not working
# Check streaming configuration
grep ENABLE_STREAMING .env

# Test range request support
curl -r 0-1000 "http://localhost:7777/getVideoData?token=jwt&loc=/path&name=video.mp4"
```

### File Management Issues

#### Directory Creation Fails

```bash
# Error: "Cannot create directory"
# Check parent directory exists and is writable
ls -la /mnt/nas-storage/data/

# Check disk space
df -h /mnt/nas-storage/
```

#### File Deletion Issues

```bash
# Error: "Cannot delete file"
# Check file permissions
ls -la /path/to/file

# Check if file is in use
sudo lsof /path/to/file

# Force deletion (careful!)
sudo rm -f /path/to/file
```

## Database Problems

### Database Connection Issues

#### Database File Access

```bash
# Error: "SQLITE_CANTOPEN: unable to open database file"
# Check database file and directory
ls -la /mnt/nas-storage/database/
ls -la /mnt/nas-storage/database/nas.sqlite

# Create database directory if missing
sudo mkdir -p /mnt/nas-storage/database
sudo chown -R nas-user:nas-group /mnt/nas-storage/database
```

#### Database Corruption

```bash
# Check database integrity
sqlite3 /mnt/nas-storage/database/nas.sqlite "PRAGMA integrity_check;"

# If corrupted, restore from backup
cp /backup/nas.sqlite /mnt/nas-storage/database/nas.sqlite.backup
sqlite3 /mnt/nas-storage/database/nas.sqlite.backup ".recover" | sqlite3 /mnt/nas-storage/database/nas.sqlite
```

### Database Lock Issues

```bash
# Error: "SQLITE_BUSY: database is locked"
# Check for processes holding database locks
sudo lsof /mnt/nas-storage/database/nas.sqlite

# Kill blocking processes (if safe)
sudo kill -9 <PID>

# Remove lock files (if application is stopped)
rm -f /mnt/nas-storage/database/nas.sqlite-shm
rm -f /mnt/nas-storage/database/nas.sqlite-wal
```

## Network & Connectivity

### CORS Issues

#### Cross-Origin Requests Blocked

```bash
# Error: "CORS policy: No 'Access-Control-Allow-Origin' header"
# Check CORS configuration
grep CORS .env

# Allow specific origins
echo "CORS_ORIGIN=https://yourdomain.com,https://www.yourdomain.com" >> .env

# Or allow all origins (development only)
echo "CORS_ORIGIN=*" >> .env
```

### Reverse Proxy Issues

#### Nginx Configuration Problems

```bash
# Test Nginx configuration
sudo nginx -t

# Check Nginx error logs
sudo tail -f /var/log/nginx/error.log

# Common fix: Increase buffer sizes
# Add to nginx config:
# proxy_buffers 16 32k;
# proxy_buffer_size 32k;
```

#### SSL Certificate Issues

```bash
# Check certificate validity
openssl x509 -in /etc/letsencrypt/live/domain.com/cert.pem -text -noout

# Renew Let's Encrypt certificate
sudo certbot renew

# Test SSL
curl -I https://yourdomain.com
```

## Performance Issues

### Slow Application Response

#### High CPU Usage

```bash
# Check CPU usage
htop
ps aux | sort -nrk 3,3 | head -10

# Check Node.js process
ps aux | grep node | grep nas
```

#### High Memory Usage

```bash
# Check memory usage
free -h
ps aux | sort -nrk 4,4 | head -10

# Add memory limit to service
sudo systemctl edit nas-app.service
# Add: MemoryLimit=2G
```

#### Slow Database Queries

```bash
# Enable SQLite query logging
echo ".timer on" | sqlite3 /mnt/nas-storage/database/nas.sqlite "SELECT COUNT(*) FROM users;"

# Optimize database
sqlite3 /mnt/nas-storage/database/nas.sqlite "VACUUM; ANALYZE;"
```

### Storage Performance Issues

#### Slow File Operations

```bash
# Test disk I/O performance
sudo hdparm -tT /dev/sdb

# Test directory I/O
time ls -la /mnt/nas-storage/data/

# Check for filesystem errors
sudo fsck /dev/sdb1
```

## Storage Problems

### Disk Space Issues

#### Out of Space

```bash
# Check disk usage
df -h /mnt/nas-storage

# Find largest files
find /mnt/nas-storage -type f -exec ls -lh {} \; | sort -nk5 | tail -20

# Clean temporary files
rm -rf /tmp/nas/*
```

#### Permission Issues

```bash
# Fix ownership
sudo chown -R nas-user:nas-group /mnt/nas-storage

# Fix permissions
find /mnt/nas-storage -type d -exec chmod 755 {} \;
find /mnt/nas-storage -type f -exec chmod 644 {} \;
```

### Network Storage Issues

#### NFS Mount Problems

```bash
# Error: "Transport endpoint is not connected"
# Remount NFS
sudo umount /mnt/nas-storage
sudo mount -a

# Check NFS server
rpcinfo -p nfs-server-ip

# Test NFS connectivity
showmount -e nfs-server-ip
```

## Configuration Issues

### Environment Variable Problems

#### Missing Configuration

```bash
# Check for missing required variables
node -e "
const required = ['NODE_ENV', 'PORT', 'PRIVATE_KEY', 'ADMIN_PASSWORD'];
required.forEach(key => {
  if (!process.env[key]) {
    console.log('Missing:', key);
  }
});
"
```

#### Invalid Configuration Values

```bash
# Validate configuration
node -e "
const config = {
  NODE_ENV: process.env.NODE_ENV,
  PORT: parseInt(process.env.PORT),
  AUTH_TYPE: process.env.AUTH_TYPE
};

console.log('NODE_ENV valid:', ['development', 'production'].includes(config.NODE_ENV));
console.log('PORT valid:', config.PORT > 0 && config.PORT < 65536);
console.log('AUTH_TYPE valid:', ['local', 'oauth', 'both'].includes(config.AUTH_TYPE));
"
```

## Log Analysis

### Reading Service Logs

#### Systemd Logs

```bash
# Recent logs
sudo journalctl -u nas-app.service -n 100

# Logs from specific time
sudo journalctl -u nas-app.service --since "2025-08-28 10:00:00"

# Follow logs in real-time
sudo journalctl -u nas-app.service -f

# Filter by priority
sudo journalctl -u nas-app.service -p err
```

#### Application Logs

```bash
# If application logs to files
tail -f /var/log/nas-app/app.log

# Search for specific errors
grep -i "error\|exception\|fail" /var/log/nas-app/app.log
```

### Common Log Patterns

#### Authentication Errors
```bash
# Look for auth failures
sudo journalctl -u nas-app.service | grep -i "auth\|login\|token"
```

#### Database Errors
```bash
# Look for database issues
sudo journalctl -u nas-app.service | grep -i "sqlite\|database\|db"
```

#### File Operation Errors
```bash
# Look for file operation failures
sudo journalctl -u nas-app.service | grep -i "file\|upload\|download\|permission"
```

## Emergency Procedures

### Service Recovery

```bash
#!/bin/bash
# emergency-recovery.sh

echo "🚨 Emergency NAS Service Recovery"

# Stop service
sudo systemctl stop nas-app.service

# Kill any remaining processes
pkill -f "nas"

# Clear temporary files
rm -rf /tmp/nas/*

# Check and fix permissions
sudo chown -R nas-user:nas-group /mnt/nas-storage
sudo chmod -R 755 /mnt/nas-storage

# Start service
sudo systemctl start nas-app.service

# Check status
sleep 5
sudo systemctl status nas-app.service
```

### Data Recovery

```bash
#!/bin/bash
# data-recovery.sh

echo "💾 Data Recovery Procedure"

# Stop application
sudo systemctl stop nas-app.service

# Check filesystem
sudo fsck /dev/sdb1

# Mount in read-only mode for safety
sudo mount -o ro /dev/sdb1 /mnt/recovery

# Copy critical data
cp -r /mnt/recovery/database /backup/emergency-db-$(date +%Y%m%d)

echo "✅ Critical data backed up to /backup/"
```

---

*For additional support, see [Monitoring Guide](monitoring.md) and [Maintenance Guide](maintenance.md). For configuration-related issues, see [Configuration Guides](../configuration/).*