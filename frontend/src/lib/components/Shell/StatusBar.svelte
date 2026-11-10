<script lang="ts">
  import Upload from "lucide-svelte/icons/upload";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { uploads } from "$lib/store/uploads.svelte";
  import { ui } from "$lib/store/ui.svelte";

  const percentDisplay = $derived(Math.round(uploads.overallProgress * 100));

  function formatBytes(value: number): string {
    if (value < 1024) return `${value} B`;
    if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
    if (value < 1024 * 1024 * 1024) return `${(value / (1024 * 1024)).toFixed(1)} MB`;
    return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
  }
</script>

<footer
  class="row-start-3 col-start-1 lg:row-start-2 lg:col-start-2 h-7 pointer-coarse:h-11 min-w-0 flex items-center justify-between gap-3 pl-1 pr-3 text-xs text-fg-muted border-t border-border-default bg-bg-surface"
>
  <button
    type="button"
    class="flex items-center gap-2 min-w-0 h-full px-2 rounded-chip hover:bg-bg-hover hover:text-fg-primary transition-colors text-left focus-visible:outline-offset-[-2px]"
    onclick={() => ui.toggleUploadsPanel()}
    aria-label="Open uploads panel"
    aria-expanded={ui.uploadsPanelOpen}
  >
    {#if uploads.hasActive}
      <Upload size="13" class="text-fg-info shrink-0" />
      <span class="truncate">
        {uploads.active.length} uploading · {formatBytes(uploads.uploadedBytes)} / {formatBytes(uploads.totalBytes)}
      </span>
      <span class="hidden md:block w-24 h-1 rounded-full bg-bg-overlay overflow-hidden shrink-0" aria-hidden="true">
        <span class="block h-full bg-fg-info" style="width: {percentDisplay}%;"></span>
      </span>
      <span class="tabular-nums shrink-0">{percentDisplay}%</span>
    {:else if uploads.failed.length > 0}
      <AlertCircle size="13" class="text-fg-danger shrink-0" />
      <span class="text-fg-danger truncate">{uploads.failed.length} failed</span>
    {:else if uploads.completed.length > 0}
      <CheckCircle2 size="13" class="text-fg-success shrink-0" />
      <span class="text-fg-success truncate">{uploads.completed.length} uploaded</span>
    {:else}
      <span>Ready</span>
    {/if}
  </button>

  <span class="shrink-0 font-mono" title="App version">v{__APP_VERSION__}</span>
</footer>
