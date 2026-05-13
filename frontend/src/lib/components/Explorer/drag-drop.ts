import { notifications } from "$lib/store/notifications.svelte";
import { moveEntry } from "./actions";

export const NAS_ENTRY_MIME = "application/x-nas-entry";

export interface DragPayload {
  name: string;
  isFolder: boolean;
  sourceLoc: string[];
}

export function locsEqual(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((segment, index) => segment === b[index]);
}

export function isDescendant(parent: string[], child: string[]): boolean {
  if (child.length < parent.length) return false;
  return parent.every((segment, index) => segment === child[index]);
}

export function buildPayload(
  loc: string[],
  entry: { name: string; isFolder: boolean },
): string {
  return JSON.stringify({ name: entry.name, isFolder: entry.isFolder, sourceLoc: loc });
}

export function readPayload(event: DragEvent): DragPayload | null {
  const raw = event.dataTransfer?.getData(NAS_ENTRY_MIME);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as DragPayload;
  } catch {
    return null;
  }
}

export function hasPayload(event: DragEvent): boolean {
  const types = event.dataTransfer?.types ?? [];
  return Array.from(types).includes(NAS_ENTRY_MIME);
}

export async function performMove(
  srcLoc: string[],
  name: string,
  isFolder: boolean,
  targetLoc: string[],
): Promise<boolean> {
  if (locsEqual(srcLoc, targetLoc)) return false;
  if (isFolder && isDescendant([...srcLoc, name], targetLoc)) {
    notifications.error("Cannot move a folder into itself or its descendant");
    return false;
  }
  try {
    await moveEntry(srcLoc, targetLoc, name);
    return true;
  } catch (cause) {
    notifications.error(`Move failed: ${(cause as Error).message}`);
    return false;
  }
}
