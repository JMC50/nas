import { auth } from "$lib/store/auth.svelte";
import { files } from "$lib/store/files.svelte";
import { tabs } from "$lib/store/tabs.svelte";
import { notifications } from "$lib/store/notifications.svelte";
import { pickViewer } from "$lib/components/Viewers/registry";
import type { FolderEntry } from "./icon-for";

export class FetchError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export function locPath(): string {
  return "/" + files.currentLoc.join("/");
}

function url(path: string, params: Record<string, string>): string {
  const search = new URLSearchParams({ ...params, token: auth.token });
  return `/server/${path}?${search.toString()}`;
}

export async function readEntries(): Promise<FolderEntry[]> {
  const response = await fetch(url("readFolder", { loc: locPath() }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  const data = await response.json();
  return (data.files ?? data ?? []) as FolderEntry[];
}

export async function createFolder(name: string): Promise<void> {
  const response = await fetch(url("makedir", { name, loc: locPath() }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Created ${name}`);
}

export async function deleteEntry(entry: FolderEntry): Promise<void> {
  const response = await fetch(url("forceDelete", { name: entry.name, loc: locPath() }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Deleted ${entry.name}`);
}

export async function renameEntry(entry: FolderEntry, next: string): Promise<void> {
  const response = await fetch(url("rename", { loc: locPath(), name: entry.name, newName: next }));
  if (!response.ok) throw new FetchError(response.status, `HTTP ${response.status}`);
  notifications.success(`Renamed to ${next}`);
}

export function downloadUrl(entry: FolderEntry): string {
  return url("download", { loc: locPath(), name: entry.name });
}

export function openEntry(entry: FolderEntry) {
  if (entry.isFolder) {
    files.navigateInto(entry.name);
    return;
  }
  const kind = pickViewer(entry.extensions);
  tabs.open({
    kind,
    title: entry.name,
    icon: kind,
    payload: { loc: locPath(), name: entry.name },
    closable: true,
  });
}
