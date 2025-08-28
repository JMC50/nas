import path from "path";
import { config } from "dotenv";

config();

const DATA_PATH = process.env.DATA_PATH || "/mnt/nas-storage";

export interface PathConfig {
  dataDir: string;
  adminDataDir: string;
  dbDir: string;
  tempDir: string;
  diskPath: string;
}

export const PATHS: PathConfig = {
  dataDir: path.join(DATA_PATH, "files"),

  adminDataDir: path.join(DATA_PATH, "admin"),

  dbDir: path.join(DATA_PATH, "database"),

  tempDir: path.join(DATA_PATH, "temp"),

  diskPath: DATA_PATH,
};

export function getDataPath(...segments: string[]): string {
  return path.join(PATHS.dataDir, ...segments);
}

export function getAdminDataPath(...segments: string[]): string {
  return path.join(PATHS.adminDataDir, ...segments);
}

export function getDbPath(filename: string = "nas.sqlite"): string {
  return path.join(PATHS.dbDir, filename);
}

export function getTempPath(filename: string): string {
  return path.join(PATHS.tempDir, `nas-${filename}`);
}

export function sanitizePath(loc: string): string {
  return loc.replace(/^\/+/, "").replace(/\/+/g, path.sep);
}

if (process.env.DEBUG_MODE === "true") {
  console.log(`   DATA_PATH: ${DATA_PATH}`);
  console.log(`   Data Dir: ${PATHS.dataDir}`);
  console.log(`   Admin Dir: ${PATHS.adminDataDir}`);
  console.log(`   DB Dir: ${PATHS.dbDir}`);
  console.log(`   Temp Dir: ${PATHS.tempDir}`);
  console.log(`   Disk Path: ${PATHS.diskPath}`);
}
