# 🚀 NAS File Manager Deployment Guide

Complete deployment guide for the NAS File Manager application across different platforms and environments.

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [Environment Setup](#environment-setup)
- [Windows Development](#windows-development)
- [Linux Production Deployment](#linux-production-deployment)
- [Docker Deployment](#docker-deployment)
- [Configuration Management](#configuration-management)
- [Monitoring & Maintenance](#monitoring--maintenance)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Windows 11 Development Setup

#### 1. Prerequisites
- Install [Node.js 20+](https://nodejs.org/)
- Install [Git](https://git-scm.com/)

#### 2. Clone and Start
```bash
git clone <your-repo-url>
cd nas-main

# Manual setup
npm install
cd backend && npm install && cd ..
cd frontend && npm install && cd ..
npm run test
```

#### 3. Access Application
- **Frontend**: http://localhost:5050
- **Backend API**: http://localhost:7777

#### 4. Test Authentication
1. Go to Account tab
2. Try both OAuth and Local login options
3. For local auth: Register with any ID/password

### Docker Quick Start

#### Development
```bash
# Start development environment
docker-compose up nas-dev
```

#### Production
```bash
# Copy and edit environment file
cp .env.example .env
# Edit .env with your settings

# Start production
docker-compose up -d nas-app
```

## Environment Setup

### Prerequisites

- **Node.js**: Version 20 or higher
- **npm**: Comes with Node.js
- **Git**: For version control
- **SQLite**: Included in dependencies

### Platform-Specific Requirements

#### Windows 11 Development
```bash
# Install Node.js from https://nodejs.org/
# Install Git from https://git-scm.com/
# Install Visual Studio Build Tools (for native modules)
npm install -g windows-build-tools
```

#### Linux Production
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install nodejs npm python3 build-essential sqlite3

# CentOS/RHEL
sudo yum install nodejs npm python3 gcc-c++ make sqlite
```

## Windows Development

### Manual Development Setup

```bash
# Install dependencies
npm install
cd backend && npm install && cd ..
cd frontend && npm install && cd ..

# Create data directories (Windows)
mkdir ..\..\nas-data
mkdir ..\..\nas-data-admin

# Start development servers
npm run test  # Starts both frontend and backend
```

### Common Commands

```bash
npm run test          # Start both frontend and backend
npm run build         # Build both frontend and backend for production
```

## Linux Production Deployment

### System Setup

```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install system dependencies
sudo apt install -y python3 build-essential sqlite3
```

### Application Setup

```bash
# Create application user
sudo useradd -r -s /bin/false -d /opt/nas nas

# Create directories
sudo mkdir -p /opt/nas
sudo mkdir -p /home/nas/nas-storage/{data,admin-data,db}

# Copy application files
sudo cp -r . /opt/nas/
sudo chown -R nas:nas /opt/nas /home/nas/nas-storage
```

### Build and Configure

```bash
cd /opt/nas

# Install dependencies and build
sudo -u nas npm install
sudo -u nas npm run build

# Setup environment
sudo -u nas cp .env.example .env
# Edit .env with production values
```

### Access Application
- **Direct**: http://your-server-ip:7777
- **With Nginx**: http://your-domain.com

## Docker Deployment

Docker is the recommended deployment method as it eliminates process management complexity and provides better isolation.

### Development with Docker

```bash
# Start development environment
docker-compose up nas-dev

# Build and start production
docker-compose up nas-app
```

### Production Docker Deployment

```bash
# Build production image
docker build -t nas-app:latest .

# Run with docker-compose
docker-compose up -d nas-app

# Or run directly
docker run -d \
  --name nas-app \
  -p 7777:7777 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/admin-data:/app/admin-data \
  -v $(pwd)/db:/app/db \
  --env-file .env \
  nas-app:latest
```

## Configuration Management

### Environment Variables

Create `.env` file from template:
```bash
cp .env.example .env
```

Key configurations:

```env
# Application
NODE_ENV=production
PORT=7777
HOST=0.0.0.0

# Authentication
AUTH_TYPE=both  # oauth, local, or both
PRIVATE_KEY=your-secure-key
ADMIN_PASSWORD=your-admin-password

# Storage (Linux Production)
NAS_DATA_DIR=/home/nas/nas-storage/data
NAS_ADMIN_DATA_DIR=/home/nas/nas-storage/admin-data
DB_PATH=/home/nas/nas-storage/db
```

### Configuration Examples

#### Development (.env with NODE_ENV=development)
```env
NODE_ENV=development
AUTH_TYPE=both
PRIVATE_KEY=dev-secret-key
ADMIN_PASSWORD=admin123
PASSWORD_MIN_LENGTH=4
```

#### Production (.env)
```env
NODE_ENV=production
AUTH_TYPE=both
PRIVATE_KEY=your-secure-secret-key-here
ADMIN_PASSWORD=your-secure-admin-password
NAS_DATA_DIR=/home/nas/nas-storage/data
```

### Platform-Specific Paths

The application automatically detects the platform and sets appropriate paths:

- **Windows Development**: Relative paths (`../../nas-data`)
- **Linux Production**: `/home/nas/nas-storage/`
- **Docker**: `/app/data`, `/app/admin-data`, `/app/db`

## Reverse Proxy Setup (Nginx)

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    client_max_body_size 50G;
    
    location / {
        proxy_pass http://localhost:7777;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Large file upload timeouts
        proxy_connect_timeout 600s;
        proxy_send_timeout 600s;
        proxy_read_timeout 600s;
    }
}
```

## Monitoring & Maintenance

### Health Checks

The application includes health check endpoint:
```bash
curl http://localhost:7777/
```

### Updates

For Docker deployments:
```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build nas-app
docker-compose up -d nas-app
```

For Linux production:
```bash
# Manual update
# Stop application (using your process manager)
# Update code
sudo -u nas npm install
sudo -u nas npm run build
# Start application
```

### Log Management

Logs are stored in:
- Docker logs: `docker-compose logs -f nas-app`
- Application logs: `./logs/` (if configured)

## Troubleshooting

### Common Issues

#### Port 7777 already in use
```bash
# Windows
netstat -ano | findstr :7777
taskkill /PID <PID> /F

# Linux
sudo netstat -tulpn | grep :7777
sudo kill -9 <PID>
```

#### Can't access application
1. Check if services are running
2. Verify firewall settings
3. Ensure correct ports are exposed

#### Database errors
1. Check data directory permissions
2. Ensure SQLite is installed
3. Verify database path in configuration

#### Permission errors (Linux)
```bash
sudo chown -R nas:nas /opt/nas
sudo chown -R nas:nas /home/nas/nas-storage
```

#### Node.js modules compilation
```bash
# Rebuild native modules
sudo -u nas npm rebuild
```

### Development vs Production Differences

| Aspect | Development (Windows) | Production (Linux) | Docker |
|--------|----------------------|-------------------|---------|
| Data Path | `../../nas-data` | `/home/nas/nas-storage/data` | `/app/data` |
| Process Manager | Direct npm scripts | Custom scripts | Docker |
| Environment | `.env` | `.env` | `.env` |
| Build | Watch mode | Built artifacts | Built artifacts |
| Database | `backend/db/` | `/home/nas/nas-storage/db/` | `/app/db` |

## Security Considerations

1. **Firewall Configuration**
   ```bash
   sudo ufw allow ssh
   sudo ufw allow 80
   sudo ufw allow 443
   sudo ufw enable
   ```

2. **SSL Certificate** (Let's Encrypt)
   ```bash
   sudo certbot --nginx -d your-domain.com
   ```

3. **File Permissions**
   - Application files: `nas:nas` with 755/644 permissions
   - Data directories: `nas:nas` with 750/640 permissions
   - Configuration files: 600 permissions

4. **Regular Updates**
   - Keep Node.js and dependencies updated
   - Monitor security advisories
   - Regular backup of data and database

## Backup Strategy

```bash
# Create backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)

# For Docker deployments
docker-compose stop nas-app
docker run --rm -v nas-app_nas-data:/data -v $(pwd)/backups:/backup alpine tar czf /backup/nas-data-$DATE.tar.gz -C /data .
docker-compose start nas-app

# For Linux deployments
sudo tar -czf /backups/nas-backup-$DATE.tar.gz \
  /home/nas/nas-storage \
  /opt/nas/.env
```

## Next Steps

1. **Security**: Change default passwords and keys
2. **SSL**: Set up HTTPS with Let's Encrypt
3. **Monitoring**: Configure log rotation and monitoring
4. **Backup**: Set up automated data backups

For Docker-specific detailed information, see [Docker Guide](docker-guide.md)