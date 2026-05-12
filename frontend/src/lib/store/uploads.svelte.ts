import type { Upload, UploadStatus } from "$lib/types";

const COMPLETED_RETENTION_MS = 60_000;

function randomId() {
  return crypto.randomUUID();
}

class UploadsStore {
  list = $state<Upload[]>([]);

  active = $derived<Upload[]>(
    this.list.filter(
      (upload) =>
        upload.status === "uploading" ||
        upload.status === "paused" ||
        upload.status === "queued",
    ),
  );

  completed = $derived<Upload[]>(
    this.list.filter((upload) => upload.status === "complete"),
  );

  failed = $derived<Upload[]>(
    this.list.filter(
      (upload) => upload.status === "error" || upload.status === "cancelled",
    ),
  );

  totalBytes = $derived<number>(
    this.active.reduce((sum, upload) => sum + upload.totalBytes, 0),
  );

  uploadedBytes = $derived<number>(
    this.active.reduce((sum, upload) => sum + upload.uploadedBytes, 0),
  );

  overallProgress = $derived<number>(
    this.totalBytes === 0 ? 0 : this.uploadedBytes / this.totalBytes,
  );

  hasActive = $derived<boolean>(this.active.length > 0);

  enqueue(input: { file: File | Blob; loc: string; filename: string }): Upload {
    const upload: Upload = {
      id: randomId(),
      filename: input.filename,
      loc: input.loc,
      totalBytes: input.file.size,
      uploadedBytes: 0,
      status: "queued",
      startedAt: Date.now(),
      file: input.file,
    };
    this.list = [...this.list, upload];
    return upload;
  }

  update(id: string, patch: Partial<Upload>) {
    this.list = this.list.map((upload) =>
      upload.id === id ? { ...upload, ...patch } : upload,
    );
  }

  setStatus(id: string, status: UploadStatus, extras?: Partial<Upload>) {
    this.update(id, { status, ...extras });
    if (status === "complete") {
      const completedAt = Date.now();
      this.update(id, { completedAt });
      setTimeout(() => this.remove(id), COMPLETED_RETENTION_MS);
    }
  }

  setProgress(id: string, uploadedBytes: number) {
    this.update(id, { uploadedBytes });
  }

  pause(id: string) {
    this.setStatus(id, "paused");
  }

  resume(id: string) {
    this.setStatus(id, "uploading");
  }

  cancel(id: string) {
    this.setStatus(id, "cancelled");
  }

  remove(id: string) {
    this.list = this.list.filter((upload) => upload.id !== id);
  }

  clearCompleted() {
    this.list = this.list.filter((upload) => upload.status !== "complete");
  }
}

export const uploads = new UploadsStore();
