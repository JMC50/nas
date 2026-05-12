// Minimal tus 1.0.0 resumable upload client.
// Backend exposes the tus endpoint at /server/files/* (router.go).
// See https://tus.io/protocols/resumable-upload for protocol details.

import { auth } from "$lib/store/auth.svelte";

const CHUNK_SIZE_BYTES = 5 * 1024 * 1024; // 5 MiB
const TUS_VERSION = "1.0.0";

export interface TusUploadOptions {
  file: File | Blob;
  loc: string;
  filename: string;
  onProgress?: (uploadedBytes: number) => void;
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

function encodeBase64(value: string): string {
  return btoa(unescape(encodeURIComponent(value)));
}

function buildMetadata(loc: string, filename: string): string {
  return [
    `loc ${encodeBase64(loc)}`,
    `name ${encodeBase64(filename)}`,
    `filename ${encodeBase64(filename)}`,
  ].join(",");
}

function tokenizedUrl(uploadUrl: string): string {
  if (uploadUrl.includes("token=")) return uploadUrl;
  const separator = uploadUrl.includes("?") ? "&" : "?";
  return `${uploadUrl}${separator}token=${encodeURIComponent(auth.token)}`;
}

async function createUpload(options: TusUploadOptions): Promise<string> {
  const response = await fetch(
    `/server/files?token=${encodeURIComponent(auth.token)}`,
    {
      method: "POST",
      headers: {
        "Tus-Resumable": TUS_VERSION,
        "Upload-Length": String(options.file.size),
        "Upload-Metadata": buildMetadata(options.loc, options.filename),
        "Content-Type": "application/offset+octet-stream",
      },
    },
  );

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

async function getOffset(uploadUrl: string): Promise<number> {
  const response = await fetch(tokenizedUrl(uploadUrl), {
    method: "HEAD",
    headers: { "Tus-Resumable": TUS_VERSION },
  });
  if (!response.ok) {
    throw new TusError(`tus HEAD failed: ${response.status}`, response.status);
  }
  const offset = response.headers.get("Upload-Offset");
  return offset ? Number.parseInt(offset, 10) : 0;
}

async function uploadChunk(
  uploadUrl: string,
  chunk: Blob,
  offset: number,
  signal: AbortSignal,
): Promise<number> {
  const response = await fetch(tokenizedUrl(uploadUrl), {
    method: "PATCH",
    headers: {
      "Tus-Resumable": TUS_VERSION,
      "Upload-Offset": String(offset),
      "Content-Type": "application/offset+octet-stream",
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
}

export async function uploadFile(
  options: TusUploadOptions,
  signal: AbortSignal,
): Promise<string> {
  const uploadUrl = await createUpload(options);
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
  let offset = await getOffset(uploadUrl);
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
