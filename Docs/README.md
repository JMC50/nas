# 📚 NAS File Manager - Complete Documentation

Welcome to the comprehensive documentation for the NAS File Manager application - a modern, full-stack web-based Network Attached Storage system.

## 🚀 Getting Started

### 🎯 New User Path (Recommended)
**90% of users should start here** - simple fork-based installation:
- **[Main README](../README.md)** (Korean) - One-click Docker installation with auto-updates
- **[Main README](../README_EN.md)** (English) - Fork-based deployment guide  
- **[Deployment Pipeline Guide](DEPLOYMENT_PIPELINE.md)** - Technical details of the automated system

> ✅ **Quick Start**: Fork repository → Edit 3 variables → `docker-compose up -d` → Done!

### 🔧 Advanced Configuration
Once your basic system is running, explore these advanced features:
- **[Environment Setup](configuration/environment-setup.md)** - Complete environment configuration guide
- **[Authentication Setup](configuration/authentication.md)** - OAuth and local authentication configuration
- **[Storage Configuration](configuration/storage-config.md)** - File system and storage settings

## 🛠️ Development & Technical Documentation

### 🏗️ System Architecture
- **[Development Guide](development/development-guide.md)** - Complete development environment setup and workflow
- **[API Reference](development/api-reference.md)** - Backend API endpoints and usage

### 🚢 Deployment Options
- **[Deployment Guide](deployment/deployment-guide.md)** - Complete deployment across all platforms
- **[Systemd Service](deployment/systemd-service.md)** - System service setup for auto-startup

### 🔧 Operations & Maintenance
- **[Troubleshooting](operations/troubleshooting.md)** - Common issues and solutions
- **[Maintenance](operations/maintenance.md)** - System maintenance procedures
- **[Monitoring](operations/monitoring.md)** - Monitoring and logging setup
- **[Backup & Restore](operations/backup-restore.md)** - Data backup and recovery procedures

## 🏗️ Architecture Overview

The NAS File Manager is a comprehensive full-stack application featuring:

- **Frontend**: Svelte 5 + TypeScript + Vite with responsive design
- **Backend**: Express.js + TypeScript + SQLite with entity-based architecture
- **Authentication**: JWT with multi-provider support (Discord/Kakao OAuth + Local ID/Password)
- **File Management**: Complete CRUD operations with media streaming support
- **Storage**: Configurable file system paths with cross-platform support
- **Deployment**: Docker containers with systemd service integration

### Key Components

#### Backend Architecture
- **Modular Design**: Separate modules for auth, file operations, and system functions
- **Entity-Based Database**: SQLite with TypeScript entities and migrations
- **Intent System**: Granular permission control (ADMIN, VIEW, DOWNLOAD, UPLOAD, etc.)
- **Streaming Support**: Range-based streaming for media files
- **Cross-Platform Paths**: Automatic path resolution for Windows/Linux/Docker

#### Frontend Architecture
- **Component-Based**: Reusable Svelte 5 components with mobile variants
- **State Management**: Centralized stores for user data and application state
- **Responsive Design**: Desktop and mobile interfaces
- **File Operations**: Upload, download, preview, and management capabilities
- **Real-time Updates**: Live file system monitoring and updates

#### Security Features
- **JWT Authentication**: Secure token-based authentication with configurable expiry
- **Multi-Provider Auth**: Support for OAuth providers and local authentication
- **Permission System**: Role-based access control with granular intents
- **Path Sanitization**: Protection against directory traversal attacks
- **Secure File Operations**: Validated file uploads with size and type restrictions

## 🎯 Use Cases

This NAS system is designed for:

- **Personal Cloud Storage**: Private file server with web access
- **Media Center**: Streaming audio and video with built-in players
- **Team Collaboration**: File sharing with user management and permissions
- **Development Assets**: Project file storage and management
- **Home Network Storage**: Central storage hub for home networks
- **Small Business**: Departmental file sharing with access controls

## 📋 System Requirements

### Development Environment
- **Node.js**: Version 20 or higher
- **npm**: Version 10 or higher (comes with Node.js)
- **Git**: For version control and development workflow
- **Modern Browser**: Chrome, Firefox, Safari, or Edge (latest versions)
- **Operating System**: Windows 10/11, macOS, or Linux

### Production Environment
- **Server OS**: Linux (Ubuntu 20.04+ recommended)
- **Runtime**: Node.js 20+ or Docker Engine
- **Memory**: 2GB RAM minimum, 4GB recommended
- **Storage**: SSD recommended for database, adequate space for user files
- **Network**: Stable internet connection for OAuth providers (if used)

### Docker Environment
- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Host Resources**: 2GB RAM, adequate storage for volumes

## 🚦 Getting Started

### Quick Start Options

#### 1. Development Setup (Windows/Linux)
```bash
git clone <your-repository>
cd nas-main
npm run test
# Access: http://localhost:5050 (frontend) + http://localhost:7777 (backend)
```

#### 2. Docker Development
```bash
docker-compose up nas-dev
# Access: http://localhost:7777
```

#### 3. Production Docker
```bash
# Configure environment
cp .env.example .env
# Edit .env with your settings

# Deploy
docker-compose up -d nas-app
# Access: http://your-server:7777
```

#### 4. Systemd Service (Linux)
```bash
# After installation
sudo cp nas-app.service /etc/systemd/system/
sudo systemctl enable nas-app.service
sudo systemctl start nas-app.service
```

## 📊 Feature Matrix

| Feature | Status | Documentation |
|---------|--------|---------------|
| File Upload/Download | ✅ Complete | [API Reference](development/api-reference.md) |
| Media Streaming | ✅ Complete | [Component Guide](development/component-guide.md) |
| User Management | ✅ Complete | [Authentication Setup](configuration/authentication.md) |
| OAuth Integration | ✅ Complete | [Authentication Setup](configuration/authentication.md) |
| Local Authentication | ✅ Complete | [Authentication Setup](configuration/authentication.md) |
| Permission System | ✅ Complete | [API Reference](development/api-reference.md) |
| Mobile Interface | ✅ Complete | [Development Guide](development/development-guide.md) |
| Docker Support | ✅ Complete | [Deployment Guide](deployment/deployment-guide.md) |
| Systemd Service | ✅ Complete | [Systemd Service](deployment/systemd-service.md) |
| Activity Logging | ✅ Complete | [Monitoring](operations/monitoring.md) |
| System Information | ✅ Complete | [Development Guide](development/development-guide.md) |
| File Preview | ✅ Complete | [Development Guide](development/development-guide.md) |

## 🔍 Documentation Standards

This documentation follows these principles:

- **Comprehensive**: Covers all aspects from development to production
- **Practical**: Includes working examples and copy-paste commands
- **Current**: Reflects the latest codebase state and features
- **Cross-Platform**: Addresses Windows development and Linux production
- **Secure**: Emphasizes security best practices throughout

## 🤝 Contributing to Documentation

To improve this documentation:

1. **Identify Gaps**: Note missing or unclear information
2. **Follow Format**: Use the established markdown structure and style
3. **Test Examples**: Verify all code examples and commands work
4. **Update References**: Maintain cross-references between documents
5. **Version Tracking**: Update version information when making changes

## 📞 Support and Resources

- **Issues**: Check [Troubleshooting Guide](operations/troubleshooting.md) first
- **Development**: See [Development Guide](development/development-guide.md) for setup help
- **Deployment**: Reference [Deployment Guide](deployment/deployment-guide.md) for production issues
- **Configuration**: Review [Environment Setup](configuration/environment-setup.md) for config problems

## 📄 Documentation Version

- **Version**: 3.0 (Complete Rewrite)
- **Last Updated**: 2025-08-28
- **Coverage**: Full-stack application with systemd integration
- **Target Audience**: Developers, system administrators, end users

---

*This documentation is maintained alongside the codebase to ensure accuracy and completeness. For the latest updates, refer to the git repository.*