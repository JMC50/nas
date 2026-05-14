<script lang="ts">
  import { onMount, onDestroy, untrack } from "svelte";
  import Folder from "lucide-svelte/icons/folder";
  import X from "lucide-svelte/icons/x";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import { files } from "$lib/store/files.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
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
    performMoveBatch,
    locsEqual,
    type DragItem,
  } from "$lib/components/Explorer/drag-drop";
  import {
    toggleEntry,
    setSingle,
    selectRange,
  } from "$lib/components/Explorer/selection";
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
  let fileInputEl: HTMLInputElement;
  let folderInputEl: HTMLInputElement;

  // Multi-select state lives in tab payload (per-tab isolated). The last
  // element of `selection` is the range anchor for Shift+click.
  const selection = $derived(
    (tabs.list.find((t) => t.id === tabId)?.payload as ExplorerPayload | null)?.selection ?? [],
  );
  const selectedSet = $derived(new Set(selection));

  function updateSelection(next: string[]) {
    tabs.update(tabId, { payload: { loc, selection: next } as ExplorerPayload });
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

  // Refresh entries on folder change.
  // NB: do NOT write tabs payload here. Navigation handlers (navigateTo etc.)
  // are responsible for resetting per-tab state including selection — writing
  // to tabs.list inside this effect creates a reactive cycle because the
  // parent (TabContent) re-passes our `loc` prop on every tab update, which
  // would re-fire this effect ad infinitum.
  $effect(() => {
    void loc;
    refresh();
  });

  $effect(() => {
    const desired = loc.length === 0 ? "Files" : loc[loc.length - 1];
    const current = tabs.list.find((t) => t.id === tabId)?.title;
    if (current !== desired) untrack(() => tabs.update(tabId, { title: desired }));
  });

  // Inspector mirrors the current single selection (pure derivation — no
  // effect required, no risk of feedback loop). When exactly one entry is
  // selected, show its details; otherwise (0 or N>1) hide the Inspector.
  const selected = $derived<FolderEntry | null>(
    selection.length === 1
      ? entries.find((e) => e.name === selection[0]) ?? null
      : null,
  );

  function navigateTo(target: string[], opts: { newTab?: boolean } = {}) {
    if (opts.newTab) {
      tabs.openExplorer(target);
      return;
    }
    // Reset selection here (not in a $effect) so per-tab state is cleared
    // exactly once per navigation event without creating a reactive cycle.
    tabs.update(tabId, {
      payload: { loc: target, selection: [] } as ExplorerPayload,
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

  // ---------- Selection click handler ----------
  // Modifier precedence (highest → lowest):
  //   1. Shift (with existing anchor) → range select
  //   2. Ctrl/Cmd (Shift NOT held) → toggle
  //   3. No modifier → set single
  // Mobile selection mode: when `selection.length > 0` on mobile, plain tap
  // toggles inclusion (ctrl-style) — so users can build up a selection by tap.
  function onSelect(event: MouseEvent, entry: FolderEntry) {
    if (event.shiftKey && selection.length > 0) {
      const anchor = selection[selection.length - 1];
      updateSelection(selectRange(sorted.map((e) => e.name), anchor, entry.name));
    } else if (event.ctrlKey || event.metaKey) {
      updateSelection(toggleEntry(selection, entry.name));
    } else if (ui.isMobile && selection.length > 0) {
      updateSelection(toggleEntry(selection, entry.name));
    } else {
      updateSelection(setSingle(entry.name));
    }
  }

  // Long-press on mobile → enter selection mode with the long-pressed entry
  // as initial selection (only when no current selection — FileGrid/FileList
  // gates the call to avoid clobbering an existing selection).
  function onLongSelect(entry: FolderEntry) {
    updateSelection(setSingle(entry.name));
  }

  // ---------- Multi-entry delete ----------
  async function deleteSelected() {
    if (selection.length === 0) return;
    const total = selection.length;
    if (!confirm(`Delete ${total} item${total === 1 ? "" : "s"}?`)) return;
    const targets = sorted.filter((e) => selectedSet.has(e.name));
    // Use Promise.allSettled so a partial failure still surfaces a summary
    // (`Deleted N (M failed)`) instead of throwing away the success count.
    const results = await Promise.allSettled(
      targets.map((entry) => deleteEntry(loc, entry, { silent: true })),
    );
    const deleted = results.filter((r) => r.status === "fulfilled").length;
    const failed = total - deleted;
    if (deleted > 0) {
      const suffix = failed > 0 ? ` (${failed} failed)` : "";
      notifications.info(`Deleted ${deleted} item${deleted === 1 ? "" : "s"}${suffix}`);
    } else if (failed > 0) {
      notifications.error(`Delete failed for all ${failed} item${failed === 1 ? "" : "s"}`);
    }
    updateSelection([]);
    refresh();
  }

  // ---------- Keyboard shortcuts ----------
  const EDITABLE_SELECTOR = "input, textarea, [contenteditable='true']";

  function onKeyDown(event: KeyboardEvent) {
    if (tabs.activeId !== tabId) return;
    const target = event.target as HTMLElement | null;
    if (target?.matches?.(EDITABLE_SELECTOR)) return;
    if ((event.ctrlKey || event.metaKey) && event.key === "a") {
      event.preventDefault();
      updateSelection(sorted.map((e) => e.name));
    } else if (event.key === "Escape") {
      if (selection.length > 0) updateSelection([]);
    } else if (event.key === "Delete" && selection.length > 0) {
      event.preventDefault();
      deleteSelected();
    }
  }

  // ---------- Drag payload builder ----------
  // When the dragged entry is part of a multi-select, ship the whole set.
  // Otherwise ship just the dragged entry (still in batch shape for uniformity).
  function dragPayload(entry: FolderEntry): string {
    const items: DragItem[] =
      selectedSet.has(entry.name) && selection.length > 1
        ? sorted
            .filter((e) => selectedSet.has(e.name))
            .map((e) => ({ name: e.name, isFolder: e.isFolder }))
        : [{ name: entry.name, isFolder: entry.isFolder }];
    return buildPayload(loc, items);
  }

  async function handleMove(targetLoc: string[], payload: ReturnType<typeof readPayload>) {
    if (!payload) return;
    await performMoveBatch(payload.sourceLoc, payload.items, targetLoc);
    if (locsEqual(loc, payload.sourceLoc) || locsEqual(loc, targetLoc)) refresh();
    // Selection is invalidated post-move (names may have moved out of this loc).
    updateSelection([]);
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

  // ---------- Marquee (rectangle) selection ----------
  let containerEl: HTMLDivElement | null = $state(null);
  let marquee = $state<{ x0: number; y0: number; x1: number; y1: number } | null>(null);
  // Snapshot entry rects at marquee-start; mutated by auto-scroll deltas so
  // entries that scroll into view mid-drag still test against current screen
  // coords.
  let entryBoxes: Array<{ name: string; rect: DOMRect }> = [];
  let pointerId: number | null = null;
  const MARQUEE_DEAD_ZONE_PX = 4;
  const AUTOSCROLL_EDGE_PX = 32;
  const AUTOSCROLL_SPEED_PX = 8;

  function snapshot() {
    if (!containerEl) {
      entryBoxes = [];
      return;
    }
    entryBoxes = Array.from(
      containerEl.querySelectorAll<HTMLElement>("[data-entry-name]"),
    ).map((el) => ({
      name: el.dataset.entryName ?? "",
      rect: el.getBoundingClientRect(),
    }));
  }

  // Marquee is only allowed to start when:
  //   - primary button (button === 0)
  //   - target is inside the canvas (data-marquee-canvas="true") but NOT
  //     inside an entry (entries set data-marquee-canvas="false")
  //   - target is not a scrollbar (offsetX/Y past client size)
  //   - target is not an editable input
  // Pointer capture ensures we keep getting move/up even when the cursor
  // leaves the container.
  function onPointerDown(event: PointerEvent) {
    if (event.button !== 0) return;
    if (!containerEl) return;
    const target = event.target as HTMLElement | null;
    if (!target) return;
    if (target.matches(EDITABLE_SELECTOR)) return;
    // Closest canvas wins: if any ancestor (incl. self) is marked
    // canvas="false" before reaching "true", we're inside an entry — skip.
    const canvasAncestor = target.closest<HTMLElement>("[data-marquee-canvas]");
    if (!canvasAncestor || canvasAncestor.dataset.marqueeCanvas !== "true") return;
    // Scrollbar click rejection: offsetX/offsetY are relative to the target
    // (the canvas div), and clientWidth/Height exclude the scrollbar gutter.
    if (
      event.offsetX > containerEl.clientWidth ||
      event.offsetY > containerEl.clientHeight
    ) {
      return;
    }
    snapshot();
    marquee = {
      x0: event.clientX,
      y0: event.clientY,
      x1: event.clientX,
      y1: event.clientY,
    };
    pointerId = event.pointerId;
    containerEl.setPointerCapture(event.pointerId);
  }

  function intersected(): string[] {
    if (!marquee) return [];
    const bx = Math.min(marquee.x0, marquee.x1);
    const by = Math.min(marquee.y0, marquee.y1);
    const bw = Math.abs(marquee.x1 - marquee.x0);
    const bh = Math.abs(marquee.y1 - marquee.y0);
    return entryBoxes
      .filter(
        ({ rect }) =>
          rect.left < bx + bw &&
          rect.right > bx &&
          rect.top < by + bh &&
          rect.bottom > by,
      )
      .map((b) => b.name);
  }

  function shiftRects(dy: number) {
    // DOMRect's top/bottom/left/right are read-only. Recreate via DOMRect ctor.
    entryBoxes = entryBoxes.map(({ name, rect }) => ({
      name,
      rect: new DOMRect(rect.x, rect.y - dy, rect.width, rect.height),
    }));
  }

  function onPointerMove(event: PointerEvent) {
    if (!marquee || !containerEl) return;
    marquee = { ...marquee, x1: event.clientX, y1: event.clientY };
    // Auto-scroll when near vertical edges. The horizontal axis rarely needs
    // it (grid wraps), so we only handle Y for v1.
    const cRect = containerEl.getBoundingClientRect();
    let scrolled = 0;
    if (event.clientY < cRect.top + AUTOSCROLL_EDGE_PX) {
      const before = containerEl.scrollTop;
      containerEl.scrollTop = Math.max(0, before - AUTOSCROLL_SPEED_PX);
      scrolled = before - containerEl.scrollTop;
    } else if (event.clientY > cRect.bottom - AUTOSCROLL_EDGE_PX) {
      const before = containerEl.scrollTop;
      containerEl.scrollTop = before + AUTOSCROLL_SPEED_PX;
      scrolled = before - containerEl.scrollTop;
    }
    if (scrolled !== 0) shiftRects(scrolled);
    updateSelection(intersected());
  }

  function endMarquee(event: PointerEvent) {
    if (!marquee) return;
    const dx = Math.abs(marquee.x1 - marquee.x0);
    const dy = Math.abs(marquee.y1 - marquee.y0);
    if (dx < MARQUEE_DEAD_ZONE_PX && dy < MARQUEE_DEAD_ZONE_PX) {
      // Plain blank click — clear selection without registering a marquee.
      updateSelection([]);
    }
    // If marquee was real but nothing intersected, `updateSelection([])`
    // already fired during the last pointermove. Nothing else to do.
    marquee = null;
    if (pointerId !== null && containerEl) {
      try {
        containerEl.releasePointerCapture(pointerId);
      } catch {
        // Pointer capture may already be released by the browser.
      }
    }
    pointerId = null;
  }

  onMount(() => {
    window.addEventListener("click", hideMenu);
    window.addEventListener("keydown", onKeyDown);
  });
  onDestroy(() => {
    window.removeEventListener("click", hideMenu);
    window.removeEventListener("keydown", onKeyDown);
  });
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

  {#if ui.isMobile && selection.length > 0}
    <!--
      Mobile selection-mode bar. Appears when 1+ entries selected on mobile so
      users have a way to cancel/delete without a right-click menu.
    -->
    <div
      class="h-10 flex items-center gap-2 px-3 border-b border-border-default bg-bg-elevated"
    >
      <button
        type="button"
        class="w-7 h-7 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-primary hover:bg-bg-hover"
        onclick={() => updateSelection([])}
        aria-label="Cancel selection"
      >
        <X size="14" />
      </button>
      <span class="text-xs text-fg-primary">
        {selection.length} selected
      </span>
      <button
        type="button"
        class="ml-auto h-7 inline-flex items-center gap-1 px-2 rounded text-xs text-fg-danger hover:bg-bg-hover"
        onclick={deleteSelected}
        aria-label="Delete selected"
      >
        <Trash2 size="12" />
        Delete
      </button>
    </div>
  {/if}

  <input type="file" multiple bind:this={fileInputEl} onchange={onPick} class="hidden" />
  <input
    type="file"
    bind:this={folderInputEl}
    onchange={onPickFolder}
    class="hidden"
    webkitdirectory
  />

  <div
    bind:this={containerEl}
    class="flex-1 overflow-auto relative"
    data-marquee-canvas="true"
    onpointerdown={onPointerDown}
    onpointermove={onPointerMove}
    onpointerup={endMarquee}
    onpointercancel={endMarquee}
  >
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
        {onSelect}
        {onLongSelect}
        selectedNames={selection}
      />
    {:else}
      <FileList
        entries={sorted}
        {dragPayload}
        onOpen={openEntry}
        onMenu={showMenu}
        {onDropOnFolder}
        {onSelect}
        {onLongSelect}
        selectedNames={selection}
      />
    {/if}

    {#if marquee && containerEl}
      {@const cRect = containerEl.getBoundingClientRect()}
      {@const left = Math.min(marquee.x0, marquee.x1) - cRect.left + containerEl.scrollLeft}
      {@const top = Math.min(marquee.y0, marquee.y1) - cRect.top + containerEl.scrollTop}
      {@const width = Math.abs(marquee.x1 - marquee.x0)}
      {@const height = Math.abs(marquee.y1 - marquee.y0)}
      <div
        class="absolute pointer-events-none border border-accent bg-accent/10"
        style="left:{left}px;top:{top}px;width:{width}px;height:{height}px"
      ></div>
    {/if}
  </div>
</section>

<Inspector entry={selected} {loc} onClose={() => updateSelection([])} />

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
