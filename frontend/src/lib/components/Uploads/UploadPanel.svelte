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
    uploading: "text-fg-accent",
    paused: "text-fg-warning",
    complete: "text-fg-success",
    error: "text-fg-danger",
    cancelled: "text-fg-disabled",
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
    class="fixed top-12 right-0 bottom-7 w-[360px] bg-bg-surface border-l border-border-default z-40 flex flex-col shadow-[0_0_24px_rgba(0,0,0,0.4)]"
  >
    <header class="flex items-center justify-between h-10 px-3 border-b border-border-default">
      <span class="text-sm font-semibold text-fg-primary">Uploads</span>
      <button
        type="button"
        class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-primary hover:bg-bg-hover"
        onclick={onClose}
        aria-label="Close uploads panel"
      >
        <X size="14" />
      </button>
    </header>

    <div class="flex-1 overflow-y-auto p-2 space-y-1.5">
      {#if uploads.list.length === 0}
        <div class="text-xs text-fg-muted text-center py-8">No uploads yet.</div>
      {/if}

      {#each uploads.list as upload (upload.id)}
        <div class="p-2 rounded-md bg-bg-elevated">
          <div class="flex items-center gap-2 mb-1">
            <span class="text-xs font-medium truncate flex-1 text-fg-primary">
              {upload.filename}
            </span>
            <span class="text-xs {STATUS_CLASS[upload.status]} font-mono">
              {upload.status}
            </span>
          </div>

          <div class="flex items-center gap-2 text-xs text-fg-muted mb-1.5">
            <span class="font-mono">
              {formatBytes(upload.uploadedBytes)} / {formatBytes(upload.totalBytes)}
            </span>
            <span class="ml-auto font-mono tabular-nums">{percentOf(upload)}%</span>
          </div>

          <div class="h-1 rounded-full bg-bg-base overflow-hidden">
            <div
              class="h-full transition-[width] duration-200 {upload.status === 'error'
                ? 'bg-fg-danger'
                : upload.status === 'complete'
                  ? 'bg-fg-success'
                  : 'bg-accent'}"
              style="width: {percentOf(upload)}%;"
            ></div>
          </div>

          {#if upload.status === "uploading"}
            <div class="flex justify-end gap-1 mt-1.5">
              <button
                type="button"
                class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-warning hover:bg-bg-hover"
                onclick={() => pauseUpload(upload.id)}
                aria-label="Pause"
              >
                <Pause size="12" />
              </button>
              <button
                type="button"
                class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-danger hover:bg-bg-hover"
                onclick={() => cancelUpload(upload.id)}
                aria-label="Cancel"
              >
                <XCircle size="12" />
              </button>
            </div>
          {:else if upload.status === "paused"}
            <div class="flex justify-end gap-1 mt-1.5">
              <button
                type="button"
                class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-success hover:bg-bg-hover"
                onclick={() => resumeUpload(upload.id)}
                aria-label="Resume"
              >
                <Play size="12" />
              </button>
              <button
                type="button"
                class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-danger hover:bg-bg-hover"
                onclick={() => cancelUpload(upload.id)}
                aria-label="Cancel"
              >
                <XCircle size="12" />
              </button>
            </div>
          {:else if upload.status === "complete"}
            <div class="flex items-center gap-1 mt-1 text-fg-success text-xs">
              <CheckCircle2 size="12" />
              <span>Done</span>
            </div>
          {:else if upload.status === "error"}
            <div class="flex items-start gap-1 mt-1 text-fg-danger text-xs">
              <AlertCircle size="12" class="mt-0.5 shrink-0" />
              <span class="break-words">{upload.errorMessage ?? "Failed"}</span>
            </div>
          {/if}
        </div>
      {/each}
    </div>

    {#if uploads.completed.length > 0}
      <footer class="border-t border-border-default p-2">
        <button
          type="button"
          class="w-full h-7 text-xs text-fg-muted hover:text-fg-primary rounded hover:bg-bg-hover"
          onclick={() => uploads.clearCompleted()}
        >
          Clear completed
        </button>
      </footer>
    {/if}
  </aside>
{/if}
