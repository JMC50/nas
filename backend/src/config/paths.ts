import path from "path";
import os from "os";

export interface PathConfig {
  dataDir: string;
  adminDataDir: string;
  dbDir: string;
  tempDir: string;
}

/**
 * Get platform-specific path configuration
 */
function getPlatformPaths(): PathConfig {
  const isDev = process.env.NODE_ENV !== 'production';
  const platform = os.platform();
  
  if (isDev) {
    // Development mode - relative paths from project root
    return {
      dataDir: path.resolve(process.cwd(), '../../nas-data'),
      adminDataDir: path.resolve(process.cwd(), '../../nas-data-admin'),
      dbDir: path.resolve(__dirname, '..', 'db'),
      tempDir: os.tmpdir()
    };
  } else {
    // Production mode
    if (platform === 'win32') {
      // Windows production
      const appData = process.env.APPDATA || path.join(os.homedir(), 'AppData', 'Roaming');
      const nasDir = path.join(appData, 'NAS');
      
      return {
        dataDir: path.join(nasDir, 'data'),
        adminDataDir: path.join(nasDir, 'admin-data'),
        dbDir: path.join(nasDir, 'db'),
        tempDir: os.tmpdir()
      };
    } else {
      // Linux/Unix production
      const homeDir = process.env.HOME || os.homedir();
      const nasDir = process.env.NAS_DATA_DIR || path.join(homeDir, 'nas-storage');
      
      return {
        dataDir: path.join(nasDir, 'data'),
        adminDataDir: path.join(nasDir, 'admin-data'),
        dbDir: path.join(nasDir, 'db'),
        tempDir: process.env.NAS_TEMP_DIR || '/tmp'
      };
    }
  }
}

export const PATHS = getPlatformPaths();

/**
 * Get file path within data directory
 */
export function getDataPath(...segments: string[]): string {
  return path.join(PATHS.dataDir, ...segments);
}

/**
 * Get file path within admin data directory
 */
export function getAdminDataPath(...segments: string[]): string {
  return path.join(PATHS.adminDataDir, ...segments);
}

/**
 * Get database path
 */
export function getDbPath(filename: string = 'nas.sqlite'): string {
  return path.join(PATHS.dbDir, filename);
}

/**
 * Get temporary file path
 */
export function getTempPath(filename: string): string {
  return path.join(PATHS.tempDir, `nas-${filename}`);
}

/**
 * Convert location string to safe file path
 */
export function sanitizePath(loc: string): string {
  return loc.replace(/^\/+/, '').replace(/\/+/g, path.sep);
}