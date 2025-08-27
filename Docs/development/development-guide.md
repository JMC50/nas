# 💻 Development Guide

Comprehensive guide for developing the NAS File Manager application.

## Development Environment Setup

### Prerequisites

- **Node.js**: Version 18 or higher
- **npm**: Comes with Node.js
- **Git**: For version control
- **Visual Studio Code**: Recommended IDE

### Quick Setup

```bash
# Clone repository
git clone <your-repo-url>
cd nas-main

# Install dependencies
npm install
cd backend && npm install && cd ..
cd frontend && npm install && cd ..

# Start development servers
npm run test  # Starts both frontend and backend
```

## Architecture Overview

### Technology Stack

- **Frontend**: Svelte 5 + TypeScript + Vite
- **Backend**: Express + TypeScript + SQLite
- **Authentication**: JWT with OAuth (Discord/Kakao) and Local ID/Password
- **Build System**: TypeScript compiler + Vite

### Project Structure

```
nas-main/
├── backend/                 # Express.js backend
│   ├── src/
│   │   ├── index.ts        # Main server file
│   │   ├── config/         # Configuration management
│   │   ├── functions/      # Business logic
│   │   ├── modules/        # Authentication modules
│   │   └── db/             # Database entities
│   └── dist/               # Compiled JavaScript
├── frontend/               # Svelte frontend
│   ├── src/
│   │   ├── App.svelte      # Main application
│   │   ├── lib/            # Components
│   │   └── store/          # State management
│   └── dist/               # Built frontend
├── Docs/                   # Documentation
└── .env                    # Environment configuration
```

## Development Workflow

### Starting Development

```bash
# Terminal 1: Backend with auto-restart
cd backend
npm start  # Compiles TypeScript and starts server

# Terminal 2: Frontend dev server
cd frontend
npm run dev  # Starts Vite dev server with HMR

# Or start both simultaneously
npm run test  # Starts both backend and frontend
```

### Available Commands

#### Root Level
```bash
npm run test              # Start both frontend and backend
npm run build             # Build both for production
npm install               # Install all dependencies
```

#### Backend
```bash
cd backend
npm start                 # Development mode (build + watch)
npm run build             # Compile TypeScript
npm run dev               # Start with nodemon
tsc -w                    # Watch TypeScript compilation
```

#### Frontend
```bash
cd frontend
npm run dev               # Development server with HMR
npm run build             # Build for production
npm run preview           # Preview production build
npm run check             # Type checking
```

## Configuration System

### Environment Configuration

The application uses a centralized `.env` file in the project root:

```env
# Development Configuration
NODE_ENV=development
PORT=7777
HOST=localhost
FRONTEND_PORT=5050

# Authentication
AUTH_TYPE=both
PRIVATE_KEY=development-secret-key
ADMIN_PASSWORD=admin123

# Storage Paths
NAS_DATA_DIR=../../nas-data
NAS_ADMIN_DATA_DIR=../../nas-data-admin
DB_PATH=./db/nas.sqlite
```

### Backend Configuration

Configuration is loaded from `backend/src/config/environment.ts`:

```typescript
import { config } from "dotenv";

// Load from root .env file
const rootEnvPath = join(__dirname, '../../../.env');
config({ path: rootEnvPath });

export class Environment {
  static readonly NODE_ENV = process.env.NODE_ENV || 'development';
  static readonly PORT = parseInt(process.env.PORT || '7777');
  // ... other configurations
}
```

### Frontend Configuration

Frontend uses Vite's `loadEnv` to read environment variables:

```typescript
// vite.config.ts
import { loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  
  return {
    define: {
      __SERVER_URL__: JSON.stringify(env.SERVER_URL),
      __KAKAO_LOGIN_URL__: JSON.stringify(env.KAKAO_LOGIN_URL),
      // ... other environment variables
    }
  };
});
```

## Backend Development

### Core Components

#### Main Server (`backend/src/index.ts`)
- Express.js server setup
- All API endpoints
- JWT authentication middleware
- File handling operations

#### Database (`backend/src/sqlite.ts`)
- SQLite connection using better-sqlite3
- Database initialization
- Entity-based table creation

#### Authentication (`backend/src/modules/authModule.ts`)
- Centralized authentication logic
- OAuth providers (Discord, Kakao)
- Local ID/Password authentication
- JWT token management

### API Endpoints

#### Authentication
```typescript
GET  /auth/config          # Get auth configuration
POST /auth/register        # Register new user
POST /auth/login           # Local login
POST /auth/change-password # Change password
GET  /login                # Discord OAuth callback
GET  /kakaoLogin           # Kakao OAuth callback
```

#### File Operations
```typescript
GET    /files              # List files
POST   /upload             # Upload files
GET    /download           # Download files
POST   /delete             # Delete files
POST   /rename             # Rename files
GET    /stream             # Stream media files
```

### Database Schema

```typescript
// User entity
interface User {
  id: string;
  discord_id?: string;
  kakao_id?: string;
  local_id?: string;
  password?: string;  // bcrypt hashed
  auth_type: 'oauth' | 'local';
  created_at: string;
}

// Permission system
interface UserIntent {
  user_id: string;
  intent: 'ADMIN' | 'VIEW' | 'DOWNLOAD' | 'UPLOAD' | 'DELETE';
}
```

## Frontend Development

### Component Architecture

#### Main Application (`App.svelte`)
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import Explorer from './lib/Explorer.svelte';
  import Account from './lib/Account.svelte';
  
  let currentTab = 'explorer';
  let userToken = '';
</script>

{#if currentTab === 'explorer'}
  <Explorer bind:token={userToken} />
{:else if currentTab === 'account'}
  <Account bind:token={userToken} />
{/if}
```

#### File Explorer (`lib/Explorer.svelte`)
- File browser interface
- Upload/download functionality
- File preview capabilities
- Drag-and-drop support

#### Authentication (`lib/Account.svelte`)
- Dynamic authentication UI
- OAuth and local login options
- User registration forms

### State Management

Using Svelte stores for global state:

```typescript
// store/store.ts
import { writable } from 'svelte/store';

export const userToken = writable<string>('');
export const currentPath = writable<string>('');
export const fileList = writable<FileItem[]>([]);
```

### API Integration

```typescript
// API helper functions
async function apiCall(endpoint: string, options = {}) {
  const response = await fetch(`${SERVER_URL}${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options
  });
  
  return response.json();
}
```

## Testing

### Backend Testing

```bash
# Test API endpoints
curl http://localhost:7777/auth/config

# Test authentication
curl -X POST http://localhost:7777/auth/login \
  -H "Content-Type: application/json" \
  -d '{"id":"test","password":"test123"}'
```

### Frontend Testing

```bash
# Type checking
cd frontend && npm run check

# Manual testing
# 1. Open http://localhost:5050
# 2. Test authentication flows
# 3. Test file operations
```

## Debugging

### Backend Debugging

```bash
# Enable debug mode in .env
DEBUG_MODE=true
LOG_LEVEL=debug

# View logs in terminal
cd backend && npm start
```

### Frontend Debugging

```bash
# Vite dev server provides:
# - Hot Module Replacement (HMR)
# - Source maps
# - Browser dev tools integration

cd frontend && npm run dev
```

### Common Issues

1. **Port conflicts**: Change PORT in .env
2. **Permission errors**: Check file system permissions
3. **Build errors**: Verify dependencies are installed
4. **CORS issues**: Check CORS_ORIGIN in .env

## Code Style

### TypeScript Configuration

```json
// tsconfig.json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "commonjs",
    "strict": true,
    "esModuleInterop": true
  }
}
```

### Formatting

- Use consistent indentation (2 spaces)
- Follow TypeScript naming conventions
- Use meaningful variable names
- Add JSDoc comments for functions

### File Organization

```
backend/src/
├── index.ts           # Main server
├── config/           # Configuration files
├── functions/        # Business logic
├── modules/          # Feature modules
├── db/              # Database entities
└── types/           # TypeScript types

frontend/src/
├── App.svelte        # Main component
├── lib/             # Reusable components
├── store/           # State management
└── types/           # TypeScript types
```

## Performance Optimization

### Backend Optimization

- Use streaming for large file transfers
- Implement proper caching headers
- Optimize database queries
- Use compression middleware

### Frontend Optimization

- Vite provides automatic code splitting
- Use Svelte's reactive features efficiently
- Implement virtual scrolling for large file lists
- Optimize asset loading

## Deployment Preparation

### Building for Production

```bash
# Build both frontend and backend
npm run build

# Verify build artifacts
ls backend/dist/
ls frontend/dist/
```

### Environment Configuration

```bash
# Production .env example
NODE_ENV=production
HOST=0.0.0.0
PRIVATE_KEY=secure-production-key
ADMIN_PASSWORD=secure-admin-password
```

## Troubleshooting

### Common Development Issues

1. **TypeScript compilation errors**
   ```bash
   cd backend && npx tsc --noEmit  # Check for errors
   ```

2. **Frontend build failures**
   ```bash
   cd frontend && npm run check    # Type check
   cd frontend && npm run build    # Build check
   ```

3. **Database connection issues**
   - Check DB_PATH in .env
   - Verify directory permissions
   - Ensure SQLite is installed

4. **Authentication problems**
   - Verify OAuth credentials
   - Check JWT token generation
   - Validate password hashing

For production deployment, see [Deployment Guide](../deployment/deployment-guide.md)