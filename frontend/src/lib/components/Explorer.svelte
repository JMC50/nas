<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Folder from "lucide-svelte/icons/folder";
  import { files } from "$lib/store/files.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { pickViewer } from "$lib/components/Viewers/registry";
  import Toolbar from "$lib/components/Explorer/Toolbar.svelte";
  import Breadcrumb from "$lib/components/Explorer/Breadcrumb.svelte";
  import FileGrid from "$lib/components/Explorer/FileGrid.svelte";
  import FileList from "$lib/components/Explorer/FileList.svelte";
  import Inspector from "$lib/components/Explorer/Inspector.svelte";
  import ContextMenu from "$lib/components/Explorer/ContextMenu.svelte";
  import ForbiddenPanel from "$lib/components/ForbiddenPanel.svelte";
  import {
    locPath,
    readEntries,
    createFolder,
    deleteEntry,
    renameEntry,
    downloadUrl,
    FetchError,
  } from "$lib/components/Explorer/actions";
  import {
    buildPayload,
    readPayload,
    performMove,
    locsEqual,
  } from "$lib/components/Explorer/drag-drop";
  import { enqueueFiles, enqueueFolder } from "$lib/components/Explorer/uploads";
  import { useBackButton } from "$lib/components/Explorer/back-button";
  import type { FolderEntry } from "$lib/components/Explorer/icon-for";
  import type { ExplorerPayload } from "$lib/types";

  interface Props {
    loc: string[];
    tabId: string;
  }
  let { loc, tabId }: Props = $props();

  let entries: FolderEntry[] = $state([]);
  let loading = $state(false);
  let forbidden = $state(false);
  let errorMessage: string | null = $state(null);
  let searchQuery = $state("");
  let selected = $state<FolderEntry | null>(null);
  let fileInputEl: HTMLInputElement;
  let folderInputEl: HTMLInputElement;

  const selectedName = $derived(selected?.name ?? null);

  function toggleSelect(entry: FolderEntry) {
    selected = selected?.name === entry.name ? null : entry;
  }

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

  const folderName = $derived(loc.length === 0 ? "Home" : loc[loc.length - 1]);

  async function refresh() {
    loading = true;
    errorMessage = null;
    forbidden = false;
    try {
      entries = await readEntries(loc);
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
    void loc;
    selected = null;
    refresh();
  });

  $effect(() => {
    const desired = loc.length === 0 ? "Files" : loc[loc.length - 1];
    const current = tabs.list.find((t) => t.id === tabId)?.title;
    if (current !== desired) tabs.update(tabId, { title: desired });
  });

  function navigateTo(target: string[], opts: { newTab?: boolean } = {}) {
    if (opts.newTab) {
      tabs.openExplorer(target);
      return;
    }
    tabs.update(tabId, {
      payload: { loc: target } as ExplorerPayload,
      title: target.length === 0 ? "Files" : target[target.length - 1],
    });
  }

  function navigateUp() {
    if (loc.length === 0) return;
    navigateTo(loc.slice(0, -1));
  }

  function download(entry: FolderEntry) {
    if (entry.isFolder) {
      notifications.info("Folder download not implemented yet.");
      return;
    }
    const anchor = document.createElement("a");
    anchor.href = downloadUrl(loc, entry);
    anchor.download = entry.name;
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  }

  function openEntry(entry: FolderEntry, opts: { newTab?: boolean } = {}) {
    if (entry.isFolder) {
      navigateTo([...loc, entry.name], opts);
      return;
    }
    const kind = pickViewer(entry.extensions);
    if (kind === null) {
      notifications.info(`No preview for .${entry.extensions} — downloading instead.`);
      download(entry);
      return;
    }
    tabs.open({
      kind,
      title: entry.name,
      icon: kind,
      payload: { loc: locPath(loc), name: entry.name },
      closable: true,
    });
  }

  function onPick(event: Event) {
    enqueueFiles(loc, event.target as HTMLInputElement);
  }

  function onPickFolder(event: Event) {
    enqueueFolder(loc, event.target as HTMLInputElement);
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
    runMutation(() => createFolder(loc, name), "Create failed");
  }

  function remove(entry: FolderEntry) {
    if (!confirm(`Delete ${entry.name}?`)) return;
    runMutation(() => deleteEntry(loc, entry), "Delete failed");
  }

  function rename(entry: FolderEntry) {
    const next = prompt(`Rename "${entry.name}" to:`, entry.name);
    if (!next || next === entry.name) return;
    runMutation(() => renameEntry(loc, entry, next), "Rename failed");
  }

  function dragPayload(entry: FolderEntry): string {
    return buildPayload(loc, entry);
  }

  async function handleMove(targetLoc: string[], payload: ReturnType<typeof readPayload>) {
    if (!payload) return;
    const moved = await performMove(payload.sourceLoc, payload.name, payload.isFolder, targetLoc);
    if (moved && (locsEqual(loc, payload.sourceLoc) || locsEqual(loc, targetLoc))) refresh();
  }

  function onDropOnFolder(event: DragEvent, target: FolderEntry) {
    if (!target.isFolder) return;
    const payload = readPayload(event);
    if (!payload) return;
    event.preventDefault();
    handleMove([...loc, target.name], payload);
  }

  function onDropOnLoc(event: DragEvent, targetLoc: string[]) {
    const payload = readPayload(event);
    if (!payload) return;
    event.preventDefault();
    handleMove(targetLoc, payload);
  }

  // ---------- Context menu ----------
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

  useBackButton(tabId, navigateUp);

  onMount(() => window.addEventListener("click", hideMenu));
  onDestroy(() => window.removeEventListener("click", hideMenu));
</script>

<section class="flex flex-col h-full bg-bg-base">
  <Toolbar
    {loc}
    {searchQuery}
    onSearch={(value) => (searchQuery = value)}
    onUp={navigateUp}
    onUpload={() => fileInputEl?.click()}
    onUploadFolder={() => folderInputEl?.click()}
    onNewFolder={newFolder}
    onDropToUp={(event) => onDropOnLoc(event, loc.slice(0, -1))}
  />

  <Breadcrumb
    {loc}
    onGoto={(index) => navigateTo(loc.slice(0, index + 1))}
    onRoot={() => navigateTo([])}
    onDropToRoot={(event) => onDropOnLoc(event, [])}
    onDropToSegment={(event, index) => onDropOnLoc(event, loc.slice(0, index + 1))}
  />

  <div class="h-8 flex items-center gap-2 px-6 border-b border-border-default bg-bg-base">
    <Folder size="14" class="text-accent shrink-0" />
    <span class="text-xs font-medium text-fg-primary truncate">{folderName}</span>
    <span class="ml-auto text-xs text-fg-muted font-mono shrink-0">
      {sorted.length} {sorted.length === 1 ? "item" : "items"}
    </span>
  </div>

  <input type="file" multiple bind:this={fileInputEl} onchange={onPick} class="hidden" />
  <input
    type="file"
    bind:this={folderInputEl}
    onchange={onPickFolder}
    class="hidden"
    webkitdirectory
  />

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
    {:else if files.viewMode === "grid"}
      <FileGrid
        entries={sorted}
        {dragPayload}
        onOpen={openEntry}
        onMenu={showMenu}
        {onDropOnFolder}
        onSelect={toggleSelect}
        {selectedName}
      />
    {:else}
      <FileList
        entries={sorted}
        {dragPayload}
        onOpen={openEntry}
        onMenu={showMenu}
        {onDropOnFolder}
        onSelect={toggleSelect}
        {selectedName}
      />
    {/if}
  </div>
</section>

<Inspector entry={selected} {loc} onClose={() => (selected = null)} />

{#if menuOpen && menuTarget}
  {@const target = menuTarget}
  <ContextMenu
    {target}
    x={menuX}
    y={menuY}
    onOpen={() => {
      openEntry(target);
      hideMenu();
    }}
    onOpenNewTab={() => {
      openEntry(target, { newTab: true });
      hideMenu();
    }}
    onDownload={() => {
      download(target);
      hideMenu();
    }}
    onRename={() => {
      rename(target);
      hideMenu();
    }}
    onCopy={() => {
      notifications.info("Copy not wired yet.");
      hideMenu();
    }}
    onDelete={() => {
      remove(target);
      hideMenu();
    }}
  />
{/if}
