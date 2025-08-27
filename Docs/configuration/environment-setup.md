# ⚙️ Environment Configuration Guide

Complete guide for configuring the NAS File Manager environment variables and settings.

## Overview

The NAS File Manager uses a centralized configuration system with a single `.env` file at the project root containing all environment variables for both frontend and backend.

## Environment File Structure

### Location
```
nas-main/
├── .env                    # Main configuration file
├── backend/
└── frontend/
```

### Complete Configuration Template

```env
# ═══════════════════════════════════════════════════════════════════════
# NAS PROJECT - CENTRALIZED CONFIGURATION
# Single source of truth for all environment settings
# ═══════════════════════════════════════════════════════════════════════

# ───────────────────────────────────────────────────────────────────────
# ENVIRONMENT & DEPLOYMENT
# ───────────────────────────────────────────────────────────────────────
NODE_ENV=development
APP_NAME=NAS File Manager
APP_VERSION=1.0.0

# Server Configuration
PORT=7777
HOST=localhost
FRONTEND_PORT=5050

# URLs (Environment-specific)
SERVER_URL=http://localhost:7777
FRONTEND_URL=http://localhost:5050
API_BASE_URL=http://localhost:7777

# ───────────────────────────────────────────────────────────────────────
# AUTHENTICATION & SECURITY
# ───────────────────────────────────────────────────────────────────────

# JWT Configuration
PRIVATE_KEY=development-secret-key
JWT_EXPIRY=24h

# Authentication Strategy (oauth, local, both)
AUTH_TYPE=both

# Admin Configuration
ADMIN_PASSWORD=admin123

# Password Requirements (Local Auth)
PASSWORD_MIN_LENGTH=4
PASSWORD_REQUIRE_UPPERCASE=false
PASSWORD_REQUIRE_LOWERCASE=false
PASSWORD_REQUIRE_NUMBER=false
PASSWORD_REQUIRE_SPECIAL=false

# ───────────────────────────────────────────────────────────────────────
# OAUTH PROVIDERS
# ───────────────────────────────────────────────────────────────────────

# Discord OAuth
DISCORD_CLIENT_ID=your-discord-client-id
DISCORD_CLIENT_SECRET=your-discord-client-secret
DISCORD_REDIRECT_URI=http://localhost:7777/login
DISCORD_LOGIN_URL=https://discord.com/oauth2/authorize?client_id=your-discord-client-id&response_type=token&redirect_uri=http://localhost:5050/login&scope=identify

# Kakao OAuth  
KAKAO_REST_API_KEY=your-kakao-dev-key
KAKAO_CLIENT_SECRET=your-kakao-dev-secret
KAKAO_REDIRECT_URI=http://localhost:5050/kakaoLogin
KAKAO_LOGIN_URL=https://kauth.kakao.com/oauth/authorize?response_type=code&client_id=your-kakao-dev-key&redirect_uri=http://localhost:5050/kakaoLogin

# ───────────────────────────────────────────────────────────────────────
# STORAGE & FILE SYSTEM
# ───────────────────────────────────────────────────────────────────────

# Storage Paths (Cross-platform - auto-detected in development)
NAS_DATA_DIR=../../nas-data
NAS_ADMIN_DATA_DIR=../../nas-data-admin
DB_PATH=./db/nas.sqlite
NAS_TEMP_DIR=/tmp/nas

# Storage Configuration
MAX_FILE_SIZE=10gb
ALLOWED_EXTENSIONS=*
ENABLE_STREAMING=true

# System Configuration
DISK_PATH=C:\
CORS_ORIGIN=*
SESSION_TIMEOUT=604800000

# ───────────────────────────────────────────────────────────────────────
# DATABASE
# ───────────────────────────────────────────────────────────────────────
DB_TYPE=sqlite
DB_ENABLE_WAL=true
DB_ENABLE_FOREIGN_KEYS=true

# ───────────────────────────────────────────────────────────────────────
# DEVELOPMENT & DEBUGGING
# ───────────────────────────────────────────────────────────────────────
DEBUG_MODE=true
LOG_LEVEL=info
ENABLE_CORS=true
ENABLE_REQUEST_LOGGING=true

# Frontend Development
VITE_HMR_PORT=5050
VITE_API_URL=http://localhost:7777
```

## Configuration Categories

### 1. Environment & Deployment

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NODE_ENV` | Application environment | `development` | Yes |
| `APP_NAME` | Application display name | `NAS File Manager` | No |
| `APP_VERSION` | Application version | `1.0.0` | No |
| `PORT` | Backend server port | `7777` | Yes |
| `HOST` | Server bind address | `localhost` | Yes |
| `FRONTEND_PORT` | Frontend dev server port | `5050` | Yes |

### 2. Authentication & Security

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PRIVATE_KEY` | JWT signing key | `development-secret-key` | Yes |
| `JWT_EXPIRY` | JWT token expiration | `24h` | No |
| `AUTH_TYPE` | Authentication method | `both` | Yes |
| `ADMIN_PASSWORD` | Admin account password | `admin123` | Yes |

**Authentication Types:**
- `oauth`: Only OAuth providers (Discord/Kakao)
- `local`: Only local ID/Password
- `both`: Both OAuth and local authentication

### 3. Password Requirements

| Variable | Description | Default | Notes |
|----------|-------------|---------|-------|
| `PASSWORD_MIN_LENGTH` | Minimum password length | `4` | Development: 4, Production: 8+ |
| `PASSWORD_REQUIRE_UPPERCASE` | Require uppercase letter | `false` | Boolean |
| `PASSWORD_REQUIRE_LOWERCASE` | Require lowercase letter | `false` | Boolean |
| `PASSWORD_REQUIRE_NUMBER` | Require number | `false` | Boolean |
| `PASSWORD_REQUIRE_SPECIAL` | Require special character | `false` | Boolean |

### 4. OAuth Configuration

#### Discord OAuth
```env
DISCORD_CLIENT_ID=123456789012345678
DISCORD_CLIENT_SECRET=abcdefghijklmnopqrstuvwxyz123456
DISCORD_REDIRECT_URI=http://localhost:7777/login
DISCORD_LOGIN_URL=https://discord.com/oauth2/authorize?client_id=123456789012345678&response_type=token&redirect_uri=http://localhost:5050/login&scope=identify
```

#### Kakao OAuth
```env
KAKAO_REST_API_KEY=abcdefghijklmnopqrstuvwxyz123456
KAKAO_CLIENT_SECRET=zyxwvutsrqponmlkjihgfedcba654321
KAKAO_REDIRECT_URI=http://localhost:5050/kakaoLogin
KAKAO_LOGIN_URL=https://kauth.kakao.com/oauth/authorize?response_type=code&client_id=abcdefghijklmnopqrstuvwxyz123456&redirect_uri=http://localhost:5050/kakaoLogin
```

### 5. Storage & File System

| Variable | Description | Default | Platform |
|----------|-------------|---------|----------|
| `NAS_DATA_DIR` | User data directory | `../../nas-data` | Windows Dev |
|  |  | `/home/nas/nas-storage/data` | Linux Prod |
|  |  | `/app/data` | Docker |
| `NAS_ADMIN_DATA_DIR` | Admin data directory | `../../nas-data-admin` | Windows Dev |
| `DB_PATH` | SQLite database path | `./db/nas.sqlite` | All |
| `NAS_TEMP_DIR` | Temporary files directory | `/tmp/nas` | Linux/Docker |
| `MAX_FILE_SIZE` | Maximum upload size | `10gb` | All |
| `ALLOWED_EXTENSIONS` | File extension filter | `*` | All |

### 6. System Configuration

| Variable | Description | Default | Notes |
|----------|-------------|---------|-------|
| `DISK_PATH` | System disk path | `C:\` (Windows), `/` (Linux) | Platform-specific |
| `CORS_ORIGIN` | CORS allowed origins | `*` | Production: specific domains |
| `SESSION_TIMEOUT` | Session timeout (ms) | `604800000` | 7 days |

## Environment-Specific Configurations

### Development Environment

```env
NODE_ENV=development
HOST=localhost
DEBUG_MODE=true
PRIVATE_KEY=development-secret-key
ADMIN_PASSWORD=admin123
PASSWORD_MIN_LENGTH=4
NAS_DATA_DIR=../../nas-data
CORS_ORIGIN=*
```

### Production Environment

```env
NODE_ENV=production
HOST=0.0.0.0
DEBUG_MODE=false
PRIVATE_KEY=your-very-secure-secret-key-here
ADMIN_PASSWORD=your-secure-admin-password
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
NAS_DATA_DIR=/home/nas/storage/data
MAX_FILE_SIZE=50gb
CORS_ORIGIN=https://your-domain.com
```

### Docker Environment

```env
NODE_ENV=production
HOST=0.0.0.0
NAS_DATA_DIR=/app/data
NAS_ADMIN_DATA_DIR=/app/admin-data
DB_PATH=/app/db
NAS_TEMP_DIR=/tmp/nas
```

## Configuration Loading

### Backend Configuration Loading

The backend loads configuration from `backend/src/config/environment.ts`:

```typescript
import { config } from "dotenv";
import { join } from "path";

// Load from root .env file only
const rootEnvPath = join(__dirname, '../../../.env');
const result = config({ path: rootEnvPath });

export class Environment {
  static readonly NODE_ENV = process.env.NODE_ENV || 'development';
  static readonly PORT = parseInt(process.env.PORT || '7777');
  static readonly PRIVATE_KEY = process.env.PRIVATE_KEY || 'development-secret-key';
  // ... other configurations
  
  static validate(): void {
    // Validation logic for required fields
  }
}
```

### Frontend Configuration Loading

The frontend uses Vite's `loadEnv` in `frontend/vite.config.ts`:

```typescript
import { loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  
  return {
    define: {
      __SERVER_URL__: JSON.stringify(env.SERVER_URL || 'http://localhost:7777'),
      __DISCORD_LOGIN_URL__: JSON.stringify(env.DISCORD_LOGIN_URL || ''),
      __KAKAO_LOGIN_URL__: JSON.stringify(env.KAKAO_LOGIN_URL || ''),
    }
  };
});
```

## Validation and Security

### Environment Validation

The system automatically validates configuration on startup:

```typescript
// Validation warnings and errors
if (!Environment.PRIVATE_KEY || Environment.PRIVATE_KEY === 'development-secret-key') {
  if (Environment.IS_PRODUCTION) {
    errors.push('PRIVATE_KEY must be set to a secure value in production');
  } else {
    warnings.push('PRIVATE_KEY is using development default');
  }
}
```

### Security Best Practices

1. **Production Secrets**: Never use development defaults in production
2. **Key Strength**: Use strong, unique keys for PRIVATE_KEY
3. **Password Complexity**: Enable password requirements in production
4. **CORS Configuration**: Restrict CORS_ORIGIN to specific domains
5. **File Permissions**: Secure `.env` file with appropriate permissions

```bash
chmod 600 .env  # Read/write for owner only
```

## OAuth Provider Setup

### Discord OAuth Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create new application
3. Go to OAuth2 section
4. Add redirect URI: `http://localhost:7777/login` (development)
5. Get Client ID and Client Secret
6. Update .env:

```env
DISCORD_CLIENT_ID=your-client-id
DISCORD_CLIENT_SECRET=your-client-secret
DISCORD_REDIRECT_URI=http://localhost:7777/login
```

### Kakao OAuth Setup

1. Go to [Kakao Developers](https://developers.kakao.com/)
2. Create new application
3. Go to App Settings > Platform
4. Add Web platform with redirect URI
5. Get REST API Key and Client Secret
6. Update .env:

```env
KAKAO_REST_API_KEY=your-rest-api-key
KAKAO_CLIENT_SECRET=your-client-secret
KAKAO_REDIRECT_URI=http://localhost:5050/kakaoLogin
```

## Troubleshooting

### Common Configuration Issues

1. **Environment not loaded**
   ```
   ⚠️ Root .env file not found, using process.env only
   ```
   - Check .env file exists in project root
   - Verify file permissions

2. **OAuth not working**
   - Verify OAuth credentials are correct
   - Check redirect URIs match exactly
   - Ensure URLs are accessible

3. **Database connection issues**
   - Check DB_PATH directory exists
   - Verify write permissions for database directory

4. **File upload issues**
   - Check NAS_DATA_DIR path exists
   - Verify directory write permissions
   - Check MAX_FILE_SIZE setting

### Environment Debugging

Enable debug logging:

```env
DEBUG_MODE=true
LOG_LEVEL=debug
ENABLE_REQUEST_LOGGING=true
```

View configuration on startup:
- Backend will log configuration validation results
- Check for warnings about default values
- Verify all required fields are loaded

For deployment-specific configuration, see [Deployment Guide](../deployment/deployment-guide.md)