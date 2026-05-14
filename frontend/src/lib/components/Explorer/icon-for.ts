import FolderIcon from "lucide-svelte/icons/folder";
import FileIcon from "lucide-svelte/icons/file";
import FileText from "lucide-svelte/icons/file-text";
import ImageIcon from "lucide-svelte/icons/image";
import Film from "lucide-svelte/icons/film";
import Music from "lucide-svelte/icons/music";
import FileArchive from "lucide-svelte/icons/file-archive";
import FileType from "lucide-svelte/icons/file-type";

const IMAGE_EXT = new Set(["jpg", "jpeg", "png", "gif", "webp", "avif", "bmp", "svg"]);
const VIDEO_EXT = new Set(["mp4", "webm", "mov", "mkv", "avi"]);
const AUDIO_EXT = new Set(["mp3", "wav", "ogg", "flac", "m4a"]);
const ARCHIVE_EXT = new Set(["zip", "tar", "gz", "rar", "7z"]);
const TEXT_EXT = new Set([
  "md", "txt", "json", "yaml", "yml", "js", "ts",
  "go", "py", "rs", "html", "css", "svelte",
]);

export interface FolderEntry {
  name: string;
  isFolder: boolean;
  extensions: string;
  size: number;
  modifiedAt: string;
}

export function iconFor(entry: FolderEntry) {
  if (entry.isFolder) return FolderIcon;
  const ext = entry.extensions.toLowerCase();
  if (IMAGE_EXT.has(ext)) return ImageIcon;
  if (VIDEO_EXT.has(ext)) return Film;
  if (AUDIO_EXT.has(ext)) return Music;
  if (ARCHIVE_EXT.has(ext)) return FileArchive;
  if (ext === "pdf") return FileType;
  if (TEXT_EXT.has(ext)) return FileText;
  return FileIcon;
}
