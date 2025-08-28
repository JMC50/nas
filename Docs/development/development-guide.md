# 🛠️ Development Guide

Complete guide for setting up and working with the NAS File Manager development environment.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Environment Setup](#environment-setup)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Development Commands](#development-commands)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Debugging](#debugging)
- [Common Issues](#common-issues)

## Prerequisites

### Required Software

**Node.js 20+**
```bash
# Verify installation
node --version  # Should be v20.x.x or higher
npm --version   # Should be 10.x.x or higher
```

**Git**
```bash
# Verify installation
git --version
```

**Code Editor**
- **Recommended**: Visual Studio Code with extensions:
  - Svelte for VS Code
  - TypeScript and JavaScript Language Server
  - Prettier - Code formatter
  - ESLint

### Platform-Specific Setup

#### Windows 11 Development
```bash
# Install Node.js from https://nodejs.org/
# Install Git from https://git-scm.com/
# Install Visual Studio Build Tools for native modules
npm install -g windows-build-tools
```

#### Linux Development
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install nodejs npm python3 build-essential sqlite3

# Verify Node.js version (should be 20+)
node --version
# If version is older, install Node.js 20:
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### macOS Development
```bash
# Using Homebrew
brew install node npm sqlite3
# Verify version
node --version
```

## Environment Setup

### 1. Clone Repository
```bash
git clone <your-repository-url>
cd nas-main
```

### 2. Install Dependencies
```bash
# Install root dependencies
npm install

# Install backend dependencies
cd backend
npm install
cd ..

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### 3. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Edit .env with development settings
```

#### Development .env Configuration
```env
# Environment
NODE_ENV=development
PORT=7777
HOST=localhost
FRONTEND_PORT=5050

# URLs
SERVER_URL=http://localhost:7777
FRONTEND_URL=http://localhost:5050
API_BASE_URL=http://localhost:7777

# Authentication (Development defaults)
AUTH_TYPE=both
PRIVATE_KEY=development-secret-key
ADMIN_PASSWORD=admin123
JWT_EXPIRY=24h

# Password Requirements (Relaxed for development)
PASSWORD_MIN_LENGTH=4
PASSWORD_REQUIRE_UPPERCASE=false
PASSWORD_REQUIRE_LOWERCASE=false
PASSWORD_REQUIRE_NUMBER=false
PASSWORD_REQUIRE_SPECIAL=false

# Storage Paths (Windows Development)
NAS_DATA_DIR=../../nas-data
NAS_ADMIN_DATA_DIR=../../nas-data-admin
DB_PATH=./backend/db/nas.sqlite
NAS_TEMP_DIR=/tmp/nas

# Storage Configuration
MAX_FILE_SIZE=10gb
ALLOWED_EXTENSIONS=*
ENABLE_STREAMING=true

# Development Features
DEBUG_MODE=true
LOG_LEVEL=debug
ENABLE_CORS=true
ENABLE_REQUEST_LOGGING=true

# CORS (Development - allow all)
CORS_ORIGIN=*

# OAuth (Optional - leave empty if not testing OAuth)
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
KAKAO_REST_API_KEY=
KAKAO_CLIENT_SECRET=
```

### 4. Create Data Directories
```bash
# Windows
mkdir ..\..\nas-data
mkdir ..\..\nas-data-admin

# Linux/macOS
mkdir -p ../../nas-data
mkdir -p ../../nas-data-admin
```

### 5. Start Development Environment
```bash
# Start both frontend and backend servers
npm run test

# This starts:
# - Frontend dev server at http://localhost:5050
# - Backend server at http://localhost:7777
# - TypeScript compiler in watch mode
# - Hot reload for both frontend and backend
```

## Development Workflow

### Daily Development Process

1. **Start Development Servers**
```bash
npm run test
```

2. **Access Application**
- **Frontend**: http://localhost:5050
- **Backend API**: http://localhost:7777
- **Health Check**: http://localhost:7777/

3. **Test Authentication**
- Navigate to Account tab
- Try local registration/login
- Test OAuth if configured

4. **Development Features**
- Hot reload for both frontend and backend changes
- TypeScript compilation errors shown in terminal
- Automatic backend restart on file changes

### Git Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and test
npm run test

# Stage and commit changes
git add .
git commit -m "[feat] your feature description"

# Push feature branch
git push origin feature/your-feature-name
```

## Project Structure

```
nas-main/
├── README.md                 # Project overview
├── package.json              # Root package.json for monorepo
├── .env                      # Environment configuration
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Docker orchestration
├── nas-app.service          # Systemd service definition
├── scripts/                 # Deployment and utility scripts
│   ├── docker-build.sh      # Docker build automation
│   └── setup-raid.sh        # RAID configuration
├── backend/                 # Backend application
│   ├── package.json         # Backend dependencies
│   ├── tsconfig.json        # TypeScript configuration
│   ├── src/                 # Source code
│   │   ├── index.ts         # Main server entry point
│   │   ├── sqlite.ts        # Database connection
│   │   ├── config/          # Configuration modules
│   │   │   ├── config.ts    # Main configuration
│   │   │   └── paths.ts     # Path resolution
│   │   ├── entity/          # Database entities
│   │   │   ├── user.entity.ts
│   │   │   ├── intents.entity.ts
│   │   │   └── log.entity.ts
│   │   ├── functions/       # Business logic
│   │   │   ├── auth.ts      # Authentication functions
│   │   │   └── general.ts   # General utilities
│   │   ├── modules/         # Feature modules
│   │   │   └── authModule.ts # Authentication module
│   │   ├── migrations/      # Database migrations
│   │   │   └── addLocalAuth.ts
│   │   └── db/             # Database utilities
│   │       ├── init.ts     # Database initialization
│   │       └── metadata.ts # Entity metadata
│   └── db/                 # SQLite database location
├── frontend/               # Frontend application
│   ├── package.json        # Frontend dependencies
│   ├── vite.config.ts      # Vite configuration
│   ├── tsconfig.json       # TypeScript configuration
│   ├── src/                # Source code
│   │   ├── main.ts         # Application entry point
│   │   ├── App.svelte      # Root component
│   │   ├── vite-env.d.ts   # Vite type definitions
│   │   ├── store/          # State management
│   │   │   └── store.ts    # Svelte stores
│   │   └── lib/            # Svelte components
│   │       ├── Explorer.svelte          # File browser
│   │       ├── Explorer_mobile.svelte   # Mobile file browser
│   │       ├── FileViewer.svelte        # File preview/editor
│   │       ├── FileViewer_mobile.svelte # Mobile file viewer
│   │       ├── FileManager.svelte       # File operations
│   │       ├── UserManager.svelte       # User management
│   │       ├── Account.svelte           # Account management
│   │       ├── LocalLogin.svelte        # Local authentication
│   │       ├── ActivityLog.svelte       # System activity
│   │       ├── SystemInfo.svelte        # System information
│   │       ├── SystemInfo_mobile.svelte # Mobile system info
│   │       └── Setting.svelte           # Application settings
└── Docs/                   # Documentation
    ├── README.md           # Documentation index
    ├── development/        # Development guides
    ├── deployment/         # Deployment guides
    ├── configuration/      # Configuration guides
    └── operations/         # Operations guides
```

### Key Architecture Components

#### Backend Architecture

**Main Server (`backend/src/index.ts`)**
- Express.js server with TypeScript
- All API endpoints in single file for simplicity
- JWT authentication middleware
- File upload/download handling
- Media streaming support

**Database Layer**
- SQLite with better-sqlite3
- Entity-based schema definition
- Migration system for schema updates
- Intent-based permission system

**Authentication System**
- Multi-provider support (OAuth + Local)
- JWT token generation and validation
- User registration and management
- Permission-based access control

#### Frontend Architecture

**Svelte 5 Components**
- Reactive component system
- TypeScript support throughout
- Mobile-responsive design
- State management with stores

**Component Categories**
- **File Management**: Explorer, FileViewer, FileManager
- **Authentication**: Account, LocalLogin, OAuth redirects
- **Administration**: UserManager, ActivityLog, SystemInfo
- **Navigation**: SideMenu, BottomMenu
- **Settings**: Setting component for configuration

## Development Commands

### Root Level Commands
```bash
# Start development environment (both frontend and backend)
npm run test

# Build both frontend and backend for production
npm run build

# Install dependencies for all packages
npm install
```

### Backend Commands
```bash
cd backend

# Compile TypeScript and watch for changes
tsc -w

# Run compiled backend with auto-restart
nodemon dist/index.js

# Development mode (compile + run)
npm start

# Build for production
npm run build

# Type check without building
tsc --noEmit
```

### Frontend Commands
```bash
cd frontend

# Start development server (port 5050)
npm run dev

# Build for production
npm run build

# Type check Svelte components
npm run check

# Preview production build
npm run preview
```

## Code Standards

### TypeScript Configuration

**Backend tsconfig.json**
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

**Frontend tsconfig.json**
```json
{
  "extends": "@tsconfig/svelte/tsconfig.json",
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "strict": true
  },
  "include": ["src/**/*.d.ts", "src/**/*.ts", "src/**/*.svelte"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

### Coding Conventions

**File Naming**
- **TypeScript**: camelCase.ts (e.g., `authModule.ts`)
- **Svelte Components**: PascalCase.svelte (e.g., `FileViewer.svelte`)
- **Configuration**: kebab-case (e.g., `tsconfig.json`)

**Code Style**
- **Indentation**: 2 spaces
- **Quotes**: Single quotes for strings
- **Semicolons**: Required
- **Trailing Commas**: Required in multiline structures

**Component Structure (Svelte)**
```svelte
<script lang="ts">
  // Imports
  import { onMount } from 'svelte';
  
  // Props
  export let prop1: string;
  export let prop2: number = 0;
  
  // Variables
  let localVariable = '';
  
  // Functions
  function handleAction() {
    // Implementation
  }
  
  // Lifecycle
  onMount(() => {
    // Initialization
  });
</script>

<!-- HTML Template -->
<div class="component-root">
  <!-- Content -->
</div>

<!-- Styles -->
<style>
  .component-root {
    /* Styles */
  }
</style>
```

## Testing

### Manual Testing Checklist

**Authentication Testing**
```bash
# 1. Start development environment
npm run test

# 2. Test local authentication
# - Go to Account tab
# - Register new user with ID/password
# - Logout and login again
# - Verify JWT token in browser storage

# 3. Test file operations (requires login)
# - Upload files
# - Browse directories
# - Download files
# - Test media streaming (audio/video files)

# 4. Test user management (admin user)
# - Login with admin credentials
# - Access User Manager
# - View user list and permissions
# - Test activity logging
```

**Cross-Browser Testing**
- Chrome (latest)
- Firefox (latest)
- Safari (if on macOS)
- Edge (latest)

**Responsive Testing**
- Desktop: 1920x1080, 1366x768
- Tablet: 768x1024
- Mobile: 375x667, 414x896

### API Testing

**Health Check**
```bash
curl http://localhost:7777/
# Should return HTML page
```

**Authentication API**
```bash
# Get auth config
curl http://localhost:7777/auth/config

# Register user (local auth)
curl -X POST http://localhost:7777/auth/register \
  -H "Content-Type: application/json" \
  -d '{"userId":"test","password":"test123"}'

# Login
curl -X POST http://localhost:7777/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userId":"test","password":"test123"}'
```

## Debugging

### Backend Debugging

**Enable Debug Mode**
```env
# In .env
DEBUG_MODE=true
LOG_LEVEL=debug
ENABLE_REQUEST_LOGGING=true
```

**Common Debug Scenarios**
```bash
# Check server logs
# Logs appear in terminal where you ran 'npm run test'

# Check database
# Database file: backend/db/nas.sqlite
# Use SQLite browser or CLI to inspect

# Check file paths
# Verify data directories exist: ../../nas-data, ../../nas-data-admin
```

**VS Code Backend Debugging**
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Backend",
      "type": "node",
      "request": "launch",
      "program": "${workspaceFolder}/backend/dist/index.js",
      "preLaunchTask": "tsc:build",
      "outFiles": ["${workspaceFolder}/backend/dist/**/*.js"],
      "envFile": "${workspaceFolder}/.env"
    }
  ]
}
```

### Frontend Debugging

**Browser Developer Tools**
- **Console**: Check for JavaScript errors
- **Network**: Monitor API calls and responses
- **Application**: Inspect localStorage for JWT tokens
- **Sources**: Debug Svelte component code

**Svelte DevTools**
Install browser extension for enhanced Svelte debugging

## Common Issues

### Port Conflicts
```bash
# Error: Port 7777 already in use
# Solution: Kill existing process
# Windows:
netstat -ano | findstr :7777
taskkill /PID <PID> /F

# Linux/macOS:
lsof -ti:7777
kill -9 <PID>

# Or change port in .env:
PORT=7778
```

### Node Modules Issues
```bash
# Error: Module resolution problems
# Solution: Clean install
rm -rf node_modules package-lock.json
rm -rf backend/node_modules backend/package-lock.json
rm -rf frontend/node_modules frontend/package-lock.json

npm install
cd backend && npm install && cd ..
cd frontend && npm install && cd ..
```

### Database Issues
```bash
# Error: Database locked or permission denied
# Solution: Check database file permissions
ls -la backend/db/

# If directory doesn't exist:
mkdir -p backend/db

# Check SQLite installation
sqlite3 --version
```

### TypeScript Compilation Issues
```bash
# Error: TypeScript compilation errors
# Solution: Check tsconfig.json settings and fix type errors

# Backend compilation
cd backend
npx tsc --noEmit

# Frontend type checking
cd frontend
npm run check
```

### OAuth Configuration
```bash
# Error: OAuth authentication failing
# Check .env configuration:
# - DISCORD_CLIENT_ID and DISCORD_CLIENT_SECRET
# - KAKAO_REST_API_KEY and KAKAO_CLIENT_SECRET
# - Redirect URIs match OAuth provider settings
```

## Next Steps

After completing development setup:

1. **Read API Reference**: [API Reference](api-reference.md)
2. **Understand Components**: [Component Guide](component-guide.md)
3. **Learn Testing**: [Testing Guide](testing-guide.md)
4. **Prepare Deployment**: [Deployment Guide](../deployment/deployment-guide.md)

## Development Resources

- **Svelte 5 Documentation**: https://svelte.dev/docs
- **TypeScript Handbook**: https://www.typescriptlang.org/docs/
- **Express.js Guide**: https://expressjs.com/en/guide/
- **Vite Documentation**: https://vitejs.dev/guide/
- **SQLite Documentation**: https://www.sqlite.org/docs.html

---

*This guide covers the complete development workflow. For production deployment, see the [Deployment Guide](../deployment/deployment-guide.md).*