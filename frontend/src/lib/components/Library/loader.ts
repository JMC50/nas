// Loader for the Music + Video library tabs. Calls the backend
// `/server/mediaLibrary?kind=audio|video` endpoint and returns the rows along
// with a truncated flag derived from the `X-Library-Truncated` response header
// (set when the walk hit `MEDIA_LIB_LIMIT`).

import { auth } from "$lib/store/auth.svelte";
import type { MediaEntry } from "$lib/types";

export interface LibraryResult {
  entries: MediaEntry[];
  truncated: boolean;
}

export async function loadLibrary(
  kind: "audio" | "video",
): Promise<LibraryResult> {
  const token = auth.token;
  const url = `/server/mediaLibrary?kind=${kind}&token=${encodeURIComponent(token)}`;
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`mediaLibrary ${kind} returned ${response.status}`);
  }
  const entries = (await response.json()) as MediaEntry[];
  return {
    entries: Array.isArray(entries) ? entries : [],
    truncated: response.headers.get("X-Library-Truncated") === "true",
  };
}
