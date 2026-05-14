// Shared formatters for Explorer file metadata.
// `formatBytes` mirrors backend `formatSize` in
// `backend/internal/files/handlers.go` (2-decimal precision above KB) so
// inline columns and any future Stat-derived strings stay consistent.

const KB = 1024;
const MB = 1024 * KB;
const GB = 1024 * MB;

export function formatBytes(n: number): string {
  if (n < KB) return `${n} B`;
  if (n < MB) return `${(n / KB).toFixed(2)} KB`;
  if (n < GB) return `${(n / MB).toFixed(2)} MB`;
  return `${(n / GB).toFixed(2)} GB`;
}

const MIN = 60;
const HOUR = 60 * MIN;
const DAY = 24 * HOUR;
const WEEK = 7 * DAY;

export function formatRelTime(rfc3339: string): string {
  const then = new Date(rfc3339).getTime();
  if (Number.isNaN(then)) return rfc3339;
  const deltaSec = Math.floor((Date.now() - then) / 1000);
  if (deltaSec < MIN) return "just now";
  if (deltaSec < HOUR) return `${Math.floor(deltaSec / MIN)}m`;
  if (deltaSec < DAY) return `${Math.floor(deltaSec / HOUR)}h`;
  if (deltaSec < WEEK) return `${Math.floor(deltaSec / DAY)}d`;
  return new Date(then).toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

export function formatFullTime(rfc3339: string): string {
  return new Date(rfc3339).toLocaleString();
}
