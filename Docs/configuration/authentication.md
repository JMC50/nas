# 🔐 Authentication Configuration

Authentication guide for the NAS File Manager - from simple admin login to advanced OAuth integration.

## 📋 Table of Contents

- [🎯 Quick Start (Admin Login)](#quick-start-admin-login)
- [Local User Management](#local-user-management)
- [🔧 Advanced: OAuth Providers](#advanced-oauth-providers)  
- [Permission System](#permission-system)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## 🎯 Quick Start (Admin Login)

**Your NAS works immediately** with the admin credentials from your `.env` file - no additional authentication setup required.

### Default Authentication

With your basic `.env` configuration:
```env
ADMIN_PASSWORD=your-secure-password
```

**You can immediately:**
- Login as admin using your `ADMIN_PASSWORD`
- Access all files and features
- Create additional local users if needed
- Full system administration

### Accessing Your NAS

1. **Navigate to**: http://localhost:7777
2. **Login with**: 
   - Username: `admin` 
   - Password: Your `ADMIN_PASSWORD` from `.env`
3. **Done!** - Full access to your NAS system

**No OAuth setup required** - the system works immediately with secure admin authentication.

## Local Authentication

### Basic Configuration

```env
# Enable local authentication
AUTH_TYPE=local  # or 'both' to include OAuth

# Admin configuration
ADMIN_PASSWORD=your-secure-admin-password

# JWT settings
PRIVATE_KEY=your-secure-jwt-signing-key
JWT_EXPIRY=24h
```

### Password Requirements

#### Development Settings
```env
# Relaxed requirements for development
PASSWORD_MIN_LENGTH=4
PASSWORD_REQUIRE_UPPERCASE=false
PASSWORD_REQUIRE_LOWERCASE=false
PASSWORD_REQUIRE_NUMBER=false
PASSWORD_REQUIRE_SPECIAL=false
```

#### Production Settings
```env
# Strong requirements for production
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=true
```

### Password Validation Rules

The system validates passwords against these criteria:

| Requirement | Description | Example |
|-------------|-------------|---------|
| Min Length | Minimum character count | `PASSWORD_MIN_LENGTH=12` |
| Uppercase | At least one uppercase letter | A-Z |
| Lowercase | At least one lowercase letter | a-z |
| Number | At least one numeric digit | 0-9 |
| Special | At least one special character | `!@#$%^&*()_+-=[]{}|;:,.<>?` |

### Local User Registration

#### API Endpoint
```http
POST /auth/register
Content-Type: application/json

{
  "userId": "username",
  "password": "SecurePassword123!"
}
```

#### Response
```json
{
  "success": true,
  "message": "User registered successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Password Validation Errors
```json
{
  "success": false,
  "error": "Password must be at least 12 characters long"
}

{
  "success": false,
  "error": "Password must contain at least one uppercase letter"
}
```

### Local User Login

#### API Endpoint
```http
POST /auth/login
Content-Type: application/json

{
  "userId": "username",
  "password": "SecurePassword123!"
}
```

#### Successful Response
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "userId": "username",
    "authType": "local",
    "isAdmin": false
  }
}
```

### Password Management

#### Change Password
```http
POST /auth/change-password?token=jwt_token
Content-Type: application/json

{
  "currentPassword": "OldPassword123!",
  "newPassword": "NewSecurePassword456!"
}
```

#### Password Reset (Admin)
Administrators can reset user passwords through the user management interface.

## 🔧 Advanced: OAuth Providers

**For advanced users only** - most users can use the simple admin authentication above. OAuth setup is completely optional.

### Why Use OAuth?

OAuth providers are useful when you want:
- Users to login with their Discord/Kakao accounts
- Integration with external user management systems
- Social login convenience for multiple users

### Supported Providers

The system supports OAuth authentication with:
- **Discord**: Discord OAuth 2.0
- **Kakao**: Kakao OAuth 2.0

### Discord OAuth Setup

#### 1. Create Discord Application
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click "New Application"
3. Enter application name (e.g., "My NAS System")
4. Go to "OAuth2" section

#### 2. Configure OAuth2 Settings
- **Redirect URIs**: Add your callback URL
  - Development: `http://localhost:7777/login`
  - Production: `https://your-domain.com/login`
- **Scopes**: Select `identify` (to read user information)

#### 3. Environment Configuration
```env
# Discord OAuth Configuration
DISCORD_CLIENT_ID=123456789012345678
DISCORD_CLIENT_SECRET=abcdefghijklmnopqrstuvwxyz123456
DISCORD_REDIRECT_URI=https://your-domain.com/login
DISCORD_LOGIN_URL=https://discord.com/oauth2/authorize?client_id=123456789012345678&response_type=token&redirect_uri=https://your-domain.com/login&scope=identify
```

#### 4. Frontend Integration
The frontend automatically generates Discord login URLs based on environment configuration.

### Kakao OAuth Setup

#### 1. Create Kakao Application
1. Go to [Kakao Developers](https://developers.kakao.com/)
2. Create new application
3. Go to "App Settings" → "Platform"
4. Add Web platform

#### 2. Configure Platform Settings
- **Site Domain**: Your domain (e.g., `https://your-domain.com`)
- **Redirect Path**: `/kakaoLogin`

#### 3. Environment Configuration
```env
# Kakao OAuth Configuration
KAKAO_REST_API_KEY=abcdefghijklmnopqrstuvwxyz123456
KAKAO_CLIENT_SECRET=zyxwvutsrqponmlkjihgfedcba654321
KAKAO_REDIRECT_URI=https://your-domain.com/kakaoLogin
KAKAO_LOGIN_URL=https://kauth.kakao.com/oauth/authorize?response_type=code&client_id=abcdefghijklmnopqrstuvwxyz123456&redirect_uri=https://your-domain.com/kakaoLogin
```

### OAuth Flow

#### 1. User Initiates Login
User clicks OAuth provider login button in frontend.

#### 2. Redirect to Provider
Browser redirects to OAuth provider's authorization page.

#### 3. User Authorizes
User grants permission to the application.

#### 4. Authorization Code
Provider redirects back to application with authorization code.

#### 5. Token Exchange
Application exchanges authorization code for access token.

#### 6. User Information
Application fetches user information from provider.

#### 7. JWT Generation
Application creates JWT token for user session.

#### 8. Frontend Redirect
User is redirected to frontend with JWT token.

### OAuth Endpoints

#### Discord Callback
```http
GET /login?code=discord_authorization_code
```

#### Kakao Callback
```http
GET /kakaoLogin?code=kakao_authorization_code
```

### OAuth Error Handling

#### Common OAuth Errors
- **Invalid Client**: Check client ID and secret
- **Redirect URI Mismatch**: Verify redirect URIs match exactly
- **Scope Issues**: Ensure required scopes are granted
- **Network Issues**: Check connectivity to OAuth providers

## User Management

### User Database Schema

```typescript
interface User {
  userId: string;          // Unique user identifier
  password?: string;       // Hashed password (local auth only)
  authType: 'local' | 'oauth';
  discordId?: string;      // Discord user ID (if Discord auth)
  kakaoId?: string;        // Kakao user ID (if Kakao auth)
  createdAt: Date;
  lastLogin?: Date;
}
```

### User Registration Flow

#### Local Registration
1. User submits username and password
2. Password validated against requirements
3. Password hashed using bcrypt
4. User record created in database
5. JWT token generated and returned

#### OAuth Registration
1. User authorizes with OAuth provider
2. Application receives user information
3. User record created with OAuth ID
4. JWT token generated and returned

### Admin User Setup

#### Initial Admin Setup
1. Set admin password in environment:
   ```env
   ADMIN_PASSWORD=your-secure-admin-password
   ```

2. Register any user account (local or OAuth)

3. Request admin privileges:
   ```http
   POST /requestAdminIntent?token=user_jwt_token
   Content-Type: application/json
   
   {
     "adminPassword": "your-secure-admin-password"
   }
   ```

## Permission System

### Intent-Based Permissions

The system uses "intents" to control user permissions:

| Intent | Description | Default |
|--------|-------------|---------|
| `ADMIN` | Administrative access | No |
| `VIEW` | View files and directories | Yes |
| `OPEN` | Open and read files | Yes |
| `DOWNLOAD` | Download files | Yes |
| `UPLOAD` | Upload files | No |
| `COPY` | Copy/move files | No |
| `DELETE` | Delete files | No |
| `RENAME` | Rename files | No |

### Permission Management

#### Check User Permissions
```http
GET /getIntents?token=jwt_token&userId=target_user_id
```

#### Grant Permission (Admin Only)
```http
GET /authorize?token=admin_jwt_token&userId=target_user&intent=UPLOAD
```

#### Revoke Permission (Admin Only)
```http
GET /unauthorize?token=admin_jwt_token&userId=target_user&intent=UPLOAD
```

#### Check Specific Permission
```http
GET /checkIntent?token=jwt_token&intent=UPLOAD
```

### Default Permission Profiles

#### New User (Default)
- `VIEW`: Yes
- `OPEN`: Yes
- `DOWNLOAD`: Yes
- All other intents: No

#### Admin User
- All intents: Yes

### Custom Permission Profiles

Create custom permission profiles for different user types:

```typescript
// Example: Create "Editor" profile
const editorPermissions = ['VIEW', 'OPEN', 'DOWNLOAD', 'UPLOAD', 'RENAME'];

// Example: Create "Viewer" profile  
const viewerPermissions = ['VIEW', 'OPEN'];
```

## Security Best Practices

### JWT Security

#### Secure Key Generation
```bash
# Generate secure private key
openssl rand -base64 32

# Or use Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

#### JWT Configuration
```env
# Use long, random key
PRIVATE_KEY=generated-secure-key-from-above

# Set appropriate expiry
JWT_EXPIRY=24h  # 24 hours
# JWT_EXPIRY=7d   # 7 days
# JWT_EXPIRY=30m  # 30 minutes
```

### Password Security

#### Strong Password Policy
```env
# Production password requirements
PASSWORD_MIN_LENGTH=12
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBER=true
PASSWORD_REQUIRE_SPECIAL=true
```

#### Password Hashing
- Uses bcrypt with salt rounds
- Passwords are never stored in plain text
- Hash comparison is timing-attack resistant

### OAuth Security

#### Secure Redirect URIs
- Use HTTPS in production
- Match exact URLs (no wildcards)
- Validate state parameters (if implemented)

#### Client Secret Protection
- Store secrets in environment variables
- Never commit secrets to version control
- Rotate secrets regularly

### Session Security

#### Token Validation
- JWT signature verification
- Token expiration checking
- Automatic token refresh (if implemented)

#### Session Management
- Tokens are stateless
- No server-side session storage
- Client-side token storage in secure cookies (recommended)

## Troubleshooting

### Common Authentication Issues

#### Local Authentication Problems

**Error: "Password does not meet requirements"**
```bash
# Check password requirements in .env
grep PASSWORD_ .env

# Test password validation manually
```

**Error: "User already exists"**
```bash
# Check if user is already registered
sqlite3 database.db "SELECT userId FROM users WHERE userId='username';"
```

#### OAuth Authentication Problems

**Error: "OAuth provider error"**
```bash
# Check OAuth configuration
grep -E "(DISCORD|KAKAO)" .env

# Verify redirect URIs match exactly
curl -I "https://your-domain.com/login"
```

**Error: "Invalid client credentials"**
```bash
# Test OAuth provider connection
curl -X POST "https://discord.com/api/oauth2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=$DISCORD_CLIENT_ID&client_secret=$DISCORD_CLIENT_SECRET&grant_type=client_credentials"
```

#### JWT Token Issues

**Error: "Invalid token"**
```bash
# Check JWT configuration
grep -E "(PRIVATE_KEY|JWT_EXPIRY)" .env

# Verify token structure (decode JWT without verification)
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...." | cut -d. -f2 | base64 -d
```

### Permission System Issues

**Error: "Insufficient permissions"**
```bash
# Check user permissions
curl "http://localhost:7777/getIntents?token=jwt_token&userId=username"

# Grant missing permissions (as admin)
curl "http://localhost:7777/authorize?token=admin_token&userId=username&intent=UPLOAD"
```

### Database Issues

**Error: "Authentication database error"**
```bash
# Check database file exists and is writable
ls -la database/nas.sqlite

# Check database schema
sqlite3 database/nas.sqlite ".schema users"

# Test database connection
sqlite3 database/nas.sqlite "SELECT COUNT(*) FROM users;"
```

### Configuration Validation

#### Authentication Config Check Script
```bash
#!/bin/bash
# check-auth-config.sh

echo "🔍 Checking authentication configuration..."

source .env

# Check required variables
if [ -z "$AUTH_TYPE" ]; then
    echo "❌ AUTH_TYPE not set"
    exit 1
fi

if [ -z "$PRIVATE_KEY" ] || [ "$PRIVATE_KEY" = "development-secret-key" ]; then
    echo "⚠️ PRIVATE_KEY should be changed from default"
fi

if [ "$AUTH_TYPE" = "local" ] || [ "$AUTH_TYPE" = "both" ]; then
    if [ -z "$ADMIN_PASSWORD" ] || [ "$ADMIN_PASSWORD" = "admin123" ]; then
        echo "⚠️ ADMIN_PASSWORD should be changed from default"
    fi
fi

if [ "$AUTH_TYPE" = "oauth" ] || [ "$AUTH_TYPE" = "both" ]; then
    if [ -z "$DISCORD_CLIENT_ID" ] && [ -z "$KAKAO_REST_API_KEY" ]; then
        echo "⚠️ No OAuth providers configured"
    fi
fi

echo "✅ Authentication configuration check complete"
```

### Testing Authentication

#### Test Authentication Script
```bash
#!/bin/bash
# test-auth.sh

echo "🧪 Testing authentication system..."

# Test local registration
echo "Testing local registration..."
curl -X POST http://localhost:7777/auth/register \
  -H "Content-Type: application/json" \
  -d '{"userId":"testuser","password":"TestPassword123!"}' \
  -w "%{http_code}\n"

# Test local login
echo "Testing local login..."
TOKEN=$(curl -s -X POST http://localhost:7777/auth/login \
  -H "Content-Type: application/json" \
  -d '{"userId":"testuser","password":"TestPassword123!"}' | jq -r '.token')

if [ "$TOKEN" != "null" ]; then
    echo "✅ Local authentication working"
    
    # Test token validation
    echo "Testing token validation..."
    curl -s "http://localhost:7777/checkIntent?token=$TOKEN&intent=VIEW" | jq '.'
else
    echo "❌ Local authentication failed"
fi

echo "🧪 Authentication test complete"
```

---

*For additional security configuration, see [Environment Setup](environment-setup.md). For user management procedures, see [Troubleshooting Guide](../operations/troubleshooting.md).*