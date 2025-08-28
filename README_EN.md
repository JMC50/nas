# 🚀 Open Source NAS System

Fully automated Docker-based NAS file management system

## ⚡ Key Features

- **🔥 One-Click Installation**: Complete setup with `docker-compose up -d`
- **🔄 Auto-Updates**: Automatic deployment on new releases
- **🐋 Lightweight Alpine**: Ultra-lightweight image under 250MB
- **🔒 Enhanced Security**: JWT authentication + non-root execution
- **📱 Responsive Web UI**: Accessible from all devices

## 🚨 Important: Fork Required!

To use this project, you **must Fork it to your own account**.

### Why is Fork necessary?

- 🔧 **Your Own Image**: Use your individual GitHub Container Registry
- 💰 **Cost Savings**: Prevents bandwidth costs on original repository
- 🎛️ **Free Customization**: Modify according to your personal requirements
- 🔄 **Independent Updates**: Manage updates on your own schedule

## 📋 Installation Guide

### Step 1: Fork Repository
```bash
# Fork this repository to your GitHub account
# https://github.com/original-author/nas → Click Fork button
```

### Step 2: Clone Forked Repository
```bash
git clone https://github.com/YOUR-USERNAME/nas.git
cd nas
```

### Step 3: Environment Setup
```bash
# Create .env file
cp .env.example .env

# Edit required settings
vim .env
```

**Required modifications:**
```bash
# Change to your GitHub repository (important!)
GITHUB_REPOSITORY=YOUR-USERNAME/nas

# Change secret key (security essential!)
JWT_SECRET=your-random-64-character-string

# Change admin password
ADMIN_PASSWORD=your-secure-password

# Data storage path
DATA_PATH=./data
```

### Step 4: One-Click Installation
```bash
# Run automated setup script
chmod +x scripts/setup.sh
./scripts/setup.sh

# Or run directly
docker-compose up -d
```

### Step 5: Access Verification
```bash
# Access web interface
http://localhost:7777

# Check status
docker-compose ps
```

## 🔄 Auto-Update System

### How It Works
1. **Push code** to your forked repository
2. **GitHub Actions** automatically builds image
3. **Your GHCR** stores the image
4. **Watchtower** checks every 5 minutes for auto-updates

### Update Flow
```bash
# Developer (you)
git add . && git commit -m "feat: add new feature"
git push origin main

# After 5 minutes automatically...
# 1. GitHub Actions starts building
# 2. New image pushed to ghcr.io/YOUR-USERNAME/nas:latest
# 3. Watchtower detects on all running servers
# 4. Zero-downtime update completes automatically ✨
```

## 🛠️ Management Commands

```bash
# View logs
docker-compose logs -f

# Restart service
docker-compose restart

# Manual update
docker-compose pull && docker-compose up -d

# Stop service
docker-compose down

# Full upgrade (using script)
./scripts/setup.sh --upgrade
```

## 📊 System Requirements

- **OS**: Linux, macOS, Windows (Docker-supported environment)
- **RAM**: Minimum 512MB, recommended 1GB
- **Storage**: Minimum 1GB (data separate)
- **Docker**: 20.10+
- **Docker Compose**: 2.0+

## 🔧 Advanced Settings

### Change Port
```bash
# In .env file
PORT=8080
```

### Change Data Path
```bash
# In .env file
DATA_PATH=/mnt/nas-storage
```

### Change Update Frequency
```bash
# In .env file (in seconds)
WATCHTOWER_POLL_INTERVAL=1800  # Every 30 minutes
```

### Disable Watchtower
```bash
# Comment out watchtower service in docker-compose.yml
# watchtower:
#   image: containrrr/watchtower:latest
#   ...
```

## 🐛 Troubleshooting

### Image Not Found
```bash
# Check GITHUB_REPOSITORY in .env
GITHUB_REPOSITORY=YOUR-USERNAME/nas  # Correct repository name

# Verify GitHub Container Registry is public
# GitHub → Your Repository → Packages → nas → Package settings → Change visibility
```

### Auto-Updates Not Working
```bash
# Check Watchtower logs
docker-compose logs watchtower

# Test manual update
docker-compose pull
```

### Port Conflicts
```bash
# Use different port in .env
PORT=8080

# Restart
docker-compose down && docker-compose up -d
```

## 🤝 Contributing Guide

1. Register issue or request feature
2. Create development branch in your Fork
3. Develop and test features
4. Create Pull Request to original repository

## 📄 License

MIT License - See [LICENSE](LICENSE) for details

## ⚠️ Important Notes

- **Security**: Always change JWT_SECRET and ADMIN_PASSWORD
- **Backup**: Perform regular data backups
- **Monitoring**: Periodically check system resource usage
- **Updates**: Data backup recommended before major updates

---

## 📚 Advanced Features & Developer Documentation

Beyond this simple installation guide, advanced features are available:

- **🔐 OAuth Authentication**: Social login integration with Discord, Kakao
- **👥 User Management**: Permission-based access control system
- **🎨 Frontend Development**: Svelte 5 + TypeScript architecture
- **🛠️ Backend API**: Complete REST API with Express.js + SQLite
- **🚀 Multiple Deployment Options**: PM2, systemd, manual installation options

For detailed information, see **[📖 Complete Documentation](Docs/README.md)**.

## 💡 Need Help?

- 📚 **Complete Documentation**: [Docs Folder](Docs/README.md) - All features and configuration guides
- 🐛 **Bug Reports**: Register [Issues](../../issues)
- 💬 **Questions**: Use [Discussions](../../discussions)
- 🇰🇷 **Korean Version**: [README.md](README.md)

**Happy NAS Life! 🎉**