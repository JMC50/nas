// Subscribes to the uploads store and drives queued items through the tus
// client. Survives navigation because the store is module-scoped and lives
// outside any component's lifecycle.

import { uploads } from "$lib/store/uploads.svelte";
import { notifications } from "$lib/store/notifications.svelte";
import {
  uploadFile,
  resumeUpload as resumeTusUpload,
  isTransient,
} from "$lib/tus-client";
import type { Upload } from "$lib/types";

const MAX_CONCURRENT = 3;
const UPLOAD_DONE_TOAST_MS = 3000;

const controllers = new Map<string, AbortController>();
let activeCount = 0;

function abortable(id: string): AbortController {
  const controller = new AbortController();
  controllers.set(id, controller);
  return controller;
}

/** Routes an upload failure into the right store state + notification.
 * AbortError → silent (user-initiated). Transient (network/5xx after retry
 * exhausted) → paused with tusUrl/uploadedBytes preserved so a resume picks
 * up from the last server-confirmed offset. Anything else (4xx etc) → error. */
function handleUploadFailure(upload: Upload, error: unknown) {
  if ((error as DOMException).name === "AbortError") return;
  const message = (error as Error).message;
  if (isTransient(error)) {
    uploads.setStatus(upload.id, "paused", { errorMessage: message });
    notifications.error(
      `Upload paused after retries (resume to continue): ${upload.filename}`,
    );
    return;
  }
  uploads.setStatus(upload.id, "error", { errorMessage: message });
  notifications.error(`Upload failed: ${upload.filename} — ${message}`);
}

/** Build the tus-client option bag for an upload. `onCreate` saves the tus
 * Location URL the moment we have it — required for a genuine resume after
 * a transient failure or user pause (without it we'd restart from offset 0). */
function buildTusOptions(upload: Upload) {
  return {
    file: upload.file,
    loc: upload.loc,
    filename: upload.filename,
    onProgress: (uploadedBytes: number) => uploads.setProgress(upload.id, uploadedBytes),
    onCreate: (uploadUrl: string) => uploads.update(upload.id, { tusUrl: uploadUrl }),
  };
}

async function runUpload(upload: Upload) {
  activeCount++;
  uploads.setStatus(upload.id, "uploading");
  const controller = abortable(upload.id);
  const tusOptions = buildTusOptions(upload);

  try {
    if (upload.tusUrl && upload.uploadedBytes > 0) {
      // Real resume: HEAD the server to learn the true offset, then PATCH
      // the remaining chunks. withRetry inside the tus client absorbs
      // transient network failures with exponential backoff.
      await resumeTusUpload(upload.tusUrl, tusOptions, controller.signal);
    } else {
      // Fresh upload (or a paused one without a saved tusUrl — falls back
      // to a clean restart, which matches pre-fix behavior).
      await uploadFile(tusOptions, controller.signal);
    }
    uploads.setStatus(upload.id, "complete", { uploadedBytes: upload.totalBytes });
    notifications.success(`Uploaded ${upload.filename}`, UPLOAD_DONE_TOAST_MS);
  } catch (error) {
    handleUploadFailure(upload, error);
  } finally {
    controllers.delete(upload.id);
    activeCount--;
    pump();
  }
}

function pump() {
  while (activeCount < MAX_CONCURRENT) {
    const next = uploads.list.find((upload) => upload.status === "queued");
    if (!next) return;
    void runUpload(next);
  }
}

// Callers (DragDropOverlay, future Explorer toolbar) call this after
// enqueueing one or more uploads to wake the worker.
export function pumpQueue() {
  pump();
}

export function cancelUpload(id: string) {
  const controller = controllers.get(id);
  if (controller) controller.abort();
  uploads.cancel(id);
}

export function pauseUpload(id: string) {
  const controller = controllers.get(id);
  if (controller) controller.abort();
  uploads.pause(id);
}

export function resumeUpload(id: string) {
  const upload = uploads.list.find((entry) => entry.id === id);
  // Real resume: if we have a saved tus URL, keep uploadedBytes so runUpload
  // takes the resumeTusUpload branch and continues from the last server-
  // confirmed offset. Without a tusUrl (legacy state / never created) we
  // must reset to 0 — the old behavior, preserved as fallback.
  if (upload?.tusUrl) {
    uploads.update(id, { status: "queued" });
  } else {
    uploads.update(id, { status: "queued", uploadedBytes: 0 });
  }
  pump();
}
