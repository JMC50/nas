import { config } from "dotenv";
import { PATHS } from "./paths";

// Load environment variables
config();

export const PORT = parseInt(process.env.PORT || "7777");
export const HOST = process.env.HOST || "0.0.0.0";
export const NODE_ENV = process.env.NODE_ENV || "development";

export const private_key = process.env.PRIVATE_KEY || "development-key";
export const admin_passowrd = process.env.ADMIN_PASSWORD || "admin123";
export const AUTH_TYPE =
  (process.env.AUTH_TYPE as "oauth" | "local" | "both") || "local";
export const JWT_EXPIRY = process.env.JWT_EXPIRY || "24h";

export const CORS_ORIGIN = process.env.CORS_ORIGIN || "*";
export const ENABLE_CORS = process.env.ENABLE_CORS !== "false";

export const MAX_FILE_SIZE = process.env.MAX_FILE_SIZE || "50gb";
export const ALLOWED_EXTENSIONS = process.env.ALLOWED_EXTENSIONS || "*";

export const PASSWORD_MIN_LENGTH = parseInt(
  process.env.PASSWORD_MIN_LENGTH || "8"
);
export const PASSWORD_REQUIRE_UPPERCASE =
  process.env.PASSWORD_REQUIRE_UPPERCASE === "true";
export const PASSWORD_REQUIRE_LOWERCASE =
  process.env.PASSWORD_REQUIRE_LOWERCASE === "true";
export const PASSWORD_REQUIRE_NUMBER =
  process.env.PASSWORD_REQUIRE_NUMBER === "true";
export const PASSWORD_REQUIRE_SPECIAL =
  process.env.PASSWORD_REQUIRE_SPECIAL === "true";

export const DEBUG_MODE = process.env.DEBUG_MODE === "true";
export const LOG_LEVEL = process.env.LOG_LEVEL || "info";
export const ENABLE_REQUEST_LOGGING =
  process.env.ENABLE_REQUEST_LOGGING === "true";

export const diskPath = PATHS.diskPath;

export const DISCORD_CLIENT_ID = process.env.DISCORD_CLIENT_ID || "";
export const DISCORD_CLIENT_SECRET = process.env.DISCORD_CLIENT_SECRET || "";
export const DISCORD_REDIRECT_URI = process.env.DISCORD_REDIRECT_URI || "";
export const DISCORD_LOGIN_URL = process.env.DISCORD_LOGIN_URL || "";

export const KAKAO_REST_API_KEY = process.env.KAKAO_REST_API_KEY || "";
export const KAKAO_REDIRECT_URL = process.env.KAKAO_REDIRECT_URI || "";
export const KAKAO_REDIRECT_URI = process.env.KAKAO_REDIRECT_URI || "";
export const KAKAO_CLIENT_SECRET = process.env.KAKAO_CLIENT_SECRET || "";
export const KAKAO_LOGIN_URL = process.env.KAKAO_LOGIN_URL || "";

if (NODE_ENV === "production") {
  if (!private_key || private_key === "development-key") {
    throw new Error("Production requires secure PRIVATE_KEY");
  }
  if (!admin_passowrd || admin_passowrd === "admin123") {
    throw new Error("Production requires secure ADMIN_PASSWORD");
  }
}

if (DEBUG_MODE) {
  console.log(`   Environment: ${NODE_ENV}`);
  console.log(`   Server: ${HOST}:${PORT}`);
  console.log(`   Auth Type: ${AUTH_TYPE}`);
  console.log(`   Data Path: ${PATHS.dataDir}`);
}
