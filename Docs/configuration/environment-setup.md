# ⚙️ Environment Configuration

Complete guide for configuring the NAS File Manager environment variables and settings.

## 📋 Table of Contents

- [Overview](#overview)
- [Environment File Structure](#environment-file-structure)
- [Configuration Categories](#configuration-categories)
- [Environment-Specific Settings](#environment-specific-settings)
- [Configuration Validation](#configuration-validation)
- [Path Configuration](#path-configuration)
- [Security Configuration](#security-configuration)
- [Troubleshooting](#troubleshooting)

## Overview

The NAS File Manager uses a centralized configuration system with a single `.env` file at the project root containing all environment variables for both frontend and backend components.

### Configuration Philosophy

- **Single Source of Truth**: One `.env` file for all settings
- **Environment-Aware**: Different settings for development/production
- **Platform-Agnostic**: Automatic path resolution for Windows/Linux/Docker
- **Security-First**: Secure defaults with production overrides
- **Validation**: Built-in configuration validation and warnings

## Environment File Structure

### File Location
```
nas-main/
├── .env                    # Main configuration file (you create this)
├── .env.example            # Template with all available options
├── backend/
└── frontend/
```

### Creating Configuration File
```bash
# Copy template to create your configuration
cp .env.example .env

# Edit with your specific settings
nano .env  # or your preferred editor
```

## Configuration Categories

### 1. Application Environment

#### Basic Application Settings
```env
# Application Environment
NODE_ENV=development                    # development | production
APP_NAME=NAS File Manager              # Application display name
APP_VERSION=1.0.0                      # Application version

# Server Configuration
PORT=7777                              # Backend server port
HOST=localhost                         # Server bind address (0.0.0.0 for production)
FRONTEND_PORT=5050                     # Frontend dev server port (development only)

# URLs (Environment-specific)
SERVER_URL=http://localhost:7777       # Backend API URL
FRONTEND_URL=http://localhost:5050     # Frontend URL (development)
API_BASE_URL=http://localhost:7777     # API base URL
```

#### Environment Variable Explanations

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `NODE_ENV` | Runtime environment | `development` | Yes |
| `PORT` | Backend server port | `7777` | Yes |
| `HOST` | Server bind address | `localhost` | Yes |
| `SERVER_URL` | Full server URL | Auto-generated | No |

### 2. Authentication & Security

#### Core Authentication Settings
```env
# JWT Configuration
PRIVATE_KEY=your-secure-secret-key     # JWT signing key (CHANGE IN PRODUCTION!)
JWT_EXPIRY=24h                         # Token expiration (24h, 7d, 30m, etc.)

# Authentication Strategy
AUTH_TYPE=both                         # oauth | local | both

# Admin Configuration
ADMIN_PASSWORD=your-secure-password    # Admin account password
```

#### Password Requirements (Local Authentication)
```env
# Password Policy
PASSWORD_MIN_LENGTH=8                  # Minimum password length
PASSWORD_REQUIRE_UPPERCASE=true        # Require uppercase letter
PASSWORD_REQUIRE_LOWERCASE=true        # Require lowercase letter  
PASSWORD_REQUIRE_NUMBER=true           # Require numeric character
PASSWORD_REQUIRE_SPECIAL=false         # Require special character (!@#$%^&*)
```

#### Authentication Types

| Auth Type | Description | Use Case |
|-----------|-------------|----------|
| `oauth` | Only OAuth providers | External authentication only |
| `local` | Only local ID/Password | Self-contained authentication |
| `both` | OAuth + Local auth | Maximum flexibility (recommended) |

### 3. OAuth Provider Configuration

#### Discord OAuth
```env
# Discord OAuth Configuration
DISCORD_CLIENT_ID=123456789012345678                           # Discord application client ID
DISCORD_CLIENT_SECRET=abcdefghijklmnopqrstuvwxyz123456          # Discord application secret
DISCORD_REDIRECT_URI=http://localhost:7777/login               # OAuth callback URL
DISCORD_LOGIN_URL=https://discord.com/oauth2/authorize?client_id=123456789012345678&response_type=token&redirect_uri=http://localhost:5050/login&scope=identify
```

#### Kakao OAuth
```env
# Kakao OAuth Configuration  
KAKAO_REST_API_KEY=abcdefghijklmnopqrstuvwxyz123456             # Kakao application REST API key
KAKAO_CLIENT_SECRET=zyxwvutsrqponmlkjihgfedcba654321           # Kakao application secret
KAKAO_REDIRECT_URI=http://localhost:5050/kakaoLogin            # OAuth callback URL
KAKAO_LOGIN_URL=https://kauth.kakao.com/oauth/authorize?response_type=code&client_id=abcdefghijklmnopqrstuvwxyz123456&redirect_uri=http://localhost:5050/kakaoLogin
```

### 4. Storage & File System

#### Core Storage Configuration
```env
# Storage Paths - Automatically configured based on environment
DATA_PATH=/mnt/nas-storage              # Primary storage path (production)
# DATA_PATH=../../nas-data              # Development path (Windows)
# DATA_PATH=/app/data                   # Docker container path

# File Upload Limits
MAX_FILE_SIZE=50gb                      # Maximum upload size
ALLOWED_EXTENSIONS=*                    # File type filter (* = all types)
ENABLE_STREAMING=true                   # Enable media streaming

# Advanced Storage Settings
TEMP_DIR=/tmp/nas                       # Temporary files directory
UPLOAD_TIMEOUT=300000                   # Upload timeout in milliseconds
```

#### Path Configuration Examples

**Development (Windows):**
```env
DATA_PATH=../../nas-data
```

**Production (Linux):**
```env
DATA_PATH=/mnt/nas-storage
```

**Docker Container:**
```env
DATA_PATH=/app/data
```

#### File Type Restrictions
```env
# Allow all file types (default)
ALLOWED_EXTENSIONS=*

# Restrict to specific types
ALLOWED_EXTENSIONS=jpg,jpeg,png,gif,pdf,doc,docx,txt,mp3,mp4,zip

# Common document types
ALLOWED_EXTENSIONS=pdf,doc,docx,xls,xlsx,ppt,pptx,txt,rtf

# Media files only
ALLOWED_EXTENSIONS=jpg,jpeg,png,gif,bmp,mp3,mp4,avi,mkv,wav,flac
```

### 5. Database Configuration

#### SQLite Settings
```env
# Database Configuration
DB_TYPE=sqlite                          # Database type (currently only SQLite)
DB_ENABLE_WAL=true                      # Write-Ahead Logging for performance
DB_ENABLE_FOREIGN_KEYS=true             # Foreign key constraints
DB_TIMEOUT=30000                        # Query timeout in milliseconds
```

### 6. Network & CORS Configuration

#### Network Settings
```env
# CORS Configuration
CORS_ORIGIN=*                          # Allowed origins (* = all, comma-separated for specific)
ENABLE_CORS=true                       # Enable CORS middleware

# Security Headers
SESSION_TIMEOUT=604800000              # Session timeout (7 days in milliseconds)
RATE_LIMIT_WINDOW=900000               # Rate limiting window (15 minutes)
RATE_LIMIT_MAX=100                     # Maximum requests per window
```

#### Production CORS Settings
```env
# Production CORS (specific domains only)
CORS_ORIGIN=https://yourdomain.com,https://www.yourdomain.com,https://app.yourdomain.com
```

### 7. Development & Debugging

#### Development Features
```env
# Development Configuration
DEBUG_MODE=true                        # Enable debug logging
LOG_LEVEL=debug                        # Logging level (debug | info | warn | error)
ENABLE_REQUEST_LOGGING=true            # Log all HTTP requests

# Frontend Development (Vite)
VITE_HMR_PORT=5050                     # Hot Module Replacement port
VITE_API_URL=http://localhost:7777     # API URL for frontend
```

#### Logging Levels

| Level | Description | Use Case |
|-------|-------------|----------|
| `debug` | All messages including debug info | Development |
| `info` | Informational messages | Production monitoring |
| `warn` | Warnings and above | Production (recommended) |
| `error` | Errors only | Minimal logging |

## Environment-Specific Settings

### Development Environment (.env)

**Complete development configuration:**
```env
# ═══════════════════════════════════════════════════════════════════════
# NAS FILE MANAGER - DEVELOPMENT CONFIGURATION
# ═══════════════════════════════════════════════════════════════════════

# Environment
NODE_ENV=development
PORT=7777
HOST=localhost
FRONTEND_PORT=5050

# URLs
SERVER_URL=http://localhost:7777
FRONTEND_URL=http://localhost:5050
API_BASE_URL=http://localhost:7777

# Authentication (Development defaults - CHANGE FOR PRODUCTION)
AUTH_TYPE=both
PRIVATE_KEY=development-secret-key-change-in-production
ADMIN_PASSWORD=admin123
JWT_EXPIRY=24h

# Relaxed Password Requirements (Development)
PASSWORD_MIN_LENGTH=4
PASSWORD_REQUIRE_UPPERCASE=false
PASSWORD_REQUIRE_LOWERCASE=false
PASSWORD_REQUIRE_NUMBER=false
PASSWORD_REQUIRE_SPECIAL=false

# Storage (Development paths)
DATA_PATH=../../nas-data
MAX_FILE_SIZE=10gb
ALLOWED_EXTENSIONS=*
ENABLE_STREAMING=true

# Database
DB_TYPE=sqlite
DB_ENABLE_WAL=true
DB_ENABLE_FOREIGN_KEYS=true

# Development Features
DEBUG_MODE=true
LOG_LEVEL=debug
ENABLE_CORS=true
CORS_ORIGIN=*
ENABLE_REQUEST_LOGGING=true

# OAuth (Optional for development - leave empty if not testing)
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=
KAKAO_REST_API_KEY=
KAKAO_CLIENT_SECRET=
```

### Production Environment (.env)

**Complete production configuration:**
```env
# ═══════════════════════════════════════════════════════════════════════
# NAS FILE MANAGER - PRODUCTION CONFIGURATION
# ═══════════════════════════════════════════════════════════════════════

# Environment
NODE_ENV=production
PORT=7777
HOST=0.0.0.0

# URLs (Update with your domain)
SERVER_URL=https://your-domain.com
API_BASE_URL=https://your-domain.com

# Authentication (SECURE PRODUCTION VALUES)
AUTH_TYPE=both
PRIVATE_KEY=your-very-secure-random-secret-key-minimum-32-characters-long
ADMIN_PASSWORD=your-secure-admin-password-with-special-chars
JWT_EXPIRY=24h

# Strong Password Requirements (Production)
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=true

# Storage (Production paths)
DATA_PATH=/mnt/nas-storage
MAX_FILE_SIZE=50gb
ALLOWED_EXTENSIONS=*
ENABLE_STREAMING=true

# Database
DB_TYPE=sqlite
DB_ENABLE_WAL=true
DB_ENABLE_FOREIGN_KEYS=true

# Production Logging
DEBUG_MODE=false
LOG_LEVEL=warn
ENABLE_REQUEST_LOGGING=true

# Security (Restrict CORS to your domains)
ENABLE_CORS=true
CORS_ORIGIN=https://your-domain.com,https://www.your-domain.com

# OAuth Production Configuration
DISCORD_CLIENT_ID=your-production-discord-client-id
DISCORD_CLIENT_SECRET=your-production-discord-secret
DISCORD_REDIRECT_URI=https://your-domain.com/login
KAKAO_REST_API_KEY=your-production-kakao-api-key
KAKAO_CLIENT_SECRET=your-production-kakao-secret
KAKAO_REDIRECT_URI=https://your-domain.com/kakaoLogin
```

### Docker Environment (.env)

**Docker-specific configuration:**
```env
# ═══════════════════════════════════════════════════════════════════════
# NAS FILE MANAGER - DOCKER CONFIGURATION
# ═══════════════════════════════════════════════════════════════════════

# Environment
NODE_ENV=production
PORT=7777
HOST=0.0.0.0

# Authentication
AUTH_TYPE=both
PRIVATE_KEY=your-secure-secret-key
ADMIN_PASSWORD=your-secure-password

# Storage (Docker container paths)
DATA_PATH=/app/data
MAX_FILE_SIZE=50gb

# Database (Docker path)
DB_TYPE=sqlite

# Production settings
DEBUG_MODE=false
LOG_LEVEL=info
ENABLE_CORS=true
CORS_ORIGIN=https://your-domain.com
```

## Configuration Validation

### Built-in Validation

The application automatically validates configuration on startup:

#### Error Conditions (Application won't start)
- Missing required environment variables
- Invalid authentication type
- Invalid log level
- Invalid JWT expiry format
- Missing OAuth credentials when AUTH_TYPE=oauth

#### Warning Conditions (Application starts with warnings)
- Using development defaults in production
- Weak admin password
- Overly permissive CORS settings
- Disabled security features

### Example Validation Output

```bash
✅ Configuration loaded successfully
⚠️  Warning: Using development PRIVATE_KEY in production
⚠️  Warning: CORS origin set to '*' - consider restricting in production
✅ Database connection established
✅ All storage directories verified
🚀 Server starting on http://0.0.0.0:7777
```

### Manual Configuration Check

#### Validate Configuration Script
```bash
#!/bin/bash
# validate-config.sh

echo "🔍 Validating NAS configuration..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "❌ .env file not found"
    exit 1
fi

# Source .env file
source .env

# Check required variables
REQUIRED_VARS=("NODE_ENV" "PORT" "PRIVATE_KEY" "ADMIN_PASSWORD" "AUTH_TYPE")

for var in "${REQUIRED_VARS[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required variable $var is not set"
        exit 1
    fi
done

echo "✅ All required variables are set"

# Check for development defaults in production
if [ "$NODE_ENV" = "production" ]; then
    if [ "$PRIVATE_KEY" = "development-secret-key" ] || [ "$PRIVATE_KEY" = "development-secret-key-change-in-production" ]; then
        echo "❌ Using development PRIVATE_KEY in production"
        exit 1
    fi
    
    if [ "$ADMIN_PASSWORD" = "admin123" ]; then
        echo "❌ Using development ADMIN_PASSWORD in production"
        exit 1
    fi
fi

echo "✅ Configuration validation passed"
```

## Path Configuration

### Automatic Path Resolution

The application uses intelligent path resolution based on the environment:

#### Path Resolution Logic
```typescript
// Simplified path resolution logic
const getBasePath = (): string => {
  if (process.env.DATA_PATH) {
    return process.env.DATA_PATH;
  }
  
  if (process.env.NODE_ENV === 'production') {
    if (fs.existsSync('/app/data')) {
      return '/app/data';  // Docker
    }
    return '/mnt/nas-storage';  // Linux production
  }
  
  return '../../nas-data';  // Development
};
```

#### Platform-Specific Paths

| Environment | Base Path | Description |
|-------------|-----------|-------------|
| Windows Dev | `../../nas-data` | Relative to project root |
| Linux Prod | `/mnt/nas-storage` | Absolute system path |
| Docker | `/app/data` | Container volume mount |
| Custom | `DATA_PATH` env var | User-defined path |

### Directory Structure

The application creates this directory structure automatically:
```
DATA_PATH/
├── data/           # User files
├── admin-data/     # Admin files  
├── database/       # SQLite database
└── temp/          # Temporary files
```

### Path Configuration Examples

#### Custom Storage Location
```env
# Use custom storage location
DATA_PATH=/custom/storage/location
```

#### Network Storage
```env
# Mount network storage first:
# sudo mount -t nfs server:/export /mnt/nfs-storage
DATA_PATH=/mnt/nfs-storage
```

#### RAID Storage
```env
# Configure RAID first, then:
DATA_PATH=/mnt/raid-storage
```

## Security Configuration

### Production Security Checklist

#### Required Security Settings
```env
# Strong authentication
PRIVATE_KEY=generate-32-plus-character-random-string-with-numbers-and-symbols
ADMIN_PASSWORD=SecureAdmin123!@#

# Restrictive CORS
CORS_ORIGIN=https://yourdomain.com,https://app.yourdomain.com

# Strong password requirements
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=true

# Secure defaults
DEBUG_MODE=false
LOG_LEVEL=warn
```

#### File Security
```bash
# Secure .env file permissions
chmod 600 .env
chown nas-user:nas-group .env

# Secure data directories
chmod 750 /mnt/nas-storage
chown -R nas-user:nas-group /mnt/nas-storage
```

### OAuth Security Setup

#### Discord OAuth Setup
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create new application
3. Go to OAuth2 section
4. Add redirect URI: `https://your-domain.com/login`
5. Copy Client ID and Client Secret to .env

#### Kakao OAuth Setup
1. Go to [Kakao Developers](https://developers.kakao.com/)
2. Create new application
3. Configure web platform with redirect URI
4. Copy REST API Key and Client Secret to .env

## Troubleshooting

### Common Configuration Issues

#### Application Won't Start

**Error: "Required environment variable missing"**
```bash
# Check .env file exists and has required variables
ls -la .env
grep -E "(NODE_ENV|PORT|PRIVATE_KEY)" .env
```

**Error: "Cannot load environment configuration"**
```bash
# Check file permissions
ls -la .env

# Check file format (no Windows line endings)
file .env
dos2unix .env  # If needed
```

#### Authentication Issues

**OAuth not working:**
```bash
# Check OAuth configuration
grep -E "(DISCORD|KAKAO)" .env

# Verify URLs match OAuth provider settings
# Check redirect URIs exactly match
```

**Local auth not working:**
```bash
# Check AUTH_TYPE setting
grep AUTH_TYPE .env

# Verify password requirements
grep PASSWORD_ .env
```

#### File Operation Issues

**Cannot upload files:**
```bash
# Check data directory exists and is writable
ls -la /mnt/nas-storage/
sudo chown -R nas-user:nas-group /mnt/nas-storage/
```

**File size limit errors:**
```bash
# Check MAX_FILE_SIZE setting
grep MAX_FILE_SIZE .env

# Convert to bytes for verification
# 50gb = 53687091200 bytes
```

#### Database Issues

**Database connection failed:**
```bash
# Check database directory
ls -la /mnt/nas-storage/database/

# Check SQLite installation
sqlite3 --version

# Check database file permissions
ls -la /mnt/nas-storage/database/nas.sqlite
```

### Configuration Testing

#### Test Configuration Script
```bash
#!/bin/bash
# test-config.sh

echo "🧪 Testing NAS configuration..."

# Start application in test mode
NODE_ENV=test npm start &
APP_PID=$!

sleep 10

# Test health endpoint
if curl -f http://localhost:7777/ > /dev/null 2>&1; then
    echo "✅ Application started successfully"
else
    echo "❌ Application failed to start"
    kill $APP_PID 2>/dev/null
    exit 1
fi

# Test authentication endpoint
if curl -f http://localhost:7777/auth/config > /dev/null 2>&1; then
    echo "✅ Authentication configuration accessible"
else
    echo "❌ Authentication configuration failed"
fi

# Clean up
kill $APP_PID 2>/dev/null
echo "✅ Configuration test completed"
```

### Environment Variable Reference

#### Complete Variable List

| Category | Variable | Type | Default | Description |
|----------|----------|------|---------|-------------|
| **Application** | NODE_ENV | string | development | Runtime environment |
| | PORT | number | 7777 | Server port |
| | HOST | string | localhost | Server host |
| **Auth** | AUTH_TYPE | enum | both | Authentication type |
| | PRIVATE_KEY | string | dev-key | JWT signing key |
| | ADMIN_PASSWORD | string | admin123 | Admin password |
| | JWT_EXPIRY | string | 24h | Token expiration |
| **Storage** | DATA_PATH | string | auto | Storage base path |
| | MAX_FILE_SIZE | string | 50gb | Upload size limit |
| | ALLOWED_EXTENSIONS | string | * | File type filter |
| **Security** | CORS_ORIGIN | string | * | CORS allowed origins |
| | PASSWORD_MIN_LENGTH | number | 8 | Min password length |
| **Logging** | DEBUG_MODE | boolean | true | Enable debug logs |
| | LOG_LEVEL | enum | info | Logging verbosity |

For additional configuration details, see [Authentication Setup](authentication.md) and [Storage Configuration](storage-config.md).

---

*This configuration guide covers all environment settings. For deployment-specific configuration, see the [Deployment Guide](../deployment/deployment-guide.md).*