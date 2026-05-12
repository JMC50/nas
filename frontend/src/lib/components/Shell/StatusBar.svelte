<script lang="ts">
  import Upload from "lucide-svelte/icons/upload";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { uploads } from "$lib/store/uploads.svelte";
  import { mode } from "mode-watcher";

  const percentDisplay = $derived(Math.round(uploads.overallProgress * 100));

  function formatBytes(value: number): string {
    if (value < 1024) return `${value} B`;
    if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
    if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
    return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  }
</script>

<footer
  class="row-start-3 col-span-2 h-7 flex items-center justify-between gap-4 px-3 text-xs text-fg-muted border-t border-border-default bg-bg-surface"
>
  <div class="flex items-center gap-3 min-w-0">
    {#if uploads.hasActive}
      <div class="flex items-center gap-2 min-w-0">
        <Upload size="12" class="text-fg-accent" />
        <span class="truncate">
          {uploads.active.length} uploading · {formatBytes(uploads.uploadedBytes)} / {formatBytes(uploads.totalBytes)}
        </span>
        <div class="w-24 h-1 rounded-full bg-bg-elevated overflow-hidden">
          <div
            class="h-full bg-accent transition-[width] duration-200"
            style="width: {percentDisplay}%;"
          ></div>
        </div>
        <span class="font-mono tabular-nums">{percentDisplay}%</span>
      </div>
    {:else if uploads.failed.length > 0}
      <div class="flex items-center gap-2 text-fg-danger">
        <AlertCircle size="12" />
        <span>{uploads.failed.length} failed</span>
      </div>
    {:else if uploads.completed.length > 0}
      <div class="flex items-center gap-2 text-fg-success">
        <CheckCircle2 size="12" />
        <span>{uploads.completed.length} uploaded</span>
      </div>
    {:else}
      <span class="text-fg-disabled">Ready</span>
    {/if}
  </div>

  <div class="flex items-center gap-3 shrink-0">
    <span>{mode.current === "light" ? "Light" : "Dark"}</span>
    <span class="font-mono">v2.0</span>
  </div>
</footer>
