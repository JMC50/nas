import { auth } from "$lib/store/auth.svelte";
import { notifications } from "$lib/store/notifications.svelte";
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

export async function deleteEntry(loc: string[], entry: FolderEntry): Promise<void> {
  const response = await fetch(url("forceDelete", { name: entry.name, loc: locPath(loc) }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Deleted ${entry.name}`);
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
): Promise<void> {
  const response = await fetch(
    url("move", {
      originLoc: locPath(originLoc),
      fileName: name,
      targetLoc: locPath(targetLoc),
    }),
  );
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Moved ${name}`);
}

export function downloadUrl(loc: string[], entry: FolderEntry): string {
  return url("download", { loc: locPath(loc), name: entry.name });
}
