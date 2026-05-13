// frontend/src/lib/components/Viewers/media-utils.ts

export type MediaKey =
  | "toggle"
  | "mute"
  | "fullscreen"
  | "seekBack"
  | "seekForward"
  | "volumeUp"
  | "volumeDown";

const SECONDS_PER_MINUTE = 60;
const SECONDS_PER_HOUR = 3600;

export function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) return "0:00";
  const total = Math.floor(seconds);
  const hours = Math.floor(total / SECONDS_PER_HOUR);
  const minutes = Math.floor((total % SECONDS_PER_HOUR) / SECONDS_PER_MINUTE);
  const secs = total % SECONDS_PER_MINUTE;
  const pad = (num: number) => num.toString().padStart(2, "0");
  return hours > 0 ? `${hours}:${pad(minutes)}:${pad(secs)}` : `${minutes}:${pad(secs)}`;
}

export function clampVolume(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.min(1, Math.max(0, value));
}

const KEY_MAP: Record<string, MediaKey> = {
  Space: "toggle",
  KeyK: "toggle",
  KeyM: "mute",
  KeyF: "fullscreen",
  ArrowLeft: "seekBack",
  ArrowRight: "seekForward",
  ArrowUp: "volumeUp",
  ArrowDown: "volumeDown",
};

export function pickMediaKey(event: KeyboardEvent): MediaKey | null {
  return KEY_MAP[event.code] ?? null;
}
