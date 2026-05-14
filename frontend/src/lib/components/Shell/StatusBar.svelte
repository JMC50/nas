<script lang="ts">
  import Upload from "lucide-svelte/icons/upload";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { uploads } from "$lib/store/uploads.svelte";
  import { ui } from "$lib/store/ui.svelte";
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
  class="row-start-3 col-span-1 md:col-span-2 h-7 flex items-center justify-between gap-4 px-3 text-xs text-fg-muted border-t border-border-default bg-bg-surface pb-[env(safe-area-inset-bottom)] md:pb-0"
>
  <button
    type="button"
    class="flex items-center gap-3 min-w-0 max-w-[60%] hover:bg-bg-hover px-2 h-full rounded transition-colors text-left"
    onclick={() => ui.toggleUploadsPanel()}
    aria-label="Open uploads panel"
  >
    {#if uploads.hasActive}
      <Upload size="12" class="text-fg-accent shrink-0" />
      <span class="truncate">
        {uploads.active.length} uploading · {formatBytes(uploads.uploadedBytes)} / {formatBytes(uploads.totalBytes)}
      </span>
      <div class="w-24 h-1 rounded-full bg-bg-elevated overflow-hidden shrink-0">
        <div
          class="h-full bg-accent transition-[width] duration-200"
          style="width: {percentDisplay}%;"
        ></div>
      </div>
      <span class="font-mono tabular-nums shrink-0">{percentDisplay}%</span>
    {:else if uploads.failed.length > 0}
      <AlertCircle size="12" class="text-fg-danger shrink-0" />
      <span class="text-fg-danger">{uploads.failed.length} failed</span>
    {:else if uploads.completed.length > 0}
      <CheckCircle2 size="12" class="text-fg-success shrink-0" />
      <span class="text-fg-success">{uploads.completed.length} uploaded</span>
    {:else}
      <span class="text-fg-disabled">Ready</span>
    {/if}
  </button>

  <div class="flex items-center gap-3 shrink-0">
    <span>{mode.current === "light" ? "Light" : "Dark"}</span>
    <span class="font-mono" title="App version">v{__APP_VERSION__}</span>
  </div>
</footer>
