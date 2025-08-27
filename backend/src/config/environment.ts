import os from "os";
import { config } from "dotenv";
import { join } from "path";
import { readFileSync } from "fs";

// Load environment variables from ROOT .env file (development only)
if (process.env.NODE_ENV === 'development') {
  const rootEnvPath = join(__dirname, '../../../.env');
  const rootResult = config({ path: rootEnvPath });
  if (!rootResult.error) {
    console.log('✅ Loaded configuration from root .env file (development)');
  }
} else {
  console.log('✅ Production mode: Using environment variables and Docker secrets');
}

// Production-first: Docker secrets with environment variable fallback
function getSecret(name: string, defaultValue: string = ''): string {
  // Production: Docker secrets first
  const fileEnvVar = `${name}_FILE`;
  if (process.env[fileEnvVar]) {
    try {
      const secretValue = readFileSync(process.env[fileEnvVar]!, 'utf8').trim();
      console.log(`🔐 Loaded ${name} from Docker secret`);
      return secretValue;
    } catch (error) {
      throw new Error(`Failed to read Docker secret file ${process.env[fileEnvVar]} for ${name}: ${error}`);
    }
  }
  
  // Development: Environment variable
  const envValue = process.env[name];
  if (envValue) {
    if (process.env.NODE_ENV === 'development') {
      console.log(`📝 Loaded ${name} from environment variable (development)`);
    }
    return envValue;
  }
  
  // Last resort: default value (development only)
  if (process.env.NODE_ENV === 'development' && defaultValue) {
    console.warn(`⚠️  Using default value for ${name} (development only)`);
    return defaultValue;
  }
  
  return '';
}

// Enhanced Environment Configuration
export class Environment {
  // Server Configuration
  static readonly NODE_ENV = process.env.NODE_ENV || 'development';
  static readonly PORT = parseInt(process.env.PORT || '7777');
  static readonly HOST = process.env.HOST || (process.env.NODE_ENV === 'production' ? '0.0.0.0' : 'localhost');
  
  // URLs
  static readonly SERVER_URL = process.env.SERVER_URL || `http://${Environment.HOST}:${Environment.PORT}`;
  static readonly FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:5050';
  static readonly API_BASE_URL = process.env.API_BASE_URL || Environment.SERVER_URL;
  
  // Database
  static readonly DB_PATH = process.env.DB_PATH || './db/nas.sqlite';
  static readonly DB_TYPE = process.env.DB_TYPE || 'sqlite';
  static readonly DB_ENABLE_WAL = process.env.DB_ENABLE_WAL === 'true';
  static readonly DB_ENABLE_FOREIGN_KEYS = process.env.DB_ENABLE_FOREIGN_KEYS !== 'false';
  
  // Authentication & Security (Production: Docker secrets, Development: .env)
  static readonly AUTH_TYPE = (process.env.AUTH_TYPE as 'oauth' | 'local' | 'both') || 'both';
  static readonly PRIVATE_KEY = getSecret('PRIVATE_KEY', process.env.NODE_ENV === 'development' ? 'development-secret-key' : '');
  static readonly JWT_EXPIRY = process.env.JWT_EXPIRY || '24h';
  static readonly ADMIN_PASSWORD = getSecret('ADMIN_PASSWORD', process.env.NODE_ENV === 'development' ? 'admin123' : '');

  // Discord OAuth (supports Docker secrets)
  static readonly DISCORD_CLIENT_ID = getSecret('DISCORD_CLIENT_ID');
  static readonly DISCORD_CLIENT_SECRET = getSecret('DISCORD_CLIENT_SECRET');
  static readonly DISCORD_REDIRECT_URI = process.env.DISCORD_REDIRECT_URI;
  static readonly DISCORD_LOGIN_URL = process.env.DISCORD_LOGIN_URL;
  
  // Kakao OAuth (supports Docker secrets)
  static readonly KAKAO_REST_API_KEY = getSecret('KAKAO_REST_API_KEY');
  static readonly KAKAO_CLIENT_SECRET = getSecret('KAKAO_CLIENT_SECRET');
  static readonly KAKAO_REDIRECT_URI = process.env.KAKAO_REDIRECT_URI;
  static readonly KAKAO_LOGIN_URL = process.env.KAKAO_LOGIN_URL;
  
  // Password Requirements
  static readonly PASSWORD_REQUIREMENTS = {
    minLength: parseInt(process.env.PASSWORD_MIN_LENGTH || '8'),
    requireUppercase: process.env.PASSWORD_REQUIRE_UPPERCASE === 'true',
    requireLowercase: process.env.PASSWORD_REQUIRE_LOWERCASE === 'true',
    requireNumber: process.env.PASSWORD_REQUIRE_NUMBER === 'true',
    requireSpecial: process.env.PASSWORD_REQUIRE_SPECIAL === 'true'
  };

  // Storage & File System
  static readonly NAS_DATA_DIR = process.env.NAS_DATA_DIR;
  static readonly NAS_ADMIN_DATA_DIR = process.env.NAS_ADMIN_DATA_DIR;
  static readonly NAS_TEMP_DIR = process.env.NAS_TEMP_DIR;
  static readonly MAX_FILE_SIZE = process.env.MAX_FILE_SIZE || '50gb';
  static readonly ALLOWED_EXTENSIONS = process.env.ALLOWED_EXTENSIONS || '*';
  static readonly ENABLE_STREAMING = process.env.ENABLE_STREAMING !== 'false';
  
  // System Information
  static readonly PLATFORM = os.platform();
  static readonly IS_WINDOWS = os.platform() === 'win32';
  static readonly IS_LINUX = os.platform() === 'linux';
  static readonly IS_PRODUCTION = process.env.NODE_ENV === 'production';
  static readonly IS_DEVELOPMENT = process.env.NODE_ENV !== 'production';
  
  // System Configuration
  static readonly DISK_PATH = process.env.DISK_PATH || (os.platform() === 'win32' ? 'C:' : '/');
  static readonly CORS_ORIGIN = process.env.CORS_ORIGIN || '*';
  static readonly SESSION_TIMEOUT = parseInt(process.env.SESSION_TIMEOUT || '604800000'); // 7 days

  // Development & Debugging
  static readonly DEBUG_MODE = process.env.DEBUG_MODE === 'true';
  static readonly LOG_LEVEL = process.env.LOG_LEVEL || 'info';
  static readonly ENABLE_CORS = process.env.ENABLE_CORS !== 'false';
  static readonly ENABLE_REQUEST_LOGGING = process.env.ENABLE_REQUEST_LOGGING === 'true';

  // Validation method
  static validate(): void {
    const errors: string[] = [];
    const warnings: string[] = [];

    // Production: Require secrets, Development: Allow defaults
    if (Environment.IS_PRODUCTION) {
      if (!Environment.PRIVATE_KEY) {
        errors.push('PRIVATE_KEY must be provided via Docker secret or environment variable in production');
      }
      if (!Environment.ADMIN_PASSWORD) {
        errors.push('ADMIN_PASSWORD must be provided via Docker secret or environment variable in production');
      }
      if (Environment.PRIVATE_KEY === 'development-secret-key') {
        errors.push('PRIVATE_KEY cannot use development default in production');
      }
      if (Environment.ADMIN_PASSWORD === 'admin123') {
        errors.push('ADMIN_PASSWORD cannot use development default in production');
      }
    } else {
      // Development warnings
      if (Environment.PRIVATE_KEY === 'development-secret-key') {
        warnings.push('Using development default for PRIVATE_KEY');
      }
      if (Environment.ADMIN_PASSWORD === 'admin123') {
        warnings.push('Using development default for ADMIN_PASSWORD');
      }
    }

    // OAuth validation
    if (Environment.AUTH_TYPE !== 'local') {
      if (!Environment.DISCORD_CLIENT_ID && !Environment.KAKAO_REST_API_KEY) {
        warnings.push('No OAuth providers configured - set DISCORD_CLIENT_ID or KAKAO_REST_API_KEY');
      }
    }

    // Display validation results
    if (warnings.length > 0) {
      console.warn('⚠️  Configuration warnings:');
      warnings.forEach(warning => console.warn(`   - ${warning}`));
    }

    if (errors.length > 0) {
      console.error('❌ Configuration errors:');
      errors.forEach(error => console.error(`   - ${error}`));
      throw new Error(`Configuration validation failed: ${errors.join(', ')}`);
    }

    if (warnings.length === 0 && errors.length === 0) {
      console.log('✅ Environment configuration is valid');
    }
  }
}

// Legacy ENVIRONMENT object for backward compatibility
export const ENVIRONMENT = {
  NODE_ENV: Environment.NODE_ENV,
  PORT: Environment.PORT,
  HOST: Environment.HOST,
  
  // Database
  DB_PATH: Environment.DB_PATH,
  
  // Authentication
  AUTH_TYPE: Environment.AUTH_TYPE,
  PRIVATE_KEY: Environment.PRIVATE_KEY,
  ADMIN_PASSWORD: Environment.ADMIN_PASSWORD,
  
  // OAuth
  KAKAO_REST_API_KEY: Environment.KAKAO_REST_API_KEY,
  KAKAO_REDIRECT_URL: Environment.KAKAO_REDIRECT_URI, // Note: URI vs URL naming difference
  KAKAO_CLIENT_SECRET: Environment.KAKAO_CLIENT_SECRET,
  
  // Password requirements (backward compatible structure)
  PASSWORD_MIN_LENGTH: Environment.PASSWORD_REQUIREMENTS.minLength,
  PASSWORD_REQUIRE_UPPERCASE: Environment.PASSWORD_REQUIREMENTS.requireUppercase,
  PASSWORD_REQUIRE_LOWERCASE: Environment.PASSWORD_REQUIREMENTS.requireLowercase,
  PASSWORD_REQUIRE_NUMBER: Environment.PASSWORD_REQUIREMENTS.requireNumber,
  PASSWORD_REQUIRE_SPECIAL: Environment.PASSWORD_REQUIREMENTS.requireSpecial,
  
  // Storage
  DATA_DIR: Environment.NAS_DATA_DIR,
  ADMIN_DATA_DIR: Environment.NAS_ADMIN_DATA_DIR,
  TEMP_DIR: Environment.NAS_TEMP_DIR,
  
  // System
  PLATFORM: Environment.PLATFORM,
  IS_WINDOWS: Environment.IS_WINDOWS,
  IS_LINUX: Environment.IS_LINUX,
  IS_PRODUCTION: Environment.IS_PRODUCTION,
  IS_DEVELOPMENT: Environment.IS_DEVELOPMENT,
  
  // Other settings
  DISK_PATH: Environment.DISK_PATH,
  CORS_ORIGIN: Environment.CORS_ORIGIN,
  MAX_FILE_SIZE: Environment.MAX_FILE_SIZE,
  SESSION_TIMEOUT: Environment.SESSION_TIMEOUT,
};

// Validate configuration on import
Environment.validate();