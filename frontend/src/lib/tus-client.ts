// Minimal tus 1.0.0 resumable upload client.
// Backend exposes the tus endpoint at /server/files/* (router.go).
// See https://tus.io/protocols/resumable-upload for protocol details.

import { auth } from "$lib/store/auth.svelte";

const CHUNK_SIZE_BYTES = 5 * 1024 * 1024; // 5 MiB
const TUS_VERSION = "1.0.0";

// Retry delays for transient network failures (ms). Cumulative ~15.5s window
// before surfacing the error. Covers Wi-Fi blips, brief proxy hiccups, and
// short server restarts (watchtower rolling restart typically takes <10s).
const RETRY_DELAYS_MS = [500, 1000, 2000, 4000, 8000];

export interface TusUploadOptions {
  file: File | Blob;
  loc: string;
  filename: string;
  onProgress?: (uploadedBytes: number) => void;
  /** Fires once the tus CREATE response yields a Location URL. The caller
   * should persist this so a later resume can pick up via getOffset(). */
  onCreate?: (uploadUrl: string) => void;
}

export class TusError extends Error {
  constructor(
    message: string,
    public readonly status: number,
  ) {
    super(message);
    this.name = "TusError";
  }
}

/** True when an error is worth retrying. Network failures (TypeError from
 * fetch reject) and 5xx server responses are transient; 4xx are permanent
 * (auth, validation, etc.); AbortError is intentional cancellation. */
export function isTransient(error: unknown): boolean {
  if (error instanceof DOMException && error.name === "AbortError") return false;
  if (error instanceof TusError) {
    return error.status === 0 || error.status >= 500;
  }
  if (error instanceof TypeError) return true;
  return false;
}

/** Resolves after `ms`, or rejects immediately if `signal` aborts first.
 * Used between retries so a paused/cancelled upload stops backing off. */
function delay(ms: number, signal: AbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const onAbort = () => {
      clearTimeout(timer);
      signal.removeEventListener("abort", onAbort);
      reject(new DOMException("Upload aborted", "AbortError"));
    };
    const timer = setTimeout(() => {
      signal.removeEventListener("abort", onAbort);
      resolve();
    }, ms);
    signal.addEventListener("abort", onAbort);
  });
}

/** Run `operation` with retry-on-transient-error and exponential backoff.
 * AbortError + permanent (4xx) errors propagate immediately. */
async function withRetry<T>(
  operation: () => Promise<T>,
  signal: AbortSignal,
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt <= RETRY_DELAYS_MS.length; attempt++) {
    if (signal.aborted) {
      throw new DOMException("Upload aborted", "AbortError");
    }
    try {
      return await operation();
    } catch (error) {
      lastError = error;
      if (!isTransient(error) || attempt === RETRY_DELAYS_MS.length) {
        throw error;
      }
      await delay(RETRY_DELAYS_MS[attempt], signal);
    }
  }
  // Unreachable: the loop either returns or throws. Keep TS happy.
  throw lastError;
}

function encodeBase64(value: string): string {
  return btoa(unescape(encodeURIComponent(value)));
}

function metaHeader(loc: string, filename: string): string {
  return [
    `loc ${encodeBase64(loc)}`,
    `name ${encodeBase64(filename)}`,
    `filename ${encodeBase64(filename)}`,
  ].join(",");
}

function bearer(): string {
  return `Bearer ${auth.token}`;
}

async function createUpload(options: TusUploadOptions): Promise<string> {
  const response = await fetch("/server/files", {
    method: "POST",
    headers: {
      "Tus-Resumable": TUS_VERSION,
      "Upload-Length": String(options.file.size),
      "Upload-Metadata": metaHeader(options.loc, options.filename),
      "Content-Type": "application/offset+octet-stream",
      Authorization: bearer(),
    },
  });

  if (response.status !== 201) {
    throw new TusError(
      `tus CREATE failed: ${response.status} ${await response.text()}`,
      response.status,
    );
  }

  const location = response.headers.get("Location");
  if (!location) {
    throw new TusError("tus CREATE missing Location header", 500);
  }
  return location;
}

async function getOffset(uploadUrl: string, signal: AbortSignal): Promise<number> {
  return withRetry(async () => {
    const response = await fetch(uploadUrl, {
      method: "HEAD",
      headers: { "Tus-Resumable": TUS_VERSION, Authorization: bearer() },
      signal,
    });
    if (!response.ok) {
      throw new TusError(`tus HEAD failed: ${response.status}`, response.status);
    }
    const offset = response.headers.get("Upload-Offset");
    return offset ? Number.parseInt(offset, 10) : 0;
  }, signal);
}

async function uploadChunk(
  uploadUrl: string,
  chunk: Blob,
  offset: number,
  signal: AbortSignal,
): Promise<number> {
  return withRetry(async () => {
    const response = await fetch(uploadUrl, {
      method: "PATCH",
      headers: {
        "Tus-Resumable": TUS_VERSION,
        "Upload-Offset": String(offset),
        "Content-Type": "application/offset+octet-stream",
        Authorization: bearer(),
      },
      body: chunk,
      signal,
    });
    if (response.status !== 204) {
      throw new TusError(
        `tus PATCH @${offset} failed: ${response.status}`,
        response.status,
      );
    }
    const nextOffset = response.headers.get("Upload-Offset");
    return nextOffset ? Number.parseInt(nextOffset, 10) : offset + chunk.size;
  }, signal);
}

export async function uploadFile(
  options: TusUploadOptions,
  signal: AbortSignal,
): Promise<string> {
  const uploadUrl = await createUpload(options);
  options.onCreate?.(uploadUrl);
  let offset = 0;
  while (offset < options.file.size) {
    if (signal.aborted) {
      throw new DOMException("Upload aborted", "AbortError");
    }
    const end = Math.min(offset + CHUNK_SIZE_BYTES, options.file.size);
    const chunk = options.file.slice(offset, end);
    offset = await uploadChunk(uploadUrl, chunk, offset, signal);
    options.onProgress?.(offset);
  }
  return uploadUrl;
}

export async function resumeUpload(
  uploadUrl: string,
  options: TusUploadOptions,
  signal: AbortSignal,
): Promise<void> {
  let offset = await getOffset(uploadUrl, signal);
  options.onProgress?.(offset);
  while (offset < options.file.size) {
    if (signal.aborted) {
      throw new DOMException("Upload aborted", "AbortError");
    }
    const end = Math.min(offset + CHUNK_SIZE_BYTES, options.file.size);
    const chunk = options.file.slice(offset, end);
    offset = await uploadChunk(uploadUrl, chunk, offset, signal);
    options.onProgress?.(offset);
  }
}
