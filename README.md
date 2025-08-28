# 🗂️ NAS File Manager

A modern, production-ready Network Attached Storage (NAS) file management system with comprehensive web interface, built with Svelte 5 and Express.js.

## ✨ Key Features

### 📁 Complete File Management
- **Full CRUD Operations**: Upload, download, rename, delete, copy, move files and directories
- **Drag & Drop Interface**: Intuitive file operations with progress tracking
- **Large File Support**: Handle files up to 50GB with resumable uploads
- **Batch Operations**: ZIP/unzip, bulk file operations, folder compression
- **File Preview**: Built-in text editor with Monaco Editor integration

### 🎵 Advanced Media Support  
- **Streaming Playback**: Range-request streaming for audio and video files
- **Format Support**: MP4, MP3, WebM, and other common media formats
- **Responsive Players**: Mobile-optimized media controls
- **Thumbnail Generation**: Automatic image previews and thumbnails

### 🔐 Enterprise Authentication
- **Multi-Provider Support**: Discord OAuth, Kakao OAuth, and local ID/Password
- **Flexible Configuration**: Choose OAuth-only, local-only, or hybrid authentication
- **Permission System**: Granular intent-based access control (ADMIN, VIEW, DOWNLOAD, UPLOAD, COPY, DELETE, RENAME)
- **Session Management**: Secure JWT-based sessions with configurable expiration
- **Password Policies**: Configurable complexity requirements for production security

### 🖥️ System Management
- **User Administration**: Complete user management with permission control
- **Activity Logging**: Comprehensive audit trails for all file operations
- **System Information**: Real-time CPU, memory, disk usage monitoring
- **Health Checks**: Built-in monitoring and alerting capabilities

### 🚀 Production Deployment
- **Systemd Integration**: Native Linux service with auto-startup and monitoring
- **Docker Support**: Multi-stage containerized deployment with volume persistence
- **Reverse Proxy Ready**: Nginx/Apache configuration with SSL support
- **Auto-Scaling**: Process management with automatic restart on failure

### 📱 Cross-Platform Design
- **Responsive Interface**: Adaptive design for desktop, tablet, and mobile
- **Mobile Components**: Dedicated mobile interface components
- **Touch-Friendly**: Optimized for touch interactions and gestures
- **Progressive Web App**: Installable web app capabilities

## 🚀 Quick Start

### Development Setup
```bash
git clone <your-repository>
cd nas-main
npm run test
# Access: http://localhost:5050 (frontend) + http://localhost:7777 (backend)
```

### Production Docker Deployment
```bash
# Configure environment
cp .env.example .env
# Edit .env with your production settings

# Deploy with Docker Compose
docker-compose up -d nas-app
# Access: http://your-server:7777
```

### Linux Systemd Service
```bash
# Install as system service (auto-start on boot)
sudo cp nas-app.service /etc/systemd/system/
sudo systemctl enable nas-app.service
sudo systemctl start nas-app.service
# Access: http://your-server:7777
```

## 🏗️ Architecture

### Technology Stack
- **Frontend**: Svelte 5 + TypeScript + Vite with hot reload
- **Backend**: Express.js + TypeScript with comprehensive API
- **Database**: SQLite with Write-Ahead Logging and entity-based schema
- **Authentication**: JWT with bcrypt password hashing
- **Storage**: Configurable file system with cross-platform path resolution
- **Deployment**: Multi-stage Docker builds with Ubuntu 22.04 base

### Component Architecture
```
📦 NAS Application
├── 🎨 Frontend (Svelte 5)
│   ├── File Explorer with mobile variants
│   ├── Media players and file viewers
│   ├── User management interface
│   ├── System monitoring dashboard
│   └── Authentication components
├── ⚙️ Backend (Express.js)
│   ├── REST API with 40+ endpoints
│   ├── JWT authentication middleware
│   ├── File operation handlers
│   ├── Media streaming engine
│   └── Database integration layer
├── 🗄️ Data Layer (SQLite)
│   ├── User management with permissions
│   ├── Activity logging system
│   └── Configuration storage
└── 🚀 Deployment
    ├── Docker containerization
    ├── Systemd service integration
    └── Reverse proxy configuration
```

## 📚 Complete Documentation

Comprehensive documentation covering all aspects from development to production:

### 🛠️ Development
- **[Development Guide](Docs/development/development-guide.md)** - Complete setup, workflow, and coding standards
- **[API Reference](Docs/development/api-reference.md)** - Full REST API documentation with examples
- **[Component Guide](Docs/development/component-guide.md)** - Frontend architecture and components
- **[Testing Guide](Docs/development/testing-guide.md)** - Testing procedures and best practices

### 🚢 Deployment & Infrastructure
- **[Deployment Guide](Docs/deployment/deployment-guide.md)** - Production deployment across all platforms
- **[Docker Guide](Docs/deployment/docker-guide.md)** - Container deployment and orchestration
- **[Systemd Service](Docs/deployment/systemd-service.md)** - Linux service setup and management
- **[Production Setup](Docs/deployment/production-setup.md)** - Production environment configuration

### ⚙️ Configuration
- **[Environment Setup](Docs/configuration/environment-setup.md)** - Complete configuration reference
- **[Authentication Config](Docs/configuration/authentication.md)** - OAuth and security setup
- **[Storage Configuration](Docs/configuration/storage-config.md)** - File systems and storage backends

### 🔧 Operations & Maintenance  
- **[Troubleshooting](Docs/operations/troubleshooting.md)** - Comprehensive problem-solving guide
- **[Maintenance](Docs/operations/maintenance.md)** - System maintenance procedures
- **[Monitoring](Docs/operations/monitoring.md)** - Performance monitoring and alerting
- **[Backup & Restore](Docs/operations/backup-restore.md)** - Data protection and disaster recovery

**📖 [Complete Documentation Index](Docs/README.md)** - Full documentation navigation

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