<script lang="ts">
  import X from "lucide-svelte/icons/x";
  import Pause from "lucide-svelte/icons/pause";
  import Play from "lucide-svelte/icons/play";
  import XCircle from "lucide-svelte/icons/x-circle";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { uploads } from "$lib/store/uploads.svelte";
  import { cancelUpload, pauseUpload, resumeUpload } from "$lib/upload-worker";
  import type { Upload, UploadStatus } from "$lib/types";

  interface Props {
    open: boolean;
    onClose: () => void;
  }

  let { open, onClose }: Props = $props();

  const STATUS_CLASS: Record<UploadStatus, string> = {
    queued: "text-fg-muted",
    uploading: "text-fg-info",
    paused: "text-fg-warning",
    complete: "text-fg-success",
    error: "text-fg-danger",
    cancelled: "text-fg-muted",
  };

  function formatBytes(value: number): string {
    if (value < 1024) return `${value} B`;
    if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
    if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
    return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  }

  function percentOf(upload: Upload): number {
    if (upload.totalBytes === 0) return 0;
    return Math.round((upload.uploadedBytes / upload.totalBytes) * 100);
  }
</script>

{#if open}
  <aside
    class="fixed inset-x-0 bottom-0 top-auto max-h-[70dvh] w-full rounded-t-panel border-t border-border-default pb-[env(safe-area-inset-bottom)] animate-sheet md:animate-fade md:top-12 lg:top-0 md:right-0 md:inset-x-auto md:bottom-7 md:max-h-none md:w-[360px] md:rounded-none md:border-t-0 md:border-l md:pb-0 bg-bg-surface z-40 flex flex-col shadow-overlay"
    aria-label="Uploads"
  >
    <div class="md:hidden w-10 h-1 rounded-full bg-border-strong mx-auto mt-2 mb-1" aria-hidden="true"></div>
    <header class="min-h-12 flex items-center justify-between gap-2 pl-4 pr-2 border-b border-border-default">
      <h2 class="text-md font-semibold text-fg-primary">Uploads</h2>
      <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={onClose} aria-label="Close uploads panel">
        <X size="18" />
      </button>
    </header>

    <ul class="flex-1 overflow-y-auto">
      {#if uploads.list.length === 0}
        <li class="text-sm text-fg-muted text-center py-10">No uploads yet.</li>
      {/if}

      {#each uploads.list as upload (upload.id)}
        <li class="px-4 py-3 border-b border-border-default">
          <div class="flex items-center gap-2">
            <span class="text-sm font-semibold truncate flex-1 min-w-0 text-fg-primary" title={upload.filename}>
              {upload.filename}
            </span>
            <span class="sg-chip {STATUS_CLASS[upload.status]}">{upload.status}</span>
          </div>

          <div class="flex items-center gap-2 mt-1.5 text-xs text-fg-muted tabular-nums">
            <span>{formatBytes(upload.uploadedBytes)} / {formatBytes(upload.totalBytes)}</span>
            <span class="ml-auto">{percentOf(upload)}%</span>
          </div>

          <div class="h-1 mt-2 rounded-full bg-bg-overlay overflow-hidden" role="progressbar" aria-label="{upload.filename} progress" aria-valuemin="0" aria-valuemax="100" aria-valuenow={percentOf(upload)}>
            <div
              class="h-full {upload.status === 'error' ? 'bg-fg-danger' : upload.status === 'complete' ? 'bg-fg-success' : 'bg-fg-info'}"
              style="width: {percentOf(upload)}%;"
            ></div>
          </div>

          {#if upload.status === "uploading"}
            <div class="flex justify-end gap-1 mt-2">
              <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={() => pauseUpload(upload.id)} aria-label="Pause">
                <Pause size="16" />
              </button>
              <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={() => cancelUpload(upload.id)} aria-label="Cancel">
                <XCircle size="16" />
              </button>
            </div>
          {:else if upload.status === "paused"}
            <div class="flex justify-end gap-1 mt-2">
              <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={() => resumeUpload(upload.id)} aria-label="Resume">
                <Play size="16" />
              </button>
              <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={() => cancelUpload(upload.id)} aria-label="Cancel">
                <XCircle size="16" />
              </button>
            </div>
          {:else if upload.status === "complete"}
            <div class="flex items-center gap-1.5 mt-2 text-fg-success text-xs">
              <CheckCircle2 size="14" />
              <span>Done</span>
            </div>
          {:else if upload.status === "error"}
            <div class="flex items-start gap-1.5 mt-2 text-fg-danger text-xs" role="alert">
              <AlertCircle size="14" class="mt-px shrink-0" />
              <span class="break-words min-w-0">{upload.errorMessage ?? "Failed"}</span>
            </div>
          {/if}
        </li>
      {/each}
    </ul>

    {#if uploads.completed.length > 0}
      <footer class="border-t border-border-default p-2">
        <button type="button" class="sg-btn sg-btn-ghost w-full" onclick={() => uploads.clearCompleted()}>
          Clear completed
        </button>
      </footer>
    {/if}
  </aside>
{/if}
