import { notifications } from "$lib/store/notifications.svelte";
import { moveEntry } from "./actions";

export const NAS_ENTRY_MIME = "application/x-nas-entry";

export interface DragItem {
  name: string;
  isFolder: boolean;
}

export interface DragPayload {
  items: DragItem[];
  sourceLoc: string[];
}

// Legacy shape from feat/ux-overhaul: { name, isFolder, sourceLoc }. Tolerated
// by readPayload for one minor version so in-flight drags during a hot reload
// or live deploy don't drop.
interface LegacyDragPayload {
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

/**
 * Build a drag payload. Accepts either a single entry (back-compat with the
 * one-entry call sites) or an array of entries for multi-select drags.
 * Result is always the batch shape `{ items, sourceLoc }`.
 */
export function buildPayload(
  loc: string[],
  entries: DragItem | DragItem[],
): string {
  const items = Array.isArray(entries) ? entries : [entries];
  return JSON.stringify({ items, sourceLoc: loc } satisfies DragPayload);
}

export function readPayload(event: DragEvent): DragPayload | null {
  const raw = event.dataTransfer?.getData(NAS_ENTRY_MIME);
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw) as DragPayload | LegacyDragPayload;
    // Batch shape — return as-is.
    if (Array.isArray((parsed as DragPayload).items)) {
      return parsed as DragPayload;
    }
    // Legacy single-entry shape — wrap into batch.
    const legacy = parsed as LegacyDragPayload;
    if (typeof legacy.name === "string") {
      return {
        items: [{ name: legacy.name, isFolder: legacy.isFolder }],
        sourceLoc: legacy.sourceLoc,
      };
    }
    return null;
  } catch {
    return null;
  }
}

export function hasPayload(event: DragEvent): boolean {
  const types = event.dataTransfer?.types ?? [];
  return Array.from(types).includes(NAS_ENTRY_MIME);
}

export async function performMove(
  sourceLoc: string[],
  name: string,
  isFolder: boolean,
  targetLoc: string[],
  opts: { silent?: boolean } = {},
): Promise<boolean> {
  if (locsEqual(sourceLoc, targetLoc)) return false;
  if (isFolder && isDescendant([...sourceLoc, name], targetLoc)) {
    notifications.error("Cannot move a folder into itself or its descendant");
    return false;
  }
  try {
    await moveEntry(sourceLoc, targetLoc, name, { silent: opts.silent });
    return true;
  } catch (cause) {
    notifications.error(`Move failed: ${(cause as Error).message}`);
    return false;
  }
}

/**
 * Move multiple entries to `targetLoc` sequentially. Returns counts so the
 * caller can decide when to `refresh()`. Emits a single summary notification
 * instead of one-per-item to avoid toast spam.
 */
export async function performMoveBatch(
  sourceLoc: string[],
  items: DragItem[],
  targetLoc: string[],
): Promise<{ moved: number; failed: number }> {
  let moved = 0;
  let failed = 0;
  for (const item of items) {
    const ok = await performMove(sourceLoc, item.name, item.isFolder, targetLoc, {
      silent: true,
    });
    if (ok) moved++;
    else failed++;
  }
  if (moved > 0) {
    const suffix = failed > 0 ? ` (${failed} failed)` : "";
    notifications.info(`Moved ${moved} item${moved === 1 ? "" : "s"}${suffix}`);
  }
  return { moved, failed };
}
