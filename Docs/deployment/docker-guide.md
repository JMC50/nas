# 🐳 Docker Deployment Guide

Complete guide for deploying the NAS File Manager using Docker containers.

## Overview

This application is designed to run efficiently in Docker containers with proper volume mounting and environment configuration. Docker eliminates the need for PM2 process management and provides better isolation and deployment consistency.

## Quick Start

### Development Environment

```bash
# Start development container
docker-compose up nas-dev

# Access application
# Frontend: http://localhost:5050
# Backend API: http://localhost:7777
```

### Production Environment

```bash
# Copy and configure environment
cp .env.example .env
# Edit .env with production values

# Start production container
docker-compose up -d nas-app

# Access application
# Application: http://your-server:7777
```

## Docker Configuration

### Multi-Stage Dockerfile

The application uses a multi-stage build process:

**Base Stage**: Installs Node.js 20 and system dependencies on Ubuntu 22.04
**Development Stage**: Builds application with hot reload capabilities
**Production Stage**: Optimized runtime with security user and health checks

### Key Features

- **Ubuntu 22.04**: Stable and secure base with regular updates
- **Non-root User**: Runs as `nasapp` user for security
- **Health Checks**: Built-in application health monitoring
- **Volume Mounts**: Persistent data storage
- **Same OS**: Matches production server Ubuntu for compatibility

## Environment Configuration

### Required Environment Variables

```env
# Application
NODE_ENV=production
PORT=7777
HOST=0.0.0.0

# Authentication
AUTH_TYPE=both
PRIVATE_KEY=your-secure-secret-key
ADMIN_PASSWORD=your-secure-admin-password

# Storage Paths (Docker)
NAS_DATA_DIR=/app/data
NAS_ADMIN_DATA_DIR=/app/admin-data
DB_PATH=/app/db
NAS_TEMP_DIR=/tmp/nas

# OAuth (if using)
DISCORD_CLIENT_ID=your-discord-client-id
DISCORD_CLIENT_SECRET=your-discord-client-secret
KAKAO_REST_API_KEY=your-kakao-api-key
KAKAO_CLIENT_SECRET=your-kakao-secret
```

### Volume Mapping

```yaml
volumes:
  - ./data:/app/data                    # User files
  - ./admin-data:/app/admin-data        # Admin files
  - ./db:/app/db                        # SQLite database
  - ./logs:/app/logs                    # Application logs (optional)
```

## Docker Compose Configuration

### Development Setup

```yaml
version: '3.8'
services:
  nas-dev:
    build:
      context: .
      target: development
    ports:
      - "7777:7777"
      - "5050:5050"
    volumes:
      - .:/app
      - /app/node_modules
      - nas-data:/app/data
      - nas-admin-data:/app/admin-data
    environment:
      - NODE_ENV=development
    command: npm run test
```

### Production Setup (Environment Variables)

```yaml
version: '3.8'
services:
  nas-app:
    build:
      context: .
      target: production
    ports:
      - "7777:7777"
    volumes:
      - nas-data:/app/data
      - nas-admin-data:/app/admin-data
      - nas-db:/app/db
    environment:
      - NODE_ENV=production
      - AUTH_TYPE=${AUTH_TYPE:-both}
      - PRIVATE_KEY=${PRIVATE_KEY}
      - ADMIN_PASSWORD=${ADMIN_PASSWORD}
      # ... other environment variables
    restart: unless-stopped

volumes:
  nas-data:
  nas-admin-data:
  nas-db:
```

## Deployment Methods

### Method 1: Direct Docker Run (Recommended)

```bash
# Build the image locally or pull from registry
docker build -t nas-app:latest .

# Run with environment variables
docker run -d \
  --name nas-app \
  -p 7777:7777 \
  -e NODE_ENV=production \
  -e PRIVATE_KEY="your-secure-private-key" \
  -e ADMIN_PASSWORD="your-secure-admin-password" \
  -e AUTH_TYPE=both \
  -e CORS_ORIGIN="https://your-domain.com" \
  -v nas-data:/app/data \
  -v nas-admin-data:/app/admin-data \
  -v nas-db:/app/db \
  nas-app:latest
```

### Method 2: Docker Compose (Multi-container)

```bash
# Create docker-compose.yml with your environment variables
# Then run:
docker-compose up -d nas-app
```

### Method 3: Environment File

```bash
# Create environment file
cat > production.env << EOF
PRIVATE_KEY=your-secure-private-key
ADMIN_PASSWORD=your-secure-admin-password
AUTH_TYPE=both
CORS_ORIGIN=https://your-domain.com
EOF

# Run with environment file
docker run -d --name nas-app --env-file production.env \
  -p 7777:7777 -v nas-data:/app/data nas-app:latest
```

## Production Considerations

### Security

- **Non-root User**: Container runs as `nasapp` (UID 1001)
- **Read-only Filesystem**: Application files are immutable
- **Secret Management**: Use Docker secrets for sensitive data
- **Network Isolation**: Use custom Docker networks

### Performance

- **Volume Types**: Use named volumes for better performance
- **Memory Limits**: Set appropriate memory constraints
- **CPU Limits**: Configure CPU usage limits
- **Health Checks**: Enable monitoring and automatic restarts

### Backup Strategy

```bash
# Create backup script
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)

# Stop application
docker-compose stop nas-app

# Backup volumes
docker run --rm -v nas-app_nas-data:/data -v $(pwd)/backups:/backup alpine tar czf /backup/nas-data-$DATE.tar.gz -C /data .
docker run --rm -v nas-app_nas-db:/data -v $(pwd)/backups:/backup alpine tar czf /backup/nas-db-$DATE.tar.gz -C /data .

# Restart application
docker-compose start nas-app
```

## Monitoring and Logs

### Container Logs

```bash
# View logs
docker-compose logs -f nas-app

# Filter logs
docker-compose logs nas-app | grep ERROR

# Log rotation (if needed)
docker-compose logs --since="24h" nas-app
```

### Health Monitoring

```bash
# Check health status
docker inspect --format='{{.State.Health.Status}}' nas-app

# Container stats
docker stats nas-app

# System resource usage
docker system df
```

## Nginx Reverse Proxy

### Configuration

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

### Docker Compose with Nginx

```yaml
version: '3.8'
services:
  nas-app:
    # ... nas app configuration
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf
      - ./ssl:/etc/nginx/ssl  # For HTTPS
    depends_on:
      - nas-app
```

## Troubleshooting

### Common Issues

1. **Container won't start**
   ```bash
   # Check logs
   docker-compose logs nas-app
   
   # Verify environment variables
   docker-compose config
   ```

2. **Permission errors**
   ```bash
   # Fix volume permissions
   sudo chown -R 1001:1001 ./data ./admin-data ./db
   ```

3. **Port conflicts**
   ```bash
   # Check port usage
   netstat -tulpn | grep :7777
   
   # Use different port
   docker-compose up -p 8888:7777 nas-app
   ```

4. **Database issues**
   ```bash
   # Check database volume
   docker volume inspect nas-app_nas-db
   
   # Reset database (caution: data loss)
   docker-compose down
   docker volume rm nas-app_nas-db
   docker-compose up -d nas-app
   ```

### Performance Optimization

```yaml
services:
  nas-app:
    # Resource limits
    deploy:
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
        reservations:
          memory: 512M
          cpus: '0.25'
```

## Updates and Maintenance

### Application Updates

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose down
docker-compose build nas-app
docker-compose up -d nas-app

# Cleanup old images
docker image prune
```

### Automated Updates

```bash
#!/bin/bash
# update-nas.sh

cd /path/to/nas-app

# Backup
./backup.sh

# Update
git pull origin main
docker-compose down
docker-compose build nas-app
docker-compose up -d nas-app

# Verify
sleep 30
curl -f http://localhost:7777/ || (echo "Health check failed" && exit 1)

echo "Update completed successfully"
```

## SSL/HTTPS Setup

### Let's Encrypt with Certbot

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Get certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### Docker Compose with SSL

```yaml
version: '3.8'
services:
  nas-app:
    # ... existing configuration
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx-ssl.conf:/etc/nginx/conf.d/default.conf
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - nas-app
```

This Docker-centric approach eliminates the complexity of PM2 process management while providing better isolation, scalability, and deployment consistency across different environments.