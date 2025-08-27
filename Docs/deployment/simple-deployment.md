# 🚀 Simple Docker Deployment Guide

The easiest way to deploy NAS File Manager using Docker.

## Prerequisites

- Ubuntu server (22.04+ recommended) with Docker installed
- Basic knowledge of Docker commands
- Note: The container uses Ubuntu 22.04 base, matching your server OS

## Quick Deployment

### 1. Build or Get the Image

**Option A: Build locally**
```bash
git clone <your-repo>
cd nas-main
docker build -t nas-app:latest .
```

**Option B: Pull from registry** (when available)
```bash
docker pull your-registry/nas-app:latest
```

### 2. Create Volumes

```bash
# Create persistent volumes for data
docker volume create nas-data
docker volume create nas-admin-data  
docker volume create nas-db
```

### 3. Run the Container

```bash
docker run -d \
  --name nas-app \
  --restart unless-stopped \
  -p 7777:7777 \
  -e NODE_ENV=production \
  -e PRIVATE_KEY="your-very-secure-private-key-here" \
  -e ADMIN_PASSWORD="your-secure-admin-password" \
  -e AUTH_TYPE=both \
  -e CORS_ORIGIN="*" \
  -e PASSWORD_MIN_LENGTH=8 \
  -e PASSWORD_REQUIRE_UPPERCASE=true \
  -e PASSWORD_REQUIRE_LOWERCASE=true \
  -e PASSWORD_REQUIRE_NUMBER=true \
  -e MAX_FILE_SIZE=50gb \
  -v nas-data:/app/data \
  -v nas-admin-data:/app/admin-data \
  -v nas-db:/app/db \
  nas-app:latest
```

### 4. Verify Deployment

```bash
# Check if container is running
docker ps

# Check application health
curl http://localhost:7777/

# View logs
docker logs nas-app
```

## Optional: OAuth Configuration

If you want to use Discord or Kakao login:

```bash
# Stop and remove existing container
docker stop nas-app && docker rm nas-app

# Run with OAuth configuration
docker run -d \
  --name nas-app \
  --restart unless-stopped \
  -p 7777:7777 \
  -e NODE_ENV=production \
  -e PRIVATE_KEY="your-private-key" \
  -e ADMIN_PASSWORD="your-admin-password" \
  -e AUTH_TYPE=both \
  -e DISCORD_CLIENT_ID="your-discord-client-id" \
  -e DISCORD_CLIENT_SECRET="your-discord-client-secret" \
  -e DISCORD_REDIRECT_URI="http://your-domain:7777/login" \
  -e DISCORD_LOGIN_URL="https://discord.com/oauth2/authorize?client_id=your-discord-client-id&response_type=token&redirect_uri=http://your-domain:5050/login&scope=identify" \
  -v nas-data:/app/data \
  -v nas-admin-data:/app/admin-data \
  -v nas-db:/app/db \
  nas-app:latest
```

## Management Commands

```bash
# Start/Stop
docker start nas-app
docker stop nas-app

# Restart
docker restart nas-app

# View logs
docker logs nas-app
docker logs -f nas-app  # Follow logs

# Update application
docker stop nas-app
docker rm nas-app
docker build -t nas-app:latest .  # or docker pull
# Run container again with same command

# Backup data
docker run --rm -v nas-data:/data -v $(pwd):/backup alpine tar czf /backup/nas-data-backup.tar.gz -C /data .
```

## Environment Variables

### Required
| Variable | Description | Example |
|----------|-------------|---------|
| `PRIVATE_KEY` | JWT signing key | `your-secure-32-char-key` |
| `ADMIN_PASSWORD` | Admin password | `your-secure-password` |

### Optional
| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_TYPE` | `both` | `oauth`, `local`, or `both` |
| `CORS_ORIGIN` | `*` | Allowed origins |
| `MAX_FILE_SIZE` | `50gb` | Max upload size |
| `PASSWORD_MIN_LENGTH` | `8` | Min password length |

## Nginx Reverse Proxy

If you want to use a domain name:

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    client_max_body_size 50G;
    
    location / {
        proxy_pass http://localhost:7777;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Troubleshooting

**Container won't start:**
```bash
docker logs nas-app
```

**Port already in use:**
```bash
netstat -tulpn | grep :7777
# Change port: -p 8888:7777
```

**Permission issues:**
```bash
docker exec -it nas-app ls -la /app/data
```

**Health check failed:**
```bash
docker exec -it nas-app wget -qO- http://localhost:7777/
```

That's it! Your NAS File Manager is now running in a Docker container.