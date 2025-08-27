# 🗂️ NAS File Manager

A modern, web-based Network Attached Storage (NAS) file management system built with Svelte 5 and Express.js.

## ✨ Features

- 📁 **Complete File Management**: Upload, download, rename, delete files and directories
- 🎵 **Media Streaming**: Built-in streaming support for audio and video files
- 🔐 **Flexible Authentication**: OAuth (Discord/Kakao) and local ID/Password support
- 📱 **Responsive Design**: Works seamlessly on desktop and mobile
- 🐳 **Docker Ready**: Containerized deployment with docker-compose
- ⚙️ **Centralized Configuration**: Single `.env` file for all settings

## 🚀 Quick Start

### Development (Windows)
```bash
git clone <your-repo>
cd nas-main
npm run test
```

### Production (Docker)
```bash
# Build image
docker build -t nas-app:latest .

# Run with environment variables
docker run -d \
  --name nas-app \
  -p 7777:7777 \
  -e PRIVATE_KEY="your-secure-key" \
  -e ADMIN_PASSWORD="your-secure-password" \
  -v nas-data:/app/data \
  -v nas-admin-data:/app/admin-data \
  -v nas-db:/app/db \
  nas-app:latest
```

**Access**: http://localhost:7777

## 🏗️ Architecture

- **Frontend**: Svelte 5 + TypeScript + Vite
- **Backend**: Express.js + TypeScript + SQLite
- **Authentication**: JWT with multiple provider support
- **Storage**: File system with configurable paths
- **Deployment**: Docker containers (Ubuntu 22.04 base) with volume persistence

## 📚 Documentation

Complete documentation is available in the `Docs/` directory:

- **[📖 Full Documentation](Docs/README.md)** - Complete documentation index
- **[🛠️ Development Guide](Docs/development/development-guide.md)** - Development setup and workflow
- **[🚢 Deployment Guide](Docs/deployment/deployment-guide.md)** - Production deployment
- **[🐳 Docker Guide](Docs/deployment/docker-guide.md)** - Docker-specific deployment
- **[⚙️ Environment Setup](Docs/configuration/environment-setup.md)** - Configuration guide
- **[🔧 Troubleshooting](Docs/troubleshooting/common-issues.md)** - Common issues and solutions

## 🎯 Use Cases

- **Personal Cloud Storage**: Host your own private file server
- **Media Center**: Stream your music and video collection
- **Team Collaboration**: Share files within small teams
- **Development Assets**: Store and manage project resources
- **Home Network Storage**: Central file hub for your home network

## 📋 Requirements

### Development
- Node.js 20+
- npm 10+
- Git
- Modern web browser

### Production
- Docker and Docker Compose
- 2GB+ RAM recommended
- Adequate storage for your files

## 🔧 Configuration

The application uses a centralized `.env` file for all configuration:

```env
# Basic setup
NODE_ENV=production
PORT=7777
AUTH_TYPE=both

# Security
PRIVATE_KEY=your-secure-secret-key
ADMIN_PASSWORD=your-secure-admin-password

# Storage paths (auto-detected for platform)
NAS_DATA_DIR=../../nas-data         # Development
# NAS_DATA_DIR=/app/data             # Docker
# NAS_DATA_DIR=/home/nas/storage     # Linux production
```

See [Environment Setup](Docs/configuration/environment-setup.md) for complete configuration options.

## 🚀 Deployment Options

### Docker (Recommended)
```bash
docker run -d --name nas-app -p 7777:7777 \
  -e PRIVATE_KEY="your-key" -e ADMIN_PASSWORD="your-password" \
  -v nas-data:/app/data nas-app:latest
```

### Manual Linux Deployment
```bash
# Install dependencies
sudo apt install nodejs npm python3 build-essential sqlite3

# Setup application
npm install && npm run build

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start application
node backend/dist/index.js
```

## 🔐 Authentication

Support for multiple authentication methods:

- **OAuth Providers**: Discord, Kakao
- **Local Authentication**: ID/Password with configurable complexity
- **Flexible Configuration**: Use OAuth only, local only, or both

## 📱 API Access

The application provides a complete REST API:

- **Base URL**: `http://localhost:7777`
- **Authentication**: JWT tokens via query parameter or header
- **Health Check**: `GET /` (returns application status)

See [API Reference](Docs/development/api-reference.md) for detailed endpoint documentation.

## 🛠️ Development

```bash
# Install dependencies
npm install
cd backend && npm install && cd ..
cd frontend && npm install && cd ..

# Start development servers
npm run test  # Starts both frontend (port 5050) and backend (port 7777)

# Build for production
npm run build
```

## 🔧 Troubleshooting

Common issues and solutions:

- **Port conflicts**: Change `PORT` in `.env`
- **Permission errors**: Check file system permissions
- **OAuth issues**: Verify OAuth provider configuration
- **Database problems**: Check SQLite path and permissions

See [Common Issues](Docs/troubleshooting/common-issues.md) for comprehensive troubleshooting.

## 📊 Project Status

- ✅ **Docker-centric deployment** (migrated from PM2)
- ✅ **Centralized configuration** (single `.env` file)
- ✅ **TypeScript support** throughout
- ✅ **Comprehensive documentation**
- ✅ **Cross-platform support** (Windows dev, Linux prod, Docker)

## 🤝 Contributing

1. Fork the repository
2. Follow the [Development Guide](Docs/development/development-guide.md)
3. Make your changes with proper testing
4. Submit a pull request

## 📄 License

This project is licensed under the terms specified in the repository.

---

*For detailed information, see the complete [Documentation](Docs/README.md)*