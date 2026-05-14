<script lang="ts">
  import X from "lucide-svelte/icons/x";
  import { iconFor, type FolderEntry } from "./icon-for";
  import { formatBytes, formatFullTime } from "./format";

  interface Props {
    entry: FolderEntry | null;
    loc: string[];
    onClose: () => void;
  }

  let { entry, loc, onClose }: Props = $props();

  const typeLabel = $derived(
    entry === null
      ? ""
      : entry.isFolder
        ? "Folder"
        : (entry.extensions || "file").toUpperCase(),
  );
  const sizeLabel = $derived(
    entry === null ? "" : entry.isFolder ? "—" : formatBytes(entry.size),
  );
  const fullPath = $derived(entry === null ? "" : [...loc, entry.name].join("/"));
</script>

{#if entry}
  {@const Icon = iconFor(entry)}
  <aside
    class="fixed inset-x-0 bottom-0 top-auto max-h-[70dvh] w-full rounded-t-lg border-t border-border-default pb-[env(safe-area-inset-bottom)] md:top-12 md:right-0 md:inset-x-auto md:bottom-7 md:max-h-none md:w-[360px] md:rounded-none md:border-t-0 md:border-l md:pb-0 bg-bg-surface z-40 flex flex-col shadow-[0_0_24px_rgba(0,0,0,0.4)]"
  >
    <div class="md:hidden w-12 h-1 bg-fg-muted/40 rounded-full mx-auto mt-2 mb-1" aria-hidden="true"></div>
    <header class="h-10 flex items-center justify-between px-3 border-b border-border-default">
      <span class="text-sm font-semibold text-fg-primary">File details</span>
      <button
        type="button"
        class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-primary hover:bg-bg-hover"
        onclick={onClose}
        aria-label="Close file details"
      >
        <X size="14" />
      </button>
    </header>

    <div class="flex-1 overflow-y-auto">
      <div class="h-8 px-3 flex items-center gap-2 text-xs">
        <Icon size="14" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
        <span class="font-mono text-fg-primary truncate" title={entry.name}>
          {entry.name}
        </span>
      </div>

      <div class="h-8 px-3 flex items-center gap-2 text-xs">
        <span class="w-20 text-fg-muted">Type</span>
        <span class="font-mono text-fg-primary truncate">{typeLabel}</span>
      </div>

      <div class="h-8 px-3 flex items-center gap-2 text-xs">
        <span class="w-20 text-fg-muted">Size</span>
        <span class="font-mono text-fg-primary truncate">{sizeLabel}</span>
      </div>

      <div class="h-8 px-3 flex items-center gap-2 text-xs">
        <span class="w-20 text-fg-muted">Modified</span>
        <span class="font-mono text-fg-primary truncate">
          {formatFullTime(entry.modifiedAt)}
        </span>
      </div>

      <div class="h-8 px-3 flex items-center gap-2 text-xs">
        <span class="w-20 text-fg-muted">Path</span>
        <span class="font-mono text-fg-primary truncate" title={fullPath}>
          {fullPath}
        </span>
      </div>
    </div>
  </aside>
{/if}
