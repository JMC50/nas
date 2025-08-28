# 🚀 Deployment Guide

Complete deployment guide for the NAS File Manager application - from simple fork-based deployment to advanced enterprise setups.

## 📋 Table of Contents

- [🎯 Quick Start (Recommended)](#quick-start-recommended)
- [System Requirements](#system-requirements)
- [⚙️ Advanced Deployment Options](#advanced-deployment-options)
- [Production Configuration](#production-configuration)
- [Reverse Proxy Setup](#reverse-proxy-setup)
- [SSL/HTTPS Setup](#sslhttps-setup)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## 🎯 Quick Start (Recommended)

### Fork-Based Auto-Updating Docker Deployment

**90% of users should use this method** - it provides automatic updates, minimal configuration, and zero maintenance.

#### 1. Fork Repository
```bash
# Fork this repository to your GitHub account
# https://github.com/original-author/nas → Click Fork button
```

#### 2. Clone Your Fork
```bash
git clone https://github.com/YOUR-USERNAME/nas.git
cd nas
```

#### 3. Configure Environment
```bash
# Create environment file
cp .env.example .env

# Edit ONLY these 3 required variables:
nano .env
```

**Required changes in .env:**
```env
# Change to your GitHub repository (important!)
GITHUB_REPOSITORY=YOUR-USERNAME/nas

# Change secret key (generate random 32+ character string)
JWT_SECRET=your-random-64-character-string

# Change admin password
ADMIN_PASSWORD=your-secure-password
```

#### 4. One-Click Deploy
```bash
# Deploy with auto-updates
docker-compose up -d

# Verify deployment
docker-compose ps
curl http://localhost:7777
```

#### 5. Access Your NAS
- **Web Interface**: http://localhost:7777
- **Login**: Use `ADMIN_PASSWORD` from your .env file
- **Auto-Updates**: Watchtower checks for updates every 5 minutes

**✅ Done!** Your NAS is running with automatic updates. When you push code changes to your fork, they'll automatically deploy to all your servers.

## ⚙️ Advanced Deployment Options

**For power users who need custom configurations**. Most users should use the [Quick Start](#quick-start-recommended) method above.

### When to Use Advanced Options

| Method | Best For | Use Case |
|--------|----------|----------|
| **Manual Linux** | Custom environments | When Docker isn't available |
| **Systemd Service** | Linux system integration | When you need OS-level service management |
| **PM2 Process Manager** | Node.js ecosystems | When you're already using PM2 |

## System Requirements

### For Fork-Based Docker Deployment (Recommended)
- **Docker**: Version 20.10+
- **Docker Compose**: Version 2.0+
- **RAM**: 512MB minimum, 2GB recommended
- **Storage**: 1GB minimum (plus space for your files)
- **OS**: Any Docker-supported system (Linux, Windows, macOS)

### For Advanced Manual Deployment
- **CPU**: 2+ cores, 2+ GHz  
- **RAM**: 4 GB available
- **Storage**: SSD, 10+ GB for application + user files
- **OS**: Ubuntu 22.04 LTS or newer
- **Node.js**: Version 20 or higher
- **Network**: Static IP, domain name for HTTPS

### Advanced Docker Configuration

**Note**: Most users should use the [Quick Start](#quick-start-recommended) method. This section is for users who need custom Docker configurations.

#### Custom Docker Compose Setup

If you need to customize volumes, networks, or other Docker settings:

```bash
# Clone your forked repository
git clone https://github.com/YOUR-USERNAME/nas.git
cd nas

# Create custom environment configuration
cp .env.example .env
```

#### 2. Configure Environment
Edit `.env` for production:
```env
# Application
NODE_ENV=production
PORT=7777
HOST=0.0.0.0

# Authentication
AUTH_TYPE=both
PRIVATE_KEY=your-secure-secret-key-here
ADMIN_PASSWORD=your-secure-admin-password
JWT_EXPIRY=24h

# Production Password Requirements  
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=false

# Storage (Docker paths)
DATA_PATH=/app/data
MAX_FILE_SIZE=50gb
ALLOWED_EXTENSIONS=*

# Security
CORS_ORIGIN=https://your-domain.com,https://www.your-domain.com
ENABLE_CORS=true

# Logging
DEBUG_MODE=false
LOG_LEVEL=info
ENABLE_REQUEST_LOGGING=true
```

#### 3. Deploy with Docker Compose
```bash
# Start production container
docker-compose up -d nas-app

# Check status
docker-compose ps
docker-compose logs -f nas-app
```

#### 4. Verify Deployment
```bash
# Health check
curl http://localhost:7777/

# Check container status
docker-compose ps nas-app

# View logs
docker-compose logs nas-app
```

### Docker Compose Configuration

**Complete docker-compose.yml:**
```yaml
version: '3.8'

services:
  nas-app:
    build:
      context: .
      dockerfile: Dockerfile
      target: production
    container_name: nas-app
    restart: unless-stopped
    ports:
      - "7777:7777"
    volumes:
      - nas-data:/app/data
      - nas-admin-data:/app/admin-data  
      - nas-db:/app/db
      - nas-temp:/tmp/nas
    env_file:
      - .env
    environment:
      - NODE_ENV=production
      - DATA_PATH=/app/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7777/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.nas.rule=Host(`your-domain.com`)"
      - "traefik.http.services.nas.loadbalancer.server.port=7777"

volumes:
  nas-data:
    driver: local
  nas-admin-data:
    driver: local
  nas-db:
    driver: local
  nas-temp:
    driver: local

networks:
  default:
    name: nas-network
```

### Docker Commands

```bash
# Build and start
docker-compose up -d nas-app

# View logs
docker-compose logs -f nas-app

# Restart application
docker-compose restart nas-app

# Stop application  
docker-compose stop nas-app

# Update and redeploy
git pull
docker-compose build nas-app
docker-compose up -d nas-app

# Backup volumes
docker run --rm -v nas-app_nas-data:/data -v $(pwd)/backups:/backup \
  alpine tar czf /backup/nas-data-$(date +%Y%m%d).tar.gz -C /data .

# Clean up
docker-compose down
docker system prune -f
```

### Manual Linux Deployment

**For advanced users only** - most users should use the [Quick Start Docker method](#quick-start-recommended).

#### 1. System Preparation

#### Install Node.js 20
```bash
# Ubuntu/Debian - Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# CentOS/RHEL/Rocky Linux
curl -fsSL https://rpm.nodesource.com/setup_20.x | sudo bash -
sudo yum install -y nodejs

# Verify installation
node --version  # Should be v20.x.x
npm --version   # Should be 10.x.x
```

#### Install System Dependencies
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y python3 build-essential sqlite3 curl git

# CentOS/RHEL
sudo yum install -y python3 gcc-c++ make sqlite curl git
```

### 2. Application Setup

#### Create Application User
```bash
# Create dedicated user for security
sudo useradd -r -s /bin/false -d /opt/nas nas

# Create directories
sudo mkdir -p /opt/nas
sudo mkdir -p /mnt/nas-storage/{data,admin-data,database,temp}

# Set permissions
sudo chown -R nas:nas /opt/nas /mnt/nas-storage
```

#### Install Application
```bash
# Clone repository
cd /opt/nas
sudo -u nas git clone <your-repository> .

# Install dependencies and build
sudo -u nas npm install
sudo -u nas npm run build

# Create environment configuration
sudo -u nas cp .env.example .env
# Edit .env with production settings (see configuration section)
```

### 3. Manual Service Management

#### Start Application
```bash
# Start as nas user
sudo -u nas NODE_ENV=production npm start

# Or start backend only
cd /opt/nas/backend
sudo -u nas node dist/index.js
```

#### Process Management with PM2
```bash
# Install PM2 globally
sudo npm install -g pm2

# Create PM2 configuration
cat > /opt/nas/ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'nas-app',
    script: './backend/dist/index.js',
    cwd: '/opt/nas',
    user: 'nas',
    env: {
      NODE_ENV: 'production',
      PORT: 7777
    },
    instances: 1,
    exec_mode: 'fork',
    watch: false,
    max_memory_restart: '1G',
    error_file: '/var/log/nas-app/error.log',
    out_file: '/var/log/nas-app/out.log',
    log_file: '/var/log/nas-app/combined.log',
    time: true
  }]
};
EOF

# Create log directory
sudo mkdir -p /var/log/nas-app
sudo chown nas:nas /var/log/nas-app

# Start with PM2
sudo -u nas pm2 start ecosystem.config.js

# Enable PM2 auto-start
sudo -u nas pm2 save
sudo -u nas pm2 startup
```

## Systemd Service Setup

The application includes a systemd service file for automatic startup and management.

### 1. Install Service
```bash
# Copy service file
sudo cp nas-app.service /etc/systemd/system/

# Reload systemd
sudo systemctl daemon-reload

# Enable auto-start
sudo systemctl enable nas-app.service
```

### 2. Service Management
```bash
# Start service
sudo systemctl start nas-app.service

# Check status
sudo systemctl status nas-app.service

# Stop service
sudo systemctl stop nas-app.service

# Restart service
sudo systemctl restart nas-app.service

# View logs
sudo journalctl -u nas-app.service -f
```

### 3. Service Configuration

**nas-app.service file:**
```ini
[Unit]
Description=NAS File Management System
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User=heesung
Environment=NODE_ENV=production
WorkingDirectory=/home/heesung/NAS
ExecStart=/usr/bin/npm start
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=nas-app

[Install]
WantedBy=multi-user.target
```

For detailed systemd setup, see [Systemd Service Guide](systemd-service.md).

## Production Configuration

### Environment Variables

**Critical Production Settings:**
```env
# Security - REQUIRED CHANGES
NODE_ENV=production
PRIVATE_KEY=your-very-secure-random-key-here
ADMIN_PASSWORD=your-secure-admin-password

# Network Security
HOST=0.0.0.0
CORS_ORIGIN=https://yourdomain.com,https://www.yourdomain.com

# Strong Password Requirements
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true  
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=true

# File Security
MAX_FILE_SIZE=50gb
ALLOWED_EXTENSIONS=*
# Or restrict: ALLOWED_EXTENSIONS=jpg,jpeg,png,pdf,doc,docx,txt,zip

# Logging
DEBUG_MODE=false
LOG_LEVEL=warn
ENABLE_REQUEST_LOGGING=true
```

### File System Security
```bash
# Set proper permissions
sudo chmod 600 .env
sudo chmod -R 750 /mnt/nas-storage
sudo chown -R nas:nas /mnt/nas-storage

# Create backup user
sudo useradd -r -s /bin/false backup
sudo usermod -a -G nas backup
```

### Firewall Configuration
```bash
# Ubuntu/Debian with UFW
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp  
sudo ufw allow 7777/tcp  # If no reverse proxy
sudo ufw enable

# CentOS/RHEL with firewalld
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --permanent --add-port=7777/tcp
sudo firewall-cmd --reload
```

## Reverse Proxy Setup

### Nginx Configuration

#### Install Nginx
```bash
# Ubuntu/Debian
sudo apt install nginx

# CentOS/RHEL
sudo yum install nginx
```

#### Configure Virtual Host
```nginx
# /etc/nginx/sites-available/nas-app
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;
    
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com www.your-domain.com;
    
    # SSL Configuration (see SSL section)
    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;
    
    # Large file upload support
    client_max_body_size 50G;
    client_body_timeout 300s;
    client_header_timeout 60s;
    
    # Proxy settings
    location / {
        proxy_pass http://127.0.0.1:7777;
        proxy_http_version 1.1;
        
        # Headers
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for large files
        proxy_connect_timeout 300s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
        
        # Buffer settings
        proxy_buffering off;
        proxy_cache off;
    }
    
    # Static asset caching
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        proxy_pass http://127.0.0.1:7777;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

#### Enable Configuration
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/nas-app /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

### Apache Configuration

#### Install Apache
```bash
# Ubuntu/Debian  
sudo apt install apache2

# CentOS/RHEL
sudo yum install httpd
```

#### Configure Virtual Host
```apache
# /etc/apache2/sites-available/nas-app.conf
<VirtualHost *:80>
    ServerName your-domain.com
    ServerAlias www.your-domain.com
    Redirect permanent / https://your-domain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName your-domain.com
    ServerAlias www.your-domain.com
    
    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/your-domain.com/cert.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/your-domain.com/privkey.pem
    SSLCertificateChainFile /etc/letsencrypt/live/your-domain.com/chain.pem
    
    # Proxy settings
    ProxyPreserveHost On
    ProxyRequests Off
    ProxyPass / http://127.0.0.1:7777/
    ProxyPassReverse / http://127.0.0.1:7777/
    
    # Large file support
    LimitRequestBody 53687091200  # 50GB
    
    # Headers
    ProxyPassReverse / http://127.0.0.1:7777/
    ProxyPassReverse / https://your-domain.com/
</VirtualHost>
```

#### Enable Modules and Site
```bash
# Enable required modules
sudo a2enmod ssl proxy proxy_http headers

# Enable site
sudo a2ensite nas-app.conf

# Restart Apache
sudo systemctl restart apache2
```

## SSL/HTTPS Setup

### Let's Encrypt with Certbot

#### Install Certbot
```bash
# Ubuntu/Debian
sudo apt install certbot python3-certbot-nginx

# CentOS/RHEL
sudo yum install certbot python3-certbot-nginx
```

#### Obtain SSL Certificate
```bash
# For Nginx
sudo certbot --nginx -d your-domain.com -d www.your-domain.com

# For Apache  
sudo certbot --apache -d your-domain.com -d www.your-domain.com

# Standalone mode (if no web server running)
sudo certbot certonly --standalone -d your-domain.com -d www.your-domain.com
```

#### Auto-Renewal Setup
```bash
# Add to crontab
echo "0 12 * * * /usr/bin/certbot renew --quiet" | sudo crontab -

# Test renewal
sudo certbot renew --dry-run
```

### Manual SSL Configuration

For custom SSL certificates, see [Environment Setup](../configuration/environment-setup.md).

## Monitoring & Maintenance

### Log Management

#### Docker Logs
```bash
# View application logs
docker-compose logs -f nas-app

# Configure log rotation in docker-compose.yml
logging:
  driver: "json-file"
  options:
    max-size: "100m"
    max-file: "5"
```

#### Systemd Logs
```bash
# View service logs
sudo journalctl -u nas-app.service -f

# Log retention configuration
sudo mkdir -p /etc/systemd/journald.conf.d
echo -e "[Journal]\nMaxRetentionSec=1month" | sudo tee /etc/systemd/journald.conf.d/retention.conf
sudo systemctl restart systemd-journald
```

### Health Monitoring

#### Create Health Check Script
```bash
#!/bin/bash
# /opt/nas/scripts/health-check.sh

HEALTH_URL="http://localhost:7777/"
TIMEOUT=10

if curl -f -s --max-time $TIMEOUT "$HEALTH_URL" > /dev/null; then
    echo "$(date): NAS service is healthy"
    exit 0
else
    echo "$(date): NAS service is unhealthy"
    # Optional: restart service
    # sudo systemctl restart nas-app.service
    exit 1
fi
```

#### Add to Cron
```bash
# Add health check every 5 minutes
echo "*/5 * * * * /opt/nas/scripts/health-check.sh >> /var/log/nas-health.log 2>&1" | crontab -
```

### Backup Strategy

#### Automated Backup Script
```bash
#!/bin/bash
# /opt/nas/scripts/backup.sh

BACKUP_DIR="/backup/nas-$(date +%Y%m%d-%H%M%S)"
DATA_DIR="/mnt/nas-storage"

mkdir -p "$BACKUP_DIR"

# Stop service for consistent backup
sudo systemctl stop nas-app.service

# Backup data and database
tar -czf "$BACKUP_DIR/nas-data.tar.gz" -C "$DATA_DIR" data admin-data database

# Backup configuration
cp /opt/nas/.env "$BACKUP_DIR/"

# Restart service
sudo systemctl start nas-app.service

# Clean old backups (keep 7 days)
find /backup -name "nas-*" -mtime +7 -exec rm -rf {} \;

echo "Backup completed: $BACKUP_DIR"
```

### Updates and Maintenance

#### Update Application
```bash
# Stop service
sudo systemctl stop nas-app.service

# Backup current version
sudo -u nas cp -r /opt/nas /opt/nas-backup-$(date +%Y%m%d)

# Update code
cd /opt/nas
sudo -u nas git pull

# Update dependencies and rebuild
sudo -u nas npm install
sudo -u nas npm run build

# Start service
sudo systemctl start nas-app.service

# Check status
sudo systemctl status nas-app.service
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check service status
sudo systemctl status nas-app.service

# Check logs
sudo journalctl -u nas-app.service -n 50

# Check configuration
cd /opt/nas
sudo -u nas node -c "require('dotenv').config(); console.log('Config loaded')"

# Check permissions
ls -la /opt/nas
ls -la /mnt/nas-storage
```

#### Port Already in Use
```bash
# Find process using port 7777
sudo lsof -i :7777
sudo netstat -tulpn | grep 7777

# Kill process
sudo kill -9 <PID>

# Or change port in .env
PORT=7778
```

#### Database Issues
```bash
# Check database file
ls -la /mnt/nas-storage/database/

# Check SQLite version
sqlite3 --version

# Test database connection
cd /opt/nas/backend
node -e "
const db = require('./dist/sqlite');
console.log('Database connected');
"
```

#### Permission Errors
```bash
# Fix file permissions
sudo chown -R nas:nas /opt/nas
sudo chown -R nas:nas /mnt/nas-storage
sudo chmod 600 /opt/nas/.env
sudo chmod -R 750 /mnt/nas-storage
```

#### High Memory Usage
```bash
# Check memory usage
free -h
ps aux | grep node

# Restart service to clear memory
sudo systemctl restart nas-app.service

# Add memory limit to service file
[Service]
MemoryLimit=2G
```

### Performance Optimization

#### Node.js Optimization
```env
# Add to .env
NODE_OPTIONS="--max-old-space-size=2048"
```

#### Database Optimization
```bash
# SQLite optimization
echo "PRAGMA optimize;" | sqlite3 /mnt/nas-storage/database/nas.sqlite
```

#### File System Optimization
```bash
# For large file operations, consider faster filesystem
# Mount SSD for database and temp files
sudo mkdir /mnt/ssd
# Add to /etc/fstab for permanent mount
```

For detailed troubleshooting, see [Operations Guide](../operations/troubleshooting.md).

---

*This deployment guide covers production deployment scenarios. For development setup, see the [Development Guide](../development/development-guide.md).*