# 📚 NAS File Manager Documentation

Welcome to the comprehensive documentation for the NAS File Manager application.

## 🚀 Quick Start

New to the project? Start here:

- **Development**: [Development Guide](development/development-guide.md)
- **Simple Deployment**: [Simple Deployment](deployment/simple-deployment.md)
- **Docker Guide**: [Docker Guide](deployment/docker-guide.md)

## 📖 Documentation Structure

### 🛠️ Development

- [Development Guide](development/development-guide.md) - Complete development environment setup and workflow
- [API Reference](development/api-reference.md) - Backend API endpoints and usage _(Coming Soon)_
- [Component Guide](development/component-guide.md) - Frontend component architecture _(Coming Soon)_

### 🚢 Deployment

- [Simple Deployment](deployment/simple-deployment.md) - Easy Docker deployment guide
- [Docker Guide](deployment/docker-guide.md) - Advanced Docker deployment options
- [Deployment Guide](deployment/deployment-guide.md) - Complete deployment across all platforms

### ⚙️ Configuration

- [Environment Setup](configuration/environment-setup.md) - Complete environment configuration guide
- [Authentication Setup](configuration/auth-setup.md) - OAuth and local authentication configuration _(Coming Soon)_
- [Storage Configuration](configuration/storage-config.md) - File system and storage settings _(Coming Soon)_

### 🔧 Troubleshooting

- [Common Issues](troubleshooting/common-issues.md) - Comprehensive troubleshooting guide
- [Performance Tuning](troubleshooting/performance.md) - Performance optimization tips _(Coming Soon)_
- [Security Guide](troubleshooting/security.md) - Security best practices _(Coming Soon)_

## 🏗️ Architecture Overview

The NAS File Manager is a full-stack web application built with:

- **Frontend**: Svelte 5 + TypeScript + Vite
- **Backend**: Express.js + TypeScript + SQLite
- **Authentication**: JWT with OAuth (Discord/Kakao) and Local ID/Password
- **Deployment**: Docker containers (Ubuntu 22.04) with volume persistence

### Key Features

- 📁 **File Management**: Upload, download, rename, delete files and directories
- 🎵 **Media Streaming**: Built-in streaming for audio and video files
- 🔐 **Flexible Authentication**: Support for OAuth providers and local accounts
- 📱 **Responsive Design**: Works on desktop and mobile devices
- 🐳 **Docker Ready**: Containerized deployment with docker-compose
- 🔧 **Centralized Configuration**: Single `.env` file for all settings

## 🎯 Use Cases

This application is perfect for:

- **Personal Cloud Storage**: Host your own file server
- **Media Center**: Stream music and videos from your collection
- **Team File Sharing**: Share files within small teams
- **Development Projects**: Store and access development assets
- **Home Network Storage**: Central storage for home network devices

## 📋 Requirements

### Development

- Node.js 20+
- npm 10+
- Git
- Modern web browser

### Production

- Linux server (Ubuntu 20.04+ recommended)
- Docker and Docker Compose
- 2GB+ RAM
- Storage space for your files

## 🚀 Quick Installation

### Development (Windows)

```bash
git clone <your-repo>
cd nas-main
npm run test
```

### Production (Docker)

```bash
docker build -t nas-app .
docker run -d --name nas-app -p 7777:7777 \
  -e PRIVATE_KEY="your-key" -e ADMIN_PASSWORD="your-password" \
  -v nas-data:/app/data nas-app:latest
```

## 🔗 Important Links

- **Application Access**: http://localhost:7777 (production) or http://localhost:5050 (development frontend)
- **API Endpoint**: http://localhost:7777 (backend API)
- **Health Check**: http://localhost:7777/ (application status)

## 📞 Getting Help

1. **Check the troubleshooting guide**: [Common Issues](troubleshooting/common-issues.md)
2. **Review logs**: Enable debug mode in your `.env` file
3. **Verify configuration**: Use the [Environment Setup](configuration/environment-setup.md) guide
4. **Test components**: Follow the [Development Guide](development/development-guide.md) testing steps

## 🔄 Migration from PM2

This application previously used PM2 for process management but has been migrated to Docker for better isolation and deployment consistency. If you're upgrading from a PM2-based deployment:

1. Stop your PM2 processes
2. Follow the [Docker Guide](deployment/docker-guide.md)
3. Migrate your data to Docker volumes
4. Update your `.env` configuration

## 🤝 Contributing

To contribute to this project:

1. Fork the repository
2. Follow the [Development Guide](development/development-guide.md) setup
3. Make your changes
4. Test thoroughly using the provided test procedures
5. Submit a pull request

## 📄 License

This project is licensed under the terms specified in the project repository.

---

_Last updated: 2025-08-27_
_Documentation version: 2.0 (Docker-centric)_
