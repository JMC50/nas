import { auth } from "$lib/store/auth.svelte";
import { notifications } from "$lib/store/notifications.svelte";
import { formatBytes } from "./format";
import type { FolderEntry } from "./icon-for";

export class FetchError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export function locPath(loc: string[]): string {
  return "/" + loc.join("/");
}

function url(path: string, params: Record<string, string>): string {
  const search = new URLSearchParams({ ...params, token: auth.token });
  return `/server/${path}?${search.toString()}`;
}

export async function readEntries(loc: string[]): Promise<FolderEntry[]> {
  const response = await fetch(url("readFolder", { loc: locPath(loc) }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  const data = await response.json();
  return (data.files ?? data ?? []) as FolderEntry[];
}

export async function createFolder(loc: string[], name: string): Promise<void> {
  const response = await fetch(url("makedir", { name, loc: locPath(loc) }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Created ${name}`);
}

export async function deleteEntry(
  loc: string[],
  entry: FolderEntry,
  opts: { silent?: boolean } = {},
): Promise<void> {
  const response = await fetch(url("forceDelete", { name: entry.name, loc: locPath(loc) }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  if (!opts.silent) notifications.success(`Deleted ${entry.name}`);
}

export async function renameEntry(loc: string[], entry: FolderEntry, next: string): Promise<void> {
  const response = await fetch(url("rename", { loc: locPath(loc), name: entry.name, change: next }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Renamed to ${next}`);
}

export async function moveEntry(
  originLoc: string[],
  targetLoc: string[],
  name: string,
  opts: { silent?: boolean } = {},
): Promise<void> {
  const response = await fetch(
    url("move", {
      originLoc: locPath(originLoc),
      fileName: name,
      targetLoc: locPath(targetLoc),
    }),
  );
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  if (!opts.silent) notifications.success(`Moved ${name}`);
}

export function downloadUrl(loc: string[], entry: FolderEntry): string {
  return url("download", { loc: locPath(loc), name: entry.name });
}

export interface PreparedZip {
  url: string;
  name: string;
}

function zipRefusalText(status: number, body: { error?: string; limit?: number }): string {
  if (status === 403) return "You don't have permission to download files.";
  if (status === 404) return "Some selected items no longer exist. Refresh and try again.";
  if (status === 429) return "Your other zip downloads are still running. Wait for one to finish.";
  if (status === 503) return "The server is busy with other zip downloads. Try again in a moment.";
  if (body.error === "too_many_entries") {
    return `This selection has more than ${body.limit} files and folders. Select fewer items.`;
  }
  if (body.error === "too_large") {
    return `This selection is larger than ${formatBytes(body.limit ?? 0)}. Select fewer items.`;
  }
  return `Download failed (HTTP ${status}).`;
}

// 서버가 경로와 한도를 먼저 검사하고, 브라우저가 일반 다운로드로 받을 짧은 수명의 주소를 돌려줌
export async function prepareZip(loc: string[], entries: FolderEntry[]): Promise<PreparedZip> {
  const response = await fetch("/server/zip-downloads", {
    method: "POST",
    headers: { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" },
    body: JSON.stringify({ items: entries.map((entry) => ({ loc: locPath(loc), name: entry.name })) }),
  });
  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    throw new FetchError(response.status, zipRefusalText(response.status, body));
  }
  const data = (await response.json()) as { path: string; name: string };
  const search = new URLSearchParams({ token: auth.token });
  return { url: `/server${data.path}?${search.toString()}`, name: data.name };
}
