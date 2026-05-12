// Subscribes to the uploads store and drives queued items through the tus
// client. Survives navigation because the store is module-scoped and lives
// outside any component's lifecycle.

import { uploads } from "$lib/store/uploads.svelte";
import { notifications } from "$lib/store/notifications.svelte";
import { uploadFile } from "$lib/tus-client";
import type { Upload } from "$lib/types";

const MAX_CONCURRENT = 3;

const controllers = new Map<string, AbortController>();
let activeCount = 0;

function abortable(id: string): AbortController {
  const controller = new AbortController();
  controllers.set(id, controller);
  return controller;
}

async function runUpload(upload: Upload) {
  activeCount++;
  uploads.setStatus(upload.id, "uploading");
  const controller = abortable(upload.id);

  try {
    await uploadFile(
      {
        file: upload.file,
        loc: upload.loc,
        filename: upload.filename,
        onProgress: (uploadedBytes) => uploads.setProgress(upload.id, uploadedBytes),
      },
      controller.signal,
    );
    uploads.setStatus(upload.id, "complete", { uploadedBytes: upload.totalBytes });
    notifications.success(`Uploaded ${upload.filename}`, 3000);
  } catch (error) {
    if ((error as DOMException).name === "AbortError") {
      // Cancellation is intentional; do not surface as error.
    } else {
      uploads.setStatus(upload.id, "error", {
        errorMessage: (error as Error).message,
      });
      notifications.error(`Upload failed: ${upload.filename} — ${(error as Error).message}`);
    }
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
export function pumpUploadQueue() {
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
  // Re-enqueue: mark as queued so pump picks it back up.
  uploads.update(id, { status: "queued", uploadedBytes: 0 });
  pump();
}
