// Import environment configuration
import { ENVIRONMENT, Environment } from './environment';

// Re-export environment variables for backward compatibility
export const private_key = ENVIRONMENT.PRIVATE_KEY;
export const admin_passowrd = ENVIRONMENT.ADMIN_PASSWORD; // Note: keeping typo for compatibility
export const PORT = ENVIRONMENT.PORT;
export const diskPath = ENVIRONMENT.DISK_PATH;

// Authentication type: 'oauth' | 'local' | 'both'
export const AUTH_TYPE = ENVIRONMENT.AUTH_TYPE;

// OAuth Configuration - Enhanced with Discord support
export const DISCORD_CLIENT_ID = Environment.DISCORD_CLIENT_ID;
export const DISCORD_CLIENT_SECRET = Environment.DISCORD_CLIENT_SECRET;
export const DISCORD_REDIRECT_URI = Environment.DISCORD_REDIRECT_URI;
export const DISCORD_LOGIN_URL = Environment.DISCORD_LOGIN_URL;

export const KAKAO_REST_API_KEY = ENVIRONMENT.KAKAO_REST_API_KEY;
export const KAKAO_REDIRECT_URL = ENVIRONMENT.KAKAO_REDIRECT_URL; // Backward compatibility: URL vs URI
export const KAKAO_REDIRECT_URI = Environment.KAKAO_REDIRECT_URI; // New consistent naming
export const KAKAO_CLIENT_SECRET = ENVIRONMENT.KAKAO_CLIENT_SECRET;
export const KAKAO_LOGIN_URL = Environment.KAKAO_LOGIN_URL;

// Password complexity requirements for local auth
export const PASSWORD_MIN_LENGTH = ENVIRONMENT.PASSWORD_MIN_LENGTH;
export const PASSWORD_REQUIRE_UPPERCASE = ENVIRONMENT.PASSWORD_REQUIRE_UPPERCASE;
export const PASSWORD_REQUIRE_LOWERCASE = ENVIRONMENT.PASSWORD_REQUIRE_LOWERCASE;
export const PASSWORD_REQUIRE_NUMBER = ENVIRONMENT.PASSWORD_REQUIRE_NUMBER;
export const PASSWORD_REQUIRE_SPECIAL = ENVIRONMENT.PASSWORD_REQUIRE_SPECIAL;

// Modern password requirements object (recommended for new code)
export const PASSWORD_REQUIREMENTS = Environment.PASSWORD_REQUIREMENTS;

// Additional configuration exports for enhanced functionality
export const JWT_EXPIRY = Environment.JWT_EXPIRY;
export const DEBUG_MODE = Environment.DEBUG_MODE;
export const ENABLE_CORS = Environment.ENABLE_CORS;
export const ENABLE_REQUEST_LOGGING = Environment.ENABLE_REQUEST_LOGGING;
export const MAX_FILE_SIZE = Environment.MAX_FILE_SIZE;
export const ENABLE_STREAMING = Environment.ENABLE_STREAMING;

// Export the modern Environment class for new code
export { Environment } from './environment';