# 📡 API Reference

Complete API documentation for the NAS File Manager backend endpoints.

## 📋 Table of Contents

- [Authentication](#authentication)
- [File Operations](#file-operations)
- [Media Operations](#media-operations)
- [Archive Operations](#archive-operations)
- [User Management](#user-management)
- [System Information](#system-information)
- [Response Formats](#response-formats)
- [Error Handling](#error-handling)

## Authentication

All API endpoints (except authentication endpoints) require JWT authentication via query parameter or Authorization header.

### Authentication Methods

**Query Parameter**
```
GET /api/endpoint?token=your_jwt_token
```

**Authorization Header**
```
GET /api/endpoint
Authorization: Bearer your_jwt_token
```

### JWT Token Structure
```typescript
interface JwtPayload {
  userId: string;
  expires_in: number;  // Expiration, Unix time in milliseconds
  issued_at?: number;  // Issue time, Unix time in milliseconds. Tokens from before v0.7.0 do not have it
}
```

The server also rejects a valid token with `401` when it was issued before the `tokens_valid_after` time of its user. This time is empty until a back-channel logout sets it (see [OpenID Connect Single Sign-On](#openid-connect-single-sign-on)).

---

## Authentication Endpoints

### Get Authentication Configuration
```http
GET /auth/config
```

**Response:**
```json
{
  "authType": "both",          // "oauth" | "local" | "both"
  "providers": {
    "discord": true,
    "kakao": true,
    "local": true
  },
  "passwordRequirements": {
    "minLength": 8,
    "requireUppercase": true,
    "requireLowercase": true,
    "requireNumber": true,
    "requireSpecial": false
  }
}
```

### Local Authentication

#### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "userId": "username",
  "password": "password123"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "User registered successfully",
  "token": "jwt_token_here"
}
```

#### Login User
```http
POST /auth/login
Content-Type: application/json

{
  "userId": "username", 
  "password": "password123"
}
```

**Response (Success):**
```json
{
  "success": true,
  "token": "jwt_token_here",
  "user": {
    "userId": "username",
    "authType": "local"
  }
}
```

#### Change Password
```http
POST /auth/change-password?token=jwt_token
Content-Type: application/json

{
  "currentPassword": "old_password",
  "newPassword": "new_password"
}
```

### OAuth Authentication

#### Discord OAuth Callback
```http
GET /login?code=discord_auth_code
```

#### Kakao OAuth Callback
```http
GET /kakaoLogin?code=kakao_auth_code
```

Both OAuth endpoints redirect to frontend with token or error parameters.

### OpenID Connect Single Sign-On

These routes exist only when `AUTH_TYPE=oidc`. In the other modes they answer 404. Setup and rules are in [Single Sign-On with OpenID Connect](../configuration/oidc-sso.md).

In this mode `GET /auth/config` reports only OIDC:

```json
{
  "authType": "oidc",
  "localAuthEnabled": false,
  "discordEnabled": false,
  "googleEnabled": false,
  "oauthEnabled": false,
  "oidcEnabled": true,
  "oidcLoginUrl": "/server/auth/oidc/login",
  "oidcLogoutUrl": "/server/auth/oidc/logout"
}
```

The response also has `discordLoginUrl`, `googleLoginUrl` (both empty), and `passwordRequirements`. The three `oidc*` fields are not in the response of the other modes.

#### Start Sign-In
```http
GET /auth/oidc/login
```

**Response:** `302` to the authorization endpoint of the issuer, with `response_type=code`, `state`, `nonce`, `code_challenge` (`S256`), and the configured `redirect_uri`. The response sets the cookie `__Host-nas_oidc_login` (`Secure`, `HttpOnly`, `SameSite=Lax`, 10 minutes). Query parameters of this request do not change a redirect target.

#### Callback
```http
GET /auth/oidc/callback?code=authorization_code&state=state
```

**Response:** always `302` to a page of the NAS.

| Location | Meaning |
|----------|---------|
| `/oidcLogin` | Sign-in is complete. The response sets the cookie `__Host-nas_oidc_session` (`Secure`, `HttpOnly`, `SameSite=Lax`, 60 seconds) with the NAS token |
| `/oidcLogin?error=denied` | The user is in none of the NAS groups, or the issuer answered `access_denied`. No user is created |
| `/oidcLogin?error=failed` | The state, the code, or the ID token (signature, issuer, audience, expiry, nonce) is not valid |
| `/oidcLogin?error=unavailable` | The issuer did not answer |
| `/oidcLogin?error=conflict` | A user with the same `userId` exists without the `oidc` identity |

#### Get the Token After the Callback
```http
POST /auth/oidc/session
Cookie: __Host-nas_oidc_session=...
```

**Response:**
```json
{
  "userId": "2f9c1e5a7b3d4c6e8a0b",
  "username": "alice",
  "krname": "",
  "global_name": "Alice",
  "token": "jwt_token_here"
}
```

The response removes the cookie. Without a valid cookie the answer is `401`.

#### Sign Out
```http
GET /auth/oidc/logout
```

**Response:** `302` to `OIDC_LOGOUT_URL`, or to the `end_session_endpoint` of the issuer with `client_id`. If the issuer has no such endpoint, the target is `/localLogin?signedout=1`. The front end removes the NAS token before it opens this address.

#### Back-Channel Logout
```http
POST /auth/oidc/backchannel
Content-Type: application/x-www-form-urlencoded

logout_token=jwt_from_the_issuer
```

The issuer calls this route. The NAS verifies the `logout_token` (signature, `iss`, `aud`, `exp`, `iat` not older than 5 minutes, the back-channel logout event, no `nonce`, a `sub`, and a `jti` that was not used before). Then it sets `tokens_valid_after` of that user to the current time, and each token of the user that was issued earlier answers `401`.

| Status | Meaning |
|--------|---------|
| 200 | Accepted. Also when no NAS user has this `sub` |
| 400 | The `logout_token` is missing, not valid, too old, or already used. The body is larger than 8 KiB |
| 500 | The time could not be saved. The issuer can send the same token again |
| 503 | The NAS could not read the discovery document or the keys of the issuer, so the token was not checked. The issuer can send the same token again |

#### Routes Closed in This Mode

`POST /auth/register`, `POST /auth/login`, `POST /auth/change-password`, `GET /auth/discord/callback`, `GET /auth/google/callback`, `POST /auth/discord/register`, `POST /auth/google/register`, `POST /auth/link/start`, `POST /auth/link/discord/complete`, `POST /auth/link/google/complete`, `DELETE /auth/identities/{provider}`, and `POST /requestAdminIntent` answer `403`. The link, identity, and admin password routes still answer `401` first when the request has no valid token.

```json
{ "success": false, "message": "This is disabled. This server allows single sign-on only." }
```

---

## File Operations

### Get File/Directory Information
```http
GET /stat?token=jwt_token&loc=/path/to/item&name=filename
```

**Parameters:**
- `loc`: Directory path (URL encoded)
- `name`: File/directory name

**Response:**
```json
{
  "isDirectory": false,
  "size": 1024000,
  "formattedSize": "1.0 MB",
  "lastModified": "2025-08-28T10:30:00.000Z",
  "formattedDate": "2025-08-28 10:30:00"
}
```

### Download File
```http
GET /download?token=jwt_token&loc=/path/to/file&name=filename
```

**Parameters:**
- `loc`: File directory path
- `name`: Filename
- `range` (optional): Byte range for partial content

**Response:** File download with appropriate headers
- `Content-Type`: Detected MIME type
- `Content-Disposition`: attachment; filename="filename"
- `Content-Range`: For partial content requests

### Download Several Items as One Zip

The browser downloads a zip in two steps. The first request sends the selection in a JSON body, so a long list does not depend on the URL length. The second request is a plain `GET`, so the browser saves the stream with its normal download progress. Both requests need a valid token and the `DOWNLOAD` intent.

```http
POST /zip-downloads
Authorization: Bearer jwt_token
Content-Type: application/json

{
  "items": [
    { "loc": "/photos", "name": "2026" },
    { "loc": "/photos", "name": "cover.png" }
  ]
}
```

**Response:**
```json
{
  "path": "/zip-downloads/<id>",
  "name": "photos.zip",
  "entries": 42,
  "bytes": 123456789
}
```

```http
GET /zip-downloads/<id>?token=jwt_token
```

**Response:** `application/zip` stream with `Content-Disposition: attachment`. The stream has no `Content-Length` and does not support range requests.

**Rules:**
- An item is a file or a folder. Each name must not be empty, `.`, `..`, or contain `/` or `\`.
- The address is valid for 2 minutes and only for the user who made it. Another user gets 404.
- Entry names start at the selected item. Folder structure and empty folders are kept. Names are stored as UTF-8.
- If two selected items have the same name, the later one gets a suffix such as `report (2).txt`.
- The file name is `<item>.zip` for one item, `<folder>.zip` for several items from one folder, and `nas-download.zip` otherwise.
- Symbolic links follow the rules in [Storage Configuration](../configuration/storage-config.md#symbolic-links-and-public-download-links).
- One user can run 2 zip downloads at the same time. The server runs 3 at the same time.
- The stream stops when the client disconnects, when the client reads nothing for 30 seconds, or when the user loses the `DOWNLOAD` intent.

**Errors of the first request:**

| Status | Meaning |
|--------|---------|
| 400 | Invalid body, empty `items`, or an invalid name |
| 401 / 403 | No valid token, or no `DOWNLOAD` intent |
| 404 | An item does not exist, or it is outside the data folder |
| 413 | Over a limit. Body: `{"error": "too_many_entries", "limit": 20000}` or `{"error": "too_large", "limit": 53687091200}`. See `ZIP_DOWNLOAD_MAX_FILES` and `ZIP_DOWNLOAD_MAX_SIZE` |
| 429 | The user already runs 2 zip downloads (`Retry-After: 30`) |
| 503 | The server already runs 3 zip downloads (`Retry-After: 30`) |

### Read Text File
```http
GET /getTextFile?token=jwt_token&loc=/path/to/file&name=filename
```

**Response:**
```json
{
  "content": "file content as string",
  "encoding": "utf8"
}
```

### Save Text File
```http
POST /saveTextFile?token=jwt_token
Content-Type: application/json

{
  "loc": "/path/to/file",
  "name": "filename.txt",
  "content": "file content"
}
```

**Response:**
```json
{
  "success": true,
  "message": "File saved successfully"
}
```

### Delete File/Directory
```http
GET /forceDelete?token=jwt_token&loc=/path&name=item_name
```

**Required Intent:** `DELETE`

**Response:**
```json
{
  "success": true,
  "message": "Item deleted successfully"
}
```

### Copy File/Directory
```http
GET /copy?token=jwt_token&loc=/source/path&name=item_name&dest=/destination/path
```

**Parameters:**
- `loc`: Source directory path
- `name`: Item name to copy
- `dest`: Destination directory path

**Required Intent:** `COPY`

**Response:**
```json
{
  "success": true,
  "message": "Item copied successfully"
}
```

### Move File/Directory
```http
GET /move?token=jwt_token&loc=/source/path&name=item_name&dest=/destination/path
```

**Required Intent:** `COPY`

### Rename File/Directory
```http
GET /rename?token=jwt_token&loc=/path&name=old_name&new_name=new_name
```

**Required Intent:** `RENAME`

### Create Directory
```http
GET /makedir?token=jwt_token&loc=/parent/path&name=directory_name
```

**Required Intent:** `UPLOAD`

### List Directory Contents
```http
GET /readFolder?token=jwt_token&loc=/directory/path
```

**Response:**
```json
{
  "folders": [
    {
      "name": "subfolder",
      "isDirectory": true,
      "size": 0,
      "lastModified": "2025-08-28T10:00:00.000Z"
    }
  ],
  "files": [
    {
      "name": "document.pdf",
      "isDirectory": false,
      "size": 2048000,
      "formattedSize": "2.0 MB",
      "lastModified": "2025-08-28T09:30:00.000Z",
      "extension": "pdf"
    }
  ]
}
```

### Search Files
```http
GET /searchInAllFiles?token=jwt_token&query=search_term&loc=/search/path
```

**Response:**
```json
{
  "results": [
    {
      "file": "/path/to/file.txt",
      "matches": 3,
      "preview": "...content preview..."
    }
  ]
}
```

---

## Media Operations

### Stream Video
```http
GET /getVideoData?token=jwt_token&loc=/path/to/video&name=video.mp4
```

**Features:**
- Range request support for seeking
- Proper MIME type detection
- Streaming optimized headers

**Response Headers:**
- `Content-Type`: video/mp4, video/webm, etc.
- `Accept-Ranges`: bytes
- `Content-Range`: bytes start-end/total (for range requests)

### Stream Audio
```http
GET /getAudioData?token=jwt_token&loc=/path/to/audio&name=audio.mp3
```

**Similar to video streaming with audio MIME types**

### Get Image
```http
GET /getImageData?token=jwt_token&loc=/path/to/image&name=image.jpg
```

**Also available via:**
```http
GET /img?token=jwt_token&loc=/path/to/image&name=image.jpg
```

---

## Archive Operations

### Create ZIP Archive
```http
POST /zipFiles?token=jwt_token
Content-Type: application/json

{
  "loc": "/source/directory",
  "files": ["file1.txt", "file2.pdf", "subfolder/"],
  "zipName": "archive.zip"
}
```

**Response:**
```json
{
  "success": true,
  "downloadId": "uuid-string",
  "message": "ZIP created successfully"
}
```

### Extract ZIP Archive
```http
POST /unzipFile?token=jwt_token
Content-Type: application/json

{
  "loc": "/destination/directory",
  "zipFile": "archive.zip"
}
```

### Download Created ZIP
```http
GET /downloadZip?token=jwt_token&id=download_id
```

### Delete Temporary ZIP
```http
GET /deleteTempZip?token=jwt_token&id=download_id
```

### Check Operation Progress
```http
GET /progress?token=jwt_token&id=operation_id
```

**Response:**
```json
{
  "progress": 75,
  "status": "processing",
  "message": "Extracting files..."
}
```

---

## File Upload

### Upload Files
```http
POST /input?token=jwt_token&loc=/destination/path
Content-Type: multipart/form-data

// Form data with file uploads
```

**Required Intent:** `UPLOAD`

**Features:**
- Multiple file upload support
- Large file support (up to configured limit)
- Progress tracking
- Automatic directory creation

### Upload ZIP File
```http
POST /inputZip?token=jwt_token&loc=/destination/path&extract=true
Content-Type: multipart/form-data

// ZIP file upload with optional extraction
```

---

## User Management

### Get User Permissions
```http
GET /getIntents?token=jwt_token&userId=target_user_id
```

**Response:**
```json
{
  "intents": ["VIEW", "DOWNLOAD", "UPLOAD"],
  "isAdmin": false
}
```

### Check Admin Status
```http
GET /checkAdmin?token=jwt_token&userId=user_id&adminPassword=admin_password
```

### Get All Users
```http
GET /getAllUsers?token=jwt_token
```

**Required Intent:** `ADMIN`

**Response:**
```json
{
  "users": [
    {
      "userId": "user1",
      "authType": "local",
      "intents": ["VIEW", "DOWNLOAD"],
      "isAdmin": false
    }
  ]
}
```

### Get Activity Logs
```http
GET /getActivityLog?token=jwt_token&limit=50&offset=0
```

**Required Intent:** `ADMIN`

**Response:**
```json
{
  "logs": [
    {
      "id": 1,
      "userId": "user1",
      "activity": "file_download",
      "description": "Downloaded file.pdf",
      "location": "/documents/",
      "timestamp": "2025-08-28T10:30:00.000Z"
    }
  ],
  "total": 150
}
```

### Check User Permission
```http
GET /checkIntent?token=jwt_token&intent=UPLOAD
```

**Response:**
```json
{
  "hasPermission": true,
  "intent": "UPLOAD"
}
```

### Grant Permission
```http
GET /authorize?token=jwt_token&userId=target_user&intent=UPLOAD
```

**Required Intent:** `ADMIN`

### Revoke Permission
```http
GET /unauthorize?token=jwt_token&userId=target_user&intent=UPLOAD
```

**Required Intent:** `ADMIN`

### Request Admin Permission
```http
POST /requestAdminIntent?token=jwt_token
Content-Type: application/json

{
  "pwd": "admin_password"
}
```

In `oidc` mode this route answers `403`. See [Routes Closed in This Mode](#routes-closed-in-this-mode).

---

## System Information

### Get System Info
```http
GET /getSystemInfo?token=jwt_token
```

**Response:**
```json
{
  "system": {
    "platform": "linux",
    "arch": "x64",
    "uptime": "7 days, 12 hours",
    "nodeVersion": "v20.5.0"
  },
  "memory": {
    "total": "16 GB",
    "used": "8.2 GB", 
    "free": "7.8 GB",
    "usage": 51.25
  },
  "cpu": {
    "model": "Intel Core i7-9700K",
    "cores": 8,
    "usage": 25.6
  },
  "disk": {
    "total": "1.0 TB",
    "used": "450 GB",
    "free": "550 GB", 
    "usage": 45.0
  },
  "network": {
    "hostname": "nas-server",
    "interfaces": ["eth0", "lo"]
  }
}
```

### Log Activity
```http
POST /log?token=jwt_token
Content-Type: application/json

{
  "activity": "custom_action",
  "description": "User performed custom action",
  "location": "/path/involved"
}
```

---

## Response Formats

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {}  // Optional data payload
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error message",
  "code": "ERROR_CODE"  // Optional error code
}
```

### File List Response
```json
{
  "folders": [...],
  "files": [...],
  "path": "/current/path",
  "breadcrumb": [
    {"name": "Home", "path": "/"},
    {"name": "Documents", "path": "/documents"}
  ]
}
```

---

## Error Handling

### HTTP Status Codes

- **200 OK**: Successful operation
- **201 Created**: Resource created successfully  
- **206 Partial Content**: Range request fulfilled
- **400 Bad Request**: Invalid request parameters
- **401 Unauthorized**: Missing or invalid JWT token
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: File or resource not found
- **409 Conflict**: Resource already exists
- **413 Payload Too Large**: File size exceeds limit
- **429 Too Many Requests**: Rate limit exceeded
- **500 Internal Server Error**: Server error

### Common Error Codes

```typescript
enum ErrorCodes {
  INVALID_TOKEN = "INVALID_TOKEN",
  INSUFFICIENT_PERMISSIONS = "INSUFFICIENT_PERMISSIONS", 
  FILE_NOT_FOUND = "FILE_NOT_FOUND",
  DIRECTORY_NOT_FOUND = "DIRECTORY_NOT_FOUND",
  FILE_ALREADY_EXISTS = "FILE_ALREADY_EXISTS",
  INVALID_FILE_TYPE = "INVALID_FILE_TYPE",
  FILE_TOO_LARGE = "FILE_TOO_LARGE",
  DISK_SPACE_FULL = "DISK_SPACE_FULL",
  OPERATION_FAILED = "OPERATION_FAILED"
}
```

### Permission Requirements

| Intent | Description | Required For |
|--------|-------------|--------------|
| `ADMIN` | Administrative access | User management, system settings |
| `VIEW` | View files and directories | File listing, navigation |
| `OPEN` | Open and read files | File viewing, text editing |
| `DOWNLOAD` | Download files | File downloads |
| `UPLOAD` | Upload files | File uploads, directory creation |
| `COPY` | Copy/move files | Copy, move operations |
| `DELETE` | Delete files | File/directory deletion |
| `RENAME` | Rename files | Rename operations |

---

## Rate Limiting

**Default Limits:**
- File operations: 100 requests per minute
- Upload operations: 10 requests per minute  
- Authentication: 5 attempts per minute per IP

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

---

## WebSocket Support

Currently not implemented. All operations use HTTP REST API with polling for progress updates.

---

## API Versioning

Current API version: **v1** (implicit)
Base URL: `http://localhost:7777/`

Future versions will use explicit versioning:
- `http://localhost:7777/v2/endpoint`

---

## SDK and Client Libraries

**JavaScript/TypeScript Client Example:**
```typescript
class NasApiClient {
  constructor(private baseUrl: string, private token: string) {}
  
  async uploadFile(file: File, path: string): Promise<ApiResponse> {
    const formData = new FormData();
    formData.append('file', file);
    
    const response = await fetch(`${this.baseUrl}/input?token=${this.token}&loc=${path}`, {
      method: 'POST',
      body: formData
    });
    
    return response.json();
  }
  
  async listDirectory(path: string): Promise<DirectoryListing> {
    const response = await fetch(`${this.baseUrl}/readFolder?token=${this.token}&loc=${path}`);
    return response.json();
  }
}
```

---

*This API reference covers all available endpoints. For implementation examples, see the [Development Guide](development-guide.md).*