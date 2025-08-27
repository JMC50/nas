# 🔧 Common Issues & Troubleshooting Guide

Comprehensive troubleshooting guide for the NAS File Manager application.

## Quick Diagnosis

### Application Won't Start

**Symptoms**: Server fails to start, connection refused
**Quick Check**:
```bash
# Check if ports are available
netstat -tulpn | grep :7777    # Linux
netstat -ano | findstr :7777   # Windows

# Check environment file
cat .env | grep NODE_ENV

# Check logs for errors
docker-compose logs nas-app     # Docker
npm run test                    # Development
```

### Authentication Issues

**Symptoms**: Login fails, token errors, OAuth redirect issues
**Quick Check**:
```bash
# Test auth endpoint
curl http://localhost:7777/auth/config

# Check OAuth configuration
cat .env | grep -E "(DISCORD|KAKAO)"

# Verify JWT key
cat .env | grep PRIVATE_KEY
```

### File Operations Not Working

**Symptoms**: Upload fails, download errors, permission denied
**Quick Check**:
```bash
# Check data directories
ls -la ../../nas-data           # Development
ls -la /home/nas/nas-storage    # Production
docker exec nas-app ls -la /app/data  # Docker

# Check disk space
df -h
```

## Detailed Troubleshooting

### 1. Server Startup Issues

#### Port Already in Use

**Error**: `Error: listen EADDRINUSE: address already in use :::7777`

**Solution**:
```bash
# Find process using port 7777
# Linux/macOS
sudo netstat -tulpn | grep :7777
sudo lsof -i :7777

# Windows
netstat -ano | findstr :7777

# Kill the process
# Linux/macOS
sudo kill -9 <PID>

# Windows
taskkill /PID <PID> /F

# Or change port in .env
echo "PORT=8888" >> .env
```

#### Environment Configuration Error

**Error**: `Configuration validation failed`

**Solution**:
```bash
# Check .env file exists
ls -la .env

# Verify required variables
grep -E "(NODE_ENV|PORT|PRIVATE_KEY|ADMIN_PASSWORD)" .env

# Fix missing variables
echo "PRIVATE_KEY=your-secure-key" >> .env
echo "ADMIN_PASSWORD=your-secure-password" >> .env
```

#### Database Connection Error

**Error**: `SQLITE_CANTOPEN: unable to open database file`

**Solution**:
```bash
# Check database directory exists
ls -la backend/db/

# Create directory if missing
mkdir -p backend/db/

# Check permissions (Linux/Docker)
chown -R nas:nas backend/db/
chmod 755 backend/db/

# For Docker
docker exec nas-app ls -la /app/db/
docker exec nas-app chown -R nasapp:nasapp /app/db/
```

### 2. Authentication Problems

#### OAuth Login Fails

**Symptoms**: OAuth redirect fails, "Invalid client" errors

**Discord OAuth Issues**:
```bash
# Verify Discord configuration
curl "https://discord.com/api/oauth2/applications/@me" \
  -H "Authorization: Bot YOUR_BOT_TOKEN"

# Check redirect URI matches exactly
grep DISCORD_REDIRECT_URI .env
# Should match: http://localhost:7777/login (development)
```

**Kakao OAuth Issues**:
```bash
# Verify Kakao configuration
curl -X POST "https://kauth.kakao.com/oauth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&client_id=YOUR_CLIENT_ID"

# Check redirect URI matches exactly
grep KAKAO_REDIRECT_URI .env
# Should match: http://localhost:5050/kakaoLogin (development)
```

#### Local Authentication Fails

**Error**: "Invalid credentials" or password validation errors

**Solution**:
```bash
# Check password requirements
grep -E "PASSWORD_" .env

# Reset password requirements for development
sed -i 's/PASSWORD_MIN_LENGTH=.*/PASSWORD_MIN_LENGTH=4/' .env
sed -i 's/PASSWORD_REQUIRE_UPPERCASE=.*/PASSWORD_REQUIRE_UPPERCASE=false/' .env

# Check admin password
grep ADMIN_PASSWORD .env

# Test local login
curl -X POST http://localhost:7777/auth/login \
  -H "Content-Type: application/json" \
  -d '{"id":"admin","password":"admin123"}'
```

#### JWT Token Issues

**Error**: "Token expired" or "Invalid token"

**Solution**:
```bash
# Check JWT configuration
grep -E "(PRIVATE_KEY|JWT_EXPIRY)" .env

# Verify token is being sent correctly
# In browser developer tools, check:
# - Local storage for token
# - Network requests include token parameter

# Reset JWT key (invalidates all tokens)
echo "PRIVATE_KEY=$(openssl rand -hex 32)" >> .env
```

### 3. File System Issues

#### Upload Failures

**Error**: "File upload failed" or 413 Request Entity Too Large

**Solution**:
```bash
# Check file size limits
grep MAX_FILE_SIZE .env

# Increase file size limit
sed -i 's/MAX_FILE_SIZE=.*/MAX_FILE_SIZE=50gb/' .env

# Check available disk space
df -h ../../nas-data           # Development
df -h /home/nas/nas-storage    # Production
docker exec nas-app df -h      # Docker

# Check directory permissions
ls -la ../../nas-data
chmod 755 ../../nas-data
```

#### Download/Streaming Issues

**Error**: "File not found" or streaming failures

**Solution**:
```bash
# Check file exists in storage
ls -la ../../nas-data/path/to/file

# Verify data directory configuration
grep NAS_DATA_DIR .env

# Check file permissions
ls -la ../../nas-data/
chmod -R 644 ../../nas-data/*
chmod 755 ../../nas-data/

# Test direct file access
curl "http://localhost:7777/download?token=YOUR_TOKEN&loc=&name=filename"
```

#### Permission Denied Errors

**Error**: "EACCES: permission denied"

**Solution**:
```bash
# Fix data directory permissions
# Linux production
sudo chown -R nas:nas /home/nas/nas-storage
sudo chmod -R 755 /home/nas/nas-storage

# Docker
docker exec nas-app chown -R nasapp:nasapp /app/data
docker exec nas-app chmod -R 755 /app/data

# Windows development
# Run terminal as administrator if needed
```

### 4. Frontend Issues

#### Frontend Won't Start

**Error**: Vite dev server fails to start

**Solution**:
```bash
# Check frontend dependencies
cd frontend && npm list

# Reinstall dependencies
cd frontend && rm -rf node_modules && npm install

# Check Vite configuration
cd frontend && npm run check

# Check port conflict
netstat -ano | findstr :5050  # Windows
lsof -i :5050                 # Linux/macOS
```

#### API Connection Issues

**Error**: "Failed to fetch" or CORS errors

**Solution**:
```bash
# Check API URL configuration
grep VITE_API_URL .env

# Verify backend is running
curl http://localhost:7777/

# Check CORS configuration
grep CORS_ORIGIN .env

# For development, use:
echo "CORS_ORIGIN=*" >> .env
echo "ENABLE_CORS=true" >> .env
```

#### Build Failures

**Error**: Frontend build fails with type errors

**Solution**:
```bash
# Type check
cd frontend && npm run check

# Clear build cache
cd frontend && rm -rf dist .vite

# Check TypeScript configuration
cd frontend && cat tsconfig.json

# Rebuild with verbose output
cd frontend && npm run build -- --verbose
```

### 5. Docker-Specific Issues

#### Container Won't Start

**Error**: Docker container exits immediately

**Solution**:
```bash
# Check container logs
docker-compose logs nas-app

# Inspect container
docker inspect nas-app

# Check image build
docker-compose build nas-app

# Run container interactively
docker run -it nas-app:latest /bin/sh
```

#### Volume Mount Issues

**Error**: "No such file or directory" in container

**Solution**:
```bash
# Check volume mounts
docker inspect nas-app | grep Mounts -A 20

# Verify host directories exist
ls -la ./data ./admin-data ./db

# Create missing directories
mkdir -p ./data ./admin-data ./db

# Fix permissions
sudo chown -R 1001:1001 ./data ./admin-data ./db
```

#### Health Check Failures

**Error**: Container unhealthy status

**Solution**:
```bash
# Check health status
docker inspect nas-app | grep Health -A 10

# Test health endpoint manually
docker exec nas-app wget --spider http://localhost:7777/

# Check application logs
docker exec nas-app cat /app/logs/app.log

# Restart container
docker-compose restart nas-app
```

### 6. Database Issues

#### SQLite Database Corruption

**Error**: "database disk image is malformed"

**Solution**:
```bash
# Check database integrity
sqlite3 backend/db/nas.sqlite "PRAGMA integrity_check;"

# Backup and repair database
cp backend/db/nas.sqlite backend/db/nas.sqlite.backup
sqlite3 backend/db/nas.sqlite ".dump" | sqlite3 backend/db/nas_repaired.sqlite
mv backend/db/nas_repaired.sqlite backend/db/nas.sqlite

# For Docker
docker exec nas-app sqlite3 /app/db/nas.sqlite "PRAGMA integrity_check;"
```

#### Migration Errors

**Error**: Database schema issues

**Solution**:
```bash
# Check database schema
sqlite3 backend/db/nas.sqlite ".schema"

# Reset database (WARNING: Data loss)
rm backend/db/nas.sqlite
# Restart application to recreate database

# For Docker
docker exec nas-app rm /app/db/nas.sqlite
docker-compose restart nas-app
```

### 7. Performance Issues

#### Slow File Operations

**Symptoms**: Upload/download takes too long

**Solution**:
```bash
# Check disk I/O
iostat -x 1    # Linux
wmic logicaldisk get size,freespace,caption  # Windows

# Check streaming configuration
grep ENABLE_STREAMING .env

# Monitor system resources
top    # Linux
htop   # Linux (if installed)
# Task Manager on Windows
```

#### Memory Issues

**Error**: "Out of memory" or container killed

**Solution**:
```bash
# Check memory usage
free -m                    # Linux
docker stats nas-app       # Docker

# Increase container memory limits
# In docker-compose.yml:
# deploy:
#   resources:
#     limits:
#       memory: 2G

# Check for memory leaks in logs
grep -i "memory" docker-logs
```

## Environment-Specific Troubleshooting

### Development (Windows)

```bash
# Common Windows issues
# 1. Long path names
git config --global core.longpaths true

# 2. Line ending issues
git config --global core.autocrlf true

# 3. Permission issues - run as administrator
# 4. Antivirus blocking - add exclusions
```

### Production (Linux)

```bash
# Check system services
systemctl status nas-app  # If using systemd

# Check system logs
journalctl -u nas-app -f

# Check disk space
df -h

# Check system limits
ulimit -a
```

### Docker Environment

```bash
# Check Docker daemon
systemctl status docker

# Check Docker logs
journalctl -u docker.service

# Clean up Docker resources
docker system prune -a

# Check Docker compose version
docker-compose --version
```

## Monitoring and Prevention

### Log Monitoring

```bash
# Enable comprehensive logging
echo "DEBUG_MODE=true" >> .env
echo "LOG_LEVEL=debug" >> .env
echo "ENABLE_REQUEST_LOGGING=true" >> .env

# Monitor logs in real-time
# Development
npm run test | grep -E "(ERROR|WARN)"

# Docker
docker-compose logs -f nas-app | grep -E "(ERROR|WARN)"
```

### Health Monitoring

```bash
# Create health check script
#!/bin/bash
HEALTH_URL="http://localhost:7777/"
if curl -f $HEALTH_URL > /dev/null 2>&1; then
  echo "✅ Application healthy"
else
  echo "❌ Application unhealthy"
  # Restart logic here
fi
```

### Backup Procedures

```bash
# Regular backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p backups

# Backup data
tar -czf backups/data-$DATE.tar.gz ../../nas-data

# Backup database
cp backend/db/nas.sqlite backups/database-$DATE.sqlite

# Backup configuration
cp .env backups/env-$DATE.backup
```

## Getting Help

If issues persist after trying these solutions:

1. **Check application logs** for specific error messages
2. **Verify environment configuration** matches your deployment type
3. **Test individual components** (database, auth, file system) separately
4. **Check system resources** (disk space, memory, CPU)
5. **Review recent changes** that might have caused the issue

For deployment-specific issues, see [Deployment Guide](../deployment/deployment-guide.md)
For configuration problems, see [Environment Setup](../configuration/environment-setup.md)