<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { files } from "$lib/store/files.svelte";
  import { uploads } from "$lib/store/uploads.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { pumpQueue } from "$lib/upload-worker";
  import Toolbar from "$lib/components/Explorer/Toolbar.svelte";
  import Breadcrumb from "$lib/components/Explorer/Breadcrumb.svelte";
  import FileGrid from "$lib/components/Explorer/FileGrid.svelte";
  import FileList from "$lib/components/Explorer/FileList.svelte";
  import ContextMenu from "$lib/components/Explorer/ContextMenu.svelte";
  import ForbiddenPanel from "$lib/components/ForbiddenPanel.svelte";
  import {
    locPath,
    readEntries,
    createFolder,
    deleteEntry,
    renameEntry,
    downloadUrl,
    openEntry,
    FetchError,
  } from "$lib/components/Explorer/actions";
  import type { FolderEntry } from "$lib/components/Explorer/icon-for";

  let entries: FolderEntry[] = $state([]);
  let loading = $state(false);
  let forbidden = $state(false);
  let errorMessage: string | null = $state(null);
  let searchQuery = $state("");
  let fileInputEl: HTMLInputElement;

  const filtered = $derived(
    searchQuery
      ? entries.filter((entry) => entry.name.toLowerCase().includes(searchQuery.toLowerCase()))
      : entries,
  );

  const sorted = $derived(
    [...filtered].sort((a, b) => {
      if (a.isFolder !== b.isFolder) return a.isFolder ? -1 : 1;
      return a.name.localeCompare(b.name);
    }),
  );

  async function refresh() {
    loading = true;
    errorMessage = null;
    forbidden = false;
    try {
      entries = await readEntries();
    } catch (cause) {
      if (cause instanceof FetchError && cause.status === 403) {
        forbidden = true;
      } else {
        errorMessage = (cause as Error).message;
      }
      entries = [];
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    void files.currentLoc;
    refresh();
  });

  function onPick(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.files || input.files.length === 0) return;
    const loc = locPath();
    for (const file of input.files) {
      uploads.enqueue({ file, loc, filename: file.name });
    }
    pumpQueue();
    input.value = "";
  }

  async function runMutation(action: () => Promise<void>, errorLabel: string) {
    try {
      await action();
      refresh();
    } catch (cause) {
      notifications.error(`${errorLabel}: ${(cause as Error).message}`);
    }
  }

  function newFolder() {
    const name = prompt("Folder name?");
    if (!name) return;
    runMutation(() => createFolder(name), "Create failed");
  }

  function remove(entry: FolderEntry) {
    if (!confirm(`Delete ${entry.name}?`)) return;
    runMutation(() => deleteEntry(entry), "Delete failed");
  }

  function rename(entry: FolderEntry) {
    const next = prompt(`Rename "${entry.name}" to:`, entry.name);
    if (!next || next === entry.name) return;
    runMutation(() => renameEntry(entry, next), "Rename failed");
  }

  function download(entry: FolderEntry) {
    if (entry.isFolder) {
      notifications.info("Folder download not implemented yet.");
      return;
    }
    window.open(downloadUrl(entry), "_blank");
  }

  let menuOpen = $state(false);
  let menuX = $state(0);
  let menuY = $state(0);
  let menuTarget: FolderEntry | null = $state(null);

  function showMenu(event: MouseEvent, entry: FolderEntry) {
    event.preventDefault();
    menuX = event.clientX;
    menuY = event.clientY;
    menuTarget = entry;
    menuOpen = true;
  }

  function hideMenu() {
    menuOpen = false;
    menuTarget = null;
  }

  onMount(() => {
    window.addEventListener("click", hideMenu);
  });

  onDestroy(() => {
    window.removeEventListener("click", hideMenu);
  });
</script>

<section class="flex flex-col h-full bg-bg-base">
  <Toolbar
    {searchQuery}
    onSearch={(value) => (searchQuery = value)}
    onUp={() => files.navigateUp()}
    onUpload={() => fileInputEl?.click()}
    onNewFolder={newFolder}
  />

  <Breadcrumb
    onGoto={(index) => files.setLoc(files.currentLoc.slice(0, index + 1))}
    onRoot={() => files.setLoc([])}
  />

  <input type="file" multiple bind:this={fileInputEl} onchange={onPick} class="hidden" />

  <div class="flex-1 overflow-auto">
    {#if forbidden}
      <ForbiddenPanel
        title="Files are locked"
        description="You don't have permission to browse files yet. Ask an administrator to grant you the View intent, or claim admin yourself with the admin password."
        onGranted={refresh}
      />
    {:else if errorMessage}
      <div class="p-6 text-sm text-fg-danger">Failed to load: {errorMessage}</div>
    {:else if loading && entries.length === 0}
      <div class="p-6 text-sm text-fg-muted">Loading…</div>
    {:else if sorted.length === 0}
      <div class="p-12 text-center text-sm text-fg-muted">
        {searchQuery ? "No matches in this folder." : "Folder is empty."}
      </div>
    {:else if files.viewMode === 'grid'}
      <FileGrid entries={sorted} onOpen={openEntry} onMenu={showMenu} />
    {:else}
      <FileList entries={sorted} onOpen={openEntry} onMenu={showMenu} />
    {/if}
  </div>
</section>

{#if menuOpen && menuTarget}
  {@const target = menuTarget}
  <ContextMenu
    {target}
    x={menuX}
    y={menuY}
    onOpen={() => { openEntry(target); hideMenu(); }}
    onDownload={() => { download(target); hideMenu(); }}
    onRename={() => { rename(target); hideMenu(); }}
    onCopy={() => { notifications.info("Copy not wired yet."); hideMenu(); }}
    onDelete={() => { remove(target); hideMenu(); }}
  />
{/if}
