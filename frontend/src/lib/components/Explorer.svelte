<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import FolderIcon from "lucide-svelte/icons/folder";
  import FilePlus from "lucide-svelte/icons/file-plus";
  import FolderPlus from "lucide-svelte/icons/folder-plus";
  import Search from "lucide-svelte/icons/search";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import LayoutList from "lucide-svelte/icons/list";
  import ArrowUp from "lucide-svelte/icons/arrow-up";
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Pencil from "lucide-svelte/icons/pencil";
  import Download from "lucide-svelte/icons/download";
  import Copy from "lucide-svelte/icons/copy";
  import HomeIcon from "lucide-svelte/icons/home";
  import FileIcon from "lucide-svelte/icons/file";
  import FileText from "lucide-svelte/icons/file-text";
  import ImageIcon from "lucide-svelte/icons/image";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import FileArchive from "lucide-svelte/icons/file-archive";
  import FileType from "lucide-svelte/icons/file-type";
  import { auth } from "$lib/store/auth.svelte";
  import { files } from "$lib/store/files.svelte";
  import { uploads } from "$lib/store/uploads.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { pumpUploadQueue } from "$lib/upload-worker";
  import { pickViewer } from "$lib/components/Viewers/registry";

  interface FolderEntry {
    name: string;
    isFolder: boolean;
    extensions: string;
  }

  let entries: FolderEntry[] = $state([]);
  let loading = $state(false);
  let error: string | null = $state(null);
  let searchQuery = $state("");
  let fileInputEl: HTMLInputElement;

  const filteredEntries = $derived(
    searchQuery
      ? entries.filter((entry) =>
          entry.name.toLowerCase().includes(searchQuery.toLowerCase()),
        )
      : entries,
  );

  const sortedEntries = $derived(
    [...filteredEntries].sort((a, b) => {
      if (a.isFolder !== b.isFolder) return a.isFolder ? -1 : 1;
      return a.name.localeCompare(b.name);
    }),
  );

  const breadcrumb = $derived(files.currentLoc);

  async function fetchEntries() {
    loading = true;
    error = null;
    try {
      const loc = "/" + files.currentLoc.join("/");
      const response = await fetch(
        `/server/readFolder?loc=${encodeURIComponent(loc)}&token=${encodeURIComponent(auth.token)}`,
      );
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      entries = (data.files ?? data ?? []) as FolderEntry[];
    } catch (err) {
      error = (err as Error).message;
      entries = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    void files.currentLoc;
    fetchEntries();
  });

  function navigateInto(name: string) {
    files.navigateInto(name);
  }

  function navigateUp() {
    files.navigateUp();
  }

  function navigateToSegment(index: number) {
    files.setLoc(files.currentLoc.slice(0, index + 1));
  }

  function navigateRoot() {
    files.setLoc([]);
  }

  function openEntry(entry: FolderEntry) {
    if (entry.isFolder) {
      navigateInto(entry.name);
      return;
    }
    const viewerKind = pickViewer(entry.extensions);
    const loc = "/" + files.currentLoc.join("/");
    tabs.open({
      kind: viewerKind,
      title: entry.name,
      icon: viewerKind,
      payload: { loc, name: entry.name },
      closable: true,
    });
  }

  function triggerUpload() {
    fileInputEl?.click();
  }

  function onFilesPicked(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;
    const loc = "/" + files.currentLoc.join("/");
    for (const file of input.files) {
      uploads.enqueue({ file, loc, filename: file.name });
    }
    pumpUploadQueue();
    input.value = "";
  }

  async function newFolder() {
    const name = prompt("Folder name?");
    if (!name) return;
    const loc = "/" + files.currentLoc.join("/");
    try {
      const response = await fetch(
        `/server/makedir?name=${encodeURIComponent(name)}&loc=${encodeURIComponent(loc)}&token=${encodeURIComponent(auth.token)}`,
      );
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      notifications.success(`Created ${name}`);
      fetchEntries();
    } catch (err) {
      notifications.error(`Failed: ${(err as Error).message}`);
    }
  }

  async function deleteEntry(entry: FolderEntry) {
    if (!confirm(`Delete ${entry.name}?`)) return;
    const loc = "/" + files.currentLoc.join("/");
    try {
      const response = await fetch(
        `/server/forceDelete?name=${encodeURIComponent(entry.name)}&loc=${encodeURIComponent(loc)}&token=${encodeURIComponent(auth.token)}`,
      );
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      notifications.success(`Deleted ${entry.name}`);
      fetchEntries();
    } catch (err) {
      notifications.error(`Delete failed: ${(err as Error).message}`);
    }
  }

  async function renameEntry(entry: FolderEntry) {
    const next = prompt(`Rename "${entry.name}" to:`, entry.name);
    if (!next || next === entry.name) return;
    const loc = "/" + files.currentLoc.join("/");
    try {
      const response = await fetch(
        `/server/rename?loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(entry.name)}&newName=${encodeURIComponent(next)}&token=${encodeURIComponent(auth.token)}`,
      );
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      notifications.success(`Renamed to ${next}`);
      fetchEntries();
    } catch (err) {
      notifications.error(`Rename failed: ${(err as Error).message}`);
    }
  }

  function downloadEntry(entry: FolderEntry) {
    if (entry.isFolder) {
      notifications.info("Folder download not implemented yet.");
      return;
    }
    const loc = "/" + files.currentLoc.join("/");
    const url = `/server/download?loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(entry.name)}&token=${encodeURIComponent(auth.token)}`;
    window.open(url, "_blank");
  }

  function iconFor(entry: FolderEntry) {
    if (entry.isFolder) return FolderIcon;
    const ext = entry.extensions.toLowerCase();
    if (["jpg", "jpeg", "png", "gif", "webp", "avif", "bmp", "svg"].includes(ext)) return ImageIcon;
    if (["mp4", "webm", "mov", "mkv", "avi"].includes(ext)) return Film;
    if (["mp3", "wav", "ogg", "flac", "m4a"].includes(ext)) return Music;
    if (["zip", "tar", "gz", "rar", "7z"].includes(ext)) return FileArchive;
    if (ext === "pdf") return FileType;
    if (["md", "txt", "json", "yaml", "yml", "js", "ts", "go", "py", "rs", "html", "css", "svelte"].includes(ext)) return FileText;
    return FileIcon;
  }

  // Context menu state.
  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);
  let menuTarget: FolderEntry | null = $state(null);

  function openMenu(event: MouseEvent, entry: FolderEntry) {
    event.preventDefault();
    menuX = event.clientX;
    menuY = event.clientY;
    menuTarget = entry;
    menuOpen = true;
  }

  function closeMenu() {
    menuOpen = false;
    menuTarget = null;
  }

  function onWindowClick() {
    if (menuOpen) closeMenu();
  }

  function onKey(event: KeyboardEvent, entry: FolderEntry) {
    if (event.key === "Enter") {
      event.preventDefault();
      openEntry(entry);
    }
  }

  onMount(() => {
    window.addEventListener("click", onWindowClick);
  });

  onDestroy(() => {
    window.removeEventListener("click", onWindowClick);
  });
</script>

<section class="flex flex-col h-full bg-bg-base">
  <!-- Toolbar -->
  <header class="flex items-center gap-2 px-4 h-12 border-b border-border-default bg-bg-surface">
    <button
      type="button"
      class="inline-flex items-center justify-center w-8 h-8 rounded-md text-fg-muted hover:text-fg-primary hover:bg-bg-hover disabled:opacity-40 disabled:hover:bg-transparent transition-colors"
      onclick={navigateUp}
      disabled={files.currentLoc.length === 0}
      aria-label="Up one directory"
    >
      <ArrowUp size="16" />
    </button>

    <button
      type="button"
      class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-primary hover:bg-bg-hover text-xs font-medium transition-colors"
      onclick={triggerUpload}
    >
      <FilePlus size="14" />
      <span>Upload</span>
    </button>

    <button
      type="button"
      class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-primary hover:bg-bg-hover text-xs font-medium transition-colors"
      onclick={newFolder}
    >
      <FolderPlus size="14" />
      <span>New folder</span>
    </button>

    <div class="flex items-center gap-1.5 ml-2 h-8 px-2.5 rounded-md bg-bg-elevated flex-1 min-w-0 max-w-md">
      <Search size="13" class="text-fg-muted shrink-0" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Search current folder…"
        class="flex-1 bg-transparent text-xs text-fg-primary placeholder:text-fg-muted outline-none min-w-0"
      />
    </div>

    <div class="ml-auto flex items-center gap-0.5 p-0.5 rounded-md bg-bg-elevated">
      <button
        type="button"
        class="inline-flex items-center justify-center w-7 h-7 rounded {files.viewMode === 'grid' ? 'bg-accent text-accent-fg' : 'text-fg-muted hover:text-fg-primary'}"
        onclick={() => files.setViewMode('grid')}
        aria-label="Grid view"
      >
        <LayoutGrid size="13" />
      </button>
      <button
        type="button"
        class="inline-flex items-center justify-center w-7 h-7 rounded {files.viewMode === 'list' ? 'bg-accent text-accent-fg' : 'text-fg-muted hover:text-fg-primary'}"
        onclick={() => files.setViewMode('list')}
        aria-label="List view"
      >
        <LayoutList size="13" />
      </button>
    </div>

    <input
      type="file"
      multiple
      bind:this={fileInputEl}
      onchange={onFilesPicked}
      class="hidden"
    />
  </header>

  <!-- Breadcrumb -->
  <nav class="flex items-center gap-1 h-8 px-4 text-xs text-fg-muted border-b border-border-default bg-bg-base">
    <button
      type="button"
      class="inline-flex items-center gap-1 px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors"
      onclick={navigateRoot}
    >
      <HomeIcon size="12" />
      <span>root</span>
    </button>
    {#each breadcrumb as segment, index (index)}
      <ChevronRight size="11" class="text-fg-disabled shrink-0" />
      <button
        type="button"
        class="px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors truncate"
        onclick={() => navigateToSegment(index)}
      >
        {segment}
      </button>
    {/each}
  </nav>

  <!-- Body -->
  <div class="flex-1 overflow-auto">
    {#if error}
      <div class="p-6 text-sm text-fg-danger">
        Failed to load: {error}
      </div>
    {:else if loading && entries.length === 0}
      <div class="p-6 text-sm text-fg-muted">Loading…</div>
    {:else if sortedEntries.length === 0}
      <div class="p-12 text-center text-sm text-fg-muted">
        {searchQuery ? "No matches in this folder." : "Folder is empty."}
      </div>
    {:else if files.viewMode === 'grid'}
      <div class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-2 p-4">
        {#each sortedEntries as entry (entry.name)}
          {@const Icon = iconFor(entry)}
          <button
            type="button"
            class="group flex flex-col items-center gap-1.5 p-3 rounded-md text-fg-primary hover:bg-bg-hover transition-colors focus-visible:outline-2 focus-visible:outline-border-focus"
            ondblclick={() => openEntry(entry)}
            oncontextmenu={(event) => openMenu(event, entry)}
            onkeydown={(event) => onKey(event, entry)}
            title={entry.name}
          >
            <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
            <span class="text-xs truncate w-full text-center">{entry.name}</span>
          </button>
        {/each}
      </div>
    {:else}
      <table class="w-full text-sm">
        <thead>
          <tr class="text-xs text-fg-muted border-b border-border-default">
            <th class="text-left font-normal px-4 py-2">Name</th>
            <th class="text-left font-normal px-4 py-2 w-24">Type</th>
          </tr>
        </thead>
        <tbody>
          {#each sortedEntries as entry (entry.name)}
            {@const Icon = iconFor(entry)}
            <tr
              class="border-b border-border-default/40 hover:bg-bg-hover cursor-pointer"
              ondblclick={() => openEntry(entry)}
              oncontextmenu={(event) => openMenu(event, entry)}
            >
              <td class="px-4 py-1.5">
                <div class="flex items-center gap-2 min-w-0">
                  <Icon size="14" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
                  <span class="truncate text-fg-primary">{entry.name}</span>
                </div>
              </td>
              <td class="px-4 py-1.5 text-xs text-fg-muted">
                {entry.isFolder ? "Folder" : entry.extensions || "file"}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</section>

<!-- Context menu (portal-style) -->
{#if menuOpen && menuTarget}
  {@const target = menuTarget}
  <div
    class="fixed z-50 min-w-[160px] py-1 rounded-md bg-bg-overlay border border-border-strong shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
    style="left: {menuX}px; top: {menuY}px;"
    role="menu"
  >
    {#if !target.isFolder}
      <button
        type="button"
        class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
        onclick={() => { openEntry(target); closeMenu(); }}
        role="menuitem"
      >
        <FileText size="12" />
        Open
      </button>
    {/if}
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
      onclick={() => { downloadEntry(target); closeMenu(); }}
      role="menuitem"
    >
      <Download size="12" />
      Download
    </button>
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
      onclick={() => { renameEntry(target); closeMenu(); }}
      role="menuitem"
    >
      <Pencil size="12" />
      Rename
    </button>
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
      onclick={() => { notifications.info("Copy not wired yet."); closeMenu(); }}
      role="menuitem"
    >
      <Copy size="12" />
      Copy
    </button>
    <div class="h-px bg-border-default mx-1 my-1"></div>
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-danger hover:bg-bg-hover text-left"
      onclick={() => { deleteEntry(target); closeMenu(); }}
      role="menuitem"
    >
      <Trash2 size="12" />
      Delete
    </button>
  </div>
{/if}
