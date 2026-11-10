<script lang="ts">
  import FilePlus from "lucide-svelte/icons/file-plus";
  import FolderPlus from "lucide-svelte/icons/folder-plus";
  import FolderUp from "lucide-svelte/icons/folder-up";
  import DownloadIcon from "lucide-svelte/icons/download";
  import Search from "lucide-svelte/icons/search";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import LayoutList from "lucide-svelte/icons/list";
  import ArrowUp from "lucide-svelte/icons/arrow-up";
  import { files } from "$lib/store/files.svelte";
  import { hasPayload } from "./drag-drop";
  import ToolbarOverflow from "./ToolbarOverflow.svelte";

  interface Props {
    loc: string[];
    searchQuery: string;
    onSearch: (value: string) => void;
    onUp: () => void;
    onUpload: () => void;
    onUploadFolder: () => void;
    onNewFolder: () => void;
    onDownload: () => void;
    downloadCount: number;
    downloadLabel: string;
    downloadBusy: boolean;
    onDropToUp: (event: DragEvent) => void;
  }

  let {
    loc,
    searchQuery,
    onSearch,
    onUp,
    onUpload,
    onUploadFolder,
    onNewFolder,
    onDownload,
    downloadCount,
    downloadLabel,
    downloadBusy,
    onDropToUp,
  }: Props = $props();

  let upDropActive = $state(false);
  // Mobile-only: search bar reveals below the toolbar when toggled from the
  // overflow menu. Desktop keeps the inline search input always visible.
  let searchOpen = $state(false);

  function handleInput(event: Event) {
    const input = event.target as HTMLInputElement;
    onSearch(input.value);
  }

  function onUpDragOver(event: DragEvent) {
    if (loc.length === 0 || !hasPayload(event)) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    upDropActive = true;
  }

  function onUpDragLeave() {
    upDropActive = false;
  }

  function onUpDrop(event: DragEvent) {
    upDropActive = false;
    if (loc.length === 0) return;
    onDropToUp(event);
  }
</script>

{#snippet searchInput(id: string)}
  <label for={id} class="sr-only">Search current folder</label>
  <div class="relative flex-1 min-w-0">
    <Search size="16" class="absolute left-2.5 top-1/2 -translate-y-1/2 text-fg-muted pointer-events-none" />
    <input
      {id}
      type="search"
      name="folder-q"
      autocomplete="off"
      autocorrect="off"
      spellcheck="false"
      value={searchQuery}
      oninput={handleInput}
      placeholder="Search current folder…"
      class="sg-input sg-input-icon"
    />
  </div>
{/snippet}

<header class="flex items-center gap-1.5 md:gap-2 px-2 md:px-gutter min-h-12 py-1 border-b border-border-default bg-bg-surface">
  <button
    type="button"
    class="sg-btn sg-btn-ghost sg-btn-icon {upDropActive ? 'drop-target' : ''}"
    onclick={onUp}
    disabled={loc.length === 0}
    aria-label="Up one directory"
    title={loc.length === 0 ? "Already at the top folder" : "Up one directory"}
    ondragover={onUpDragOver}
    ondragleave={onUpDragLeave}
    ondrop={onUpDrop}
  >
    <ArrowUp size="18" />
  </button>

  <button type="button" class="sg-btn sg-btn-primary" onclick={onUpload} aria-label="Upload">
    <FilePlus size="16" />
    <span class="hidden md:inline">Upload</span>
  </button>

  <button type="button" class="hidden md:inline-flex sg-btn" onclick={onUploadFolder}>
    <FolderUp size="16" />
    <span>Upload folder</span>
  </button>

  <button type="button" class="hidden md:inline-flex sg-btn" onclick={onNewFolder}>
    <FolderPlus size="16" />
    <span>New folder</span>
  </button>

  <button
    type="button"
    class="hidden md:inline-flex sg-btn"
    onclick={onDownload}
    disabled={downloadCount === 0 || downloadBusy}
    aria-label={downloadCount === 0 ? "Download" : downloadLabel}
    title={downloadCount === 0 ? "Select files or folders to download" : downloadLabel}
  >
    <DownloadIcon size="16" />
    <span class="hidden lg:inline">Download</span>
  </button>

  <div class="hidden md:flex flex-1 min-w-0 max-w-md ml-2">
    {@render searchInput("folder-q-desktop")}
  </div>

  <div class="ml-auto flex items-center gap-1.5 md:gap-2">
    <div class="sg-seg" role="group" aria-label="File view">
      <button
        type="button"
        class="sg-seg-btn"
        onclick={() => files.setViewMode("grid")}
        aria-label="Grid view"
        aria-pressed={files.viewMode === "grid"}
      >
        <LayoutGrid size="16" />
      </button>
      <button
        type="button"
        class="sg-seg-btn"
        onclick={() => files.setViewMode("list")}
        aria-label="List view"
        aria-pressed={files.viewMode === "list"}
      >
        <LayoutList size="16" />
      </button>
    </div>

    <div class="md:hidden">
      <ToolbarOverflow
        {onUploadFolder}
        {onNewFolder}
        onToggleSearch={() => (searchOpen = !searchOpen)}
      />
    </div>
  </div>
</header>

{#if searchOpen}
  <div class="md:hidden flex items-center px-2 py-1.5 border-b border-border-default bg-bg-surface">
    {@render searchInput("folder-q-mobile")}
  </div>
{/if}

<style>
  .drop-target {
    background: var(--color-bg-selected);
    color: var(--color-accent);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
</style>
