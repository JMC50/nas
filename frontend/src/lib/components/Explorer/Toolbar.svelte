<script lang="ts">
  import FilePlus from "lucide-svelte/icons/file-plus";
  import FolderPlus from "lucide-svelte/icons/folder-plus";
  import FolderUp from "lucide-svelte/icons/folder-up";
  import Search from "lucide-svelte/icons/search";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import LayoutList from "lucide-svelte/icons/list";
  import ArrowUp from "lucide-svelte/icons/arrow-up";
  import { files } from "$lib/store/files.svelte";
  import { hasPayload } from "./drag-drop";

  interface Props {
    loc: string[];
    searchQuery: string;
    onSearch: (value: string) => void;
    onUp: () => void;
    onUpload: () => void;
    onUploadFolder: () => void;
    onNewFolder: () => void;
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
    onDropToUp,
  }: Props = $props();

  let upDropActive = $state(false);

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

<header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
  <button
    type="button"
    class="inline-flex items-center justify-center w-8 h-8 rounded-md text-fg-muted hover:text-fg-primary hover:bg-bg-hover disabled:opacity-40 disabled:hover:bg-transparent transition-colors {upDropActive
      ? 'bg-bg-elevated text-fg-accent ring-1 ring-accent'
      : ''}"
    onclick={onUp}
    disabled={loc.length === 0}
    aria-label="Up one directory"
    ondragover={onUpDragOver}
    ondragleave={onUpDragLeave}
    ondrop={onUpDrop}
  >
    <ArrowUp size="16" />
  </button>

  <button
    type="button"
    class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-primary hover:bg-bg-hover text-xs font-medium transition-colors"
    onclick={onUpload}
  >
    <FilePlus size="14" />
    <span>Upload</span>
  </button>

  <button
    type="button"
    class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-primary hover:bg-bg-hover text-xs font-medium transition-colors"
    onclick={onUploadFolder}
  >
    <FolderUp size="14" />
    <span>Upload folder</span>
  </button>

  <button
    type="button"
    class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-primary hover:bg-bg-hover text-xs font-medium transition-colors"
    onclick={onNewFolder}
  >
    <FolderPlus size="14" />
    <span>New folder</span>
  </button>

  <div
    class="flex items-center gap-1.5 ml-2 h-8 px-2.5 rounded-md bg-bg-elevated flex-1 min-w-0 max-w-md"
  >
    <Search size="13" class="text-fg-muted shrink-0" />
    <input
      type="text"
      value={searchQuery}
      oninput={handleInput}
      placeholder="Search current folder…"
      class="flex-1 bg-transparent text-xs text-fg-primary placeholder:text-fg-muted outline-none min-w-0"
    />
  </div>

  <div class="ml-auto flex items-center gap-0.5 p-0.5 rounded-md bg-bg-elevated">
    <button
      type="button"
      class="inline-flex items-center justify-center w-7 h-7 rounded {files.viewMode === 'grid'
        ? 'bg-accent text-accent-fg'
        : 'text-fg-muted hover:text-fg-primary'}"
      onclick={() => files.setViewMode('grid')}
      aria-label="Grid view"
    >
      <LayoutGrid size="13" />
    </button>
    <button
      type="button"
      class="inline-flex items-center justify-center w-7 h-7 rounded {files.viewMode === 'list'
        ? 'bg-accent text-accent-fg'
        : 'text-fg-muted hover:text-fg-primary'}"
      onclick={() => files.setViewMode('list')}
      aria-label="List view"
    >
      <LayoutList size="13" />
    </button>
  </div>
</header>
