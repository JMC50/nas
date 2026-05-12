import type { TabKind } from "$lib/types";

const TEXT_EXTENSIONS = new Set([
  "md",
  "txt",
  "json",
  "yaml",
  "yml",
  "toml",
  "ini",
  "conf",
  "log",
  "csv",
  "tsv",
  "js",
  "jsx",
  "ts",
  "tsx",
  "mjs",
  "cjs",
  "py",
  "go",
  "rs",
  "java",
  "kt",
  "swift",
  "c",
  "cpp",
  "cc",
  "cs",
  "h",
  "hpp",
  "rb",
  "php",
  "sh",
  "bash",
  "zsh",
  "ps1",
  "sql",
  "html",
  "htm",
  "xml",
  "svg",
  "css",
  "scss",
  "sass",
  "less",
  "vue",
  "svelte",
  "gitignore",
  "dockerfile",
  "makefile",
]);

const IMAGE_EXTENSIONS = new Set(["jpg", "jpeg", "png", "gif", "webp", "avif", "bmp", "ico"]);
const VIDEO_EXTENSIONS = new Set(["mp4", "webm", "mov", "mkv", "avi"]);
const AUDIO_EXTENSIONS = new Set(["mp3", "wav", "ogg", "flac", "m4a", "aac"]);
const PDF_EXTENSIONS = new Set(["pdf"]);

export function pickViewer(extension: string): TabKind {
  const ext = extension.toLowerCase().replace(/^\./, "");
  if (TEXT_EXTENSIONS.has(ext)) return "text";
  if (IMAGE_EXTENSIONS.has(ext)) return "image";
  if (VIDEO_EXTENSIONS.has(ext)) return "video";
  if (AUDIO_EXTENSIONS.has(ext)) return "audio";
  if (PDF_EXTENSIONS.has(ext)) return "pdf";
  return "text";
}

export function viewerIconName(kind: TabKind): string {
  switch (kind) {
    case "text":
      return "file-text";
    case "image":
      return "image";
    case "video":
      return "film";
    case "audio":
      return "music";
    case "pdf":
      return "file-type";
    default:
      return "file";
  }
}
