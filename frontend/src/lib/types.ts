// Shared TypeScript types for the NAS frontend.
// Stores (lib/store/*) and components both import from here so the shapes stay in lockstep.

export type Intent =
  | "ADMIN"
  | "VIEW"
  | "OPEN"
  | "DOWNLOAD"
  | "UPLOAD"
  | "COPY"
  | "DELETE"
  | "RENAME";

export interface User {
  userId: string;
  username: string;
  krname: string;
  global_name: string;
  intents?: Intent[];
}

export interface AuthUser extends User {
  token: string;
}

export interface FileEntry {
  name: string;
  loc: string;
  extensions: string;
  modified: boolean;
}

export type TabKind =
  | "explorer"
  | "text"
  | "image"
  | "video"
  | "audio"
  | "pdf"
  | "user-manager"
  | "settings"
  | "activity"
  | "account"
  | "system";

export interface Tab {
  id: string;
  kind: TabKind;
  title: string;
  icon: string; // lucide icon name
  payload: unknown; // tab-specific (e.g. {loc, name} for files)
  dirty?: boolean;
  closable: boolean;
}

export type UploadStatus =
  | "queued"
  | "uploading"
  | "paused"
  | "complete"
  | "error"
  | "cancelled";

export interface Upload {
  id: string;
  filename: string;
  loc: string;
  totalBytes: number;
  uploadedBytes: number;
  status: UploadStatus;
  startedAt: number;
  completedAt?: number;
  errorMessage?: string;
  tusUrl?: string;
  file: File | Blob;
}

export type ViewMode = "grid" | "list";
export type SortKey = "name" | "modified" | "size" | "type";
export type SortDir = "asc" | "desc";

export type NotificationKind = "info" | "success" | "warning" | "error";

export interface Notification {
  id: string;
  kind: NotificationKind;
  message: string;
  durationMs: number;
  createdAt: number;
}

export type Breakpoint = "sm" | "md" | "lg" | "xl";
