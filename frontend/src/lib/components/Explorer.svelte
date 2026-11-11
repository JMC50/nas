<script lang="ts">
  import { onMount, onDestroy, untrack } from "svelte";
  import Folder from "lucide-svelte/icons/folder";
  import X from "lucide-svelte/icons/x";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import DownloadIcon from "lucide-svelte/icons/download";
  import UploadIcon from "lucide-svelte/icons/upload";
  import FolderUp from "lucide-svelte/icons/folder-up";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import Search from "lucide-svelte/icons/search";
  import FolderOpen from "lucide-svelte/icons/folder-open";
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
  import ShareLinkDialog from "$lib/components/Share/ShareLinkDialog.svelte";
  import ForbiddenPanel from "$lib/components/ForbiddenPanel.svelte";
  import {
    locPath,
    readEntries,
    createFolder,
    deleteEntry,
    renameEntry,
    downloadUrl,
    prepareZip,
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
  // Monotonic token used to discard stale `readEntries` responses. Without
  // this, rapid navigation (mount-fetch + immediate user click) could land
  // the OLDER response last, overwriting the newer location's entries.
  let refreshToken = 0;

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
    const myToken = ++refreshToken;
    loading = true;
    errorMessage = null;
    forbidden = false;
    try {
      const next = await readEntries(loc);
      if (myToken !== refreshToken) return; // stale — newer refresh in flight
      entries = next;
    } catch (cause) {
      if (myToken !== refreshToken) return;
      if (cause instanceof FetchError && cause.status === 403) {
        forbidden = true;
      } else {
        errorMessage = (cause as Error).message;
      }
      entries = [];
    } finally {
      if (myToken === refreshToken) loading = false;
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

  // Inspector entry. Decoupled from selection so single-click only highlights
  // the row (selection ring) — the Inspector opens ONLY via the context-menu
  // "Details" action (or programmatically via inspect()). User decision: a
  // selection-driven Inspector felt too pushy / kept covering the file area
  // on accidental clicks.
  let inspectedEntry = $state<FolderEntry | null>(null);

  function inspect(entry: FolderEntry) {
    inspectedEntry = entry;
  }

  function closeInspector() {
    inspectedEntry = null;
  }

  // 대화 상자가 열린 동안 폴더를 옮겨도 연 때의 위치로 링크를 만들도록 loc을 복사해 둠
  let share = $state<{ entry: FolderEntry; loc: string[] } | null>(null);
  let shareOpener: HTMLElement | null = null;

  function openShare(entry: FolderEntry) {
    shareOpener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    share = { entry, loc: [...loc] };
  }

  // 메뉴에서 열었으면 메뉴 항목이 사라졌으므로 그 파일의 행으로 초점을 돌려줌
  function closeShare() {
    const name = share?.entry.name ?? "";
    share = null;
    const row = containerEl?.querySelector<HTMLElement>(`[data-entry-name="${CSS.escape(name)}"]`);
    (shareOpener?.isConnected ? shareOpener : row)?.focus();
    shareOpener = null;
  }

  // If the inspected entry disappears (e.g. deleted, renamed, or out of the
  // filtered set after refresh), drop the Inspector. $derived rather than an
  // effect so there's no write-back loop into entries.
  const inspectedExists = $derived(
    inspectedEntry === null
      ? true
      : entries.some((e) => e.name === inspectedEntry?.name),
  );
  $effect(() => {
    if (!inspectedExists && inspectedEntry !== null) {
      untrack(() => (inspectedEntry = null));
    }
  });

  function navigateTo(target: string[], opts: { newTab?: boolean } = {}) {
    if (opts.newTab) {
      tabs.openExplorer(target);
      return;
    }
    // Folder change invalidates Inspector contents.
    inspectedEntry = null;
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

  function saveFrom(href: string, name: string) {
    const anchor = document.createElement("a");
    anchor.href = href;
    anchor.download = name;
    anchor.rel = "noopener";
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
  }

  let zipPending = $state(false);

  // 파일 하나는 그대로 받고, 폴더나 여러 항목은 서버가 zip 하나로 묶어 보냄
  async function download(targets: FolderEntry[]) {
    if (targets.length === 1 && !targets[0].isFolder) {
      saveFrom(downloadUrl(loc, targets[0]), targets[0].name);
      return;
    }
    if (targets.length === 0 || zipPending) return;
    zipPending = true;
    try {
      const zip = await prepareZip(loc, targets);
      saveFrom(zip.url, zip.name);
      notifications.info(`Downloading ${zip.name}`);
    } catch (cause) {
      notifications.error(cause instanceof FetchError ? cause.message : "Could not reach the server.");
    } finally {
      zipPending = false;
    }
  }

  const selectedEntries = $derived(sorted.filter((entry) => selectedSet.has(entry.name)));

  // 메뉴를 연 항목이 여러 개 고른 것 중 하나면 고른 것 전체를 받음
  function downloadTargets(entry: FolderEntry): FolderEntry[] {
    return selectedEntries.length > 1 && selectedSet.has(entry.name) ? selectedEntries : [entry];
  }

  function downloadLabel(targets: FolderEntry[]): string {
    if (targets.length > 1) return `Download ${targets.length} items as .zip`;
    if (targets.length === 1 && targets[0].isFolder) return "Download as .zip";
    return "Download";
  }

  function openEntry(entry: FolderEntry, opts: { newTab?: boolean } = {}) {
    if (entry.isFolder) {
      navigateTo([...loc, entry.name], opts);
      return;
    }
    const kind = pickViewer(entry.extensions);
    if (kind === null) {
      notifications.info(`No preview for .${entry.extensions} — downloading instead.`);
      download([entry]);
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
  //   3. Touch + active selection → toggle (build up selection by tap)
  //   4. Touch + no selection → OPEN the entry (single-tap is the primary
  //      touch gesture; selection is entered explicitly via the context
  //      menu's "Select" item — see ContextMenu.svelte)
  //   5. Desktop, no modifier → set single (selection ring without opening)
  function onSelect(event: MouseEvent, entry: FolderEntry) {
    if (event.shiftKey && selection.length > 0) {
      const anchor = selection[selection.length - 1];
      updateSelection(selectRange(sorted.map((e) => e.name), anchor, entry.name));
    } else if (event.ctrlKey || event.metaKey) {
      updateSelection(toggleEntry(selection, entry.name));
    } else if (ui.touchLayout && selection.length > 0) {
      updateSelection(toggleEntry(selection, entry.name));
    } else if (ui.touchLayout) {
      openEntry(entry);
    } else {
      updateSelection(setSingle(entry.name));
    }
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
    if (tabs.activeId !== tabId || share) return;
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
    bgMenuOpen = false;
  }

  function hideMenu() {
    menuOpen = false;
    menuTarget = null;
  }

  // ---------- Empty-area (background) context menu ----------
  // Fires when the user right-clicks the canvas blank space (no entry under
  // pointer). Offers folder-level actions: new folder, upload, upload folder,
  // refresh — the same operations otherwise reachable only via the toolbar.
  let bgMenuOpen = $state(false);
  let bgMenuX = $state(0);
  let bgMenuY = $state(0);
  let bgMenuEl: HTMLDivElement | undefined = $state();
  let bgMenuSize = $state({ width: 0, height: 0 });

  $effect(() => {
    if (bgMenuEl) bgMenuSize = { width: bgMenuEl.offsetWidth, height: bgMenuEl.offsetHeight };
  });

  function showBackgroundMenu(event: MouseEvent) {
    // Only fire when the contextmenu target is the canvas itself (not an
    // entry, which sets data-marquee-canvas="false" on its outer wrapper).
    const target = event.target as HTMLElement | null;
    if (!target) return;
    if (target.closest('[data-marquee-canvas="false"]')) return;
    const canvasAncestor = target.closest<HTMLElement>("[data-marquee-canvas]");
    if (!canvasAncestor || canvasAncestor.dataset.marqueeCanvas !== "true") return;
    if (target.matches?.(EDITABLE_SELECTOR)) return;
    event.preventDefault();
    bgMenuX = event.clientX;
    bgMenuY = event.clientY;
    bgMenuOpen = true;
    menuOpen = false;
  }

  function hideBackgroundMenu() {
    bgMenuOpen = false;
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

  function hideAnyMenu() {
    hideMenu();
    hideBackgroundMenu();
  }

  onMount(() => {
    window.addEventListener("click", hideAnyMenu);
    window.addEventListener("keydown", onKeyDown);
  });
  onDestroy(() => {
    window.removeEventListener("click", hideAnyMenu);
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
    onDownload={() => download(selectedEntries)}
    downloadCount={selectedEntries.length}
    downloadLabel={downloadLabel(selectedEntries)}
    downloadBusy={zipPending}
    onDropToUp={(event) => onDropOnLoc(event, loc.slice(0, -1))}
  />

  <Breadcrumb
    {loc}
    onGoto={(index) => navigateTo(loc.slice(0, index + 1))}
    onRoot={() => navigateTo([])}
    onDropToRoot={(event) => onDropOnLoc(event, [])}
    onDropToSegment={(event, index) => onDropOnLoc(event, loc.slice(0, index + 1))}
  />

  <div class="h-10 flex items-center gap-2 px-gutter border-b border-border-default bg-bg-base">
    <Folder size="16" class="text-fg-secondary shrink-0" />
    <h2 class="text-sm font-semibold text-fg-primary truncate min-w-0">{folderName}</h2>
    <span class="ml-auto shrink-0 text-xs text-fg-muted tabular-nums">
      {#if selectedEntries.length > 0 && !ui.touchLayout}{selectedEntries.length} selected ·{/if}
      {sorted.length} {sorted.length === 1 ? "item" : "items"}
    </span>
  </div>

  {#if ui.touchLayout && selection.length > 0}
    <!--
      Mobile selection-mode bar. Appears when 1+ entries selected on mobile so
      users have a way to cancel/download/delete without a right-click menu.
    -->
    <div
      class="flex items-center gap-2 px-2 py-1 border-b border-border-default bg-bg-selected shadow-[inset_2px_0_0_var(--color-accent)]"
    >
      <button
        type="button"
        class="sg-btn sg-btn-ghost sg-btn-icon"
        onclick={() => updateSelection([])}
        aria-label="Cancel selection"
      >
        <X size="18" />
      </button>
      <span class="text-sm font-semibold text-fg-primary tabular-nums">
        {selection.length} selected
      </span>
      <button
        type="button"
        class="ml-auto sg-btn sg-btn-ghost sg-btn-icon"
        onclick={() => download(selectedEntries)}
        disabled={zipPending}
        aria-label={downloadLabel(selectedEntries)}
        title={downloadLabel(selectedEntries)}
      >
        <DownloadIcon size="18" />
      </button>
      <button
        type="button"
        class="sg-btn sg-btn-danger"
        onclick={deleteSelected}
        aria-label="Delete selected"
      >
        <Trash2 size="16" />
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

  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    bind:this={containerEl}
    class="flex-1 overflow-auto relative"
    data-marquee-canvas="true"
    onpointerdown={onPointerDown}
    onpointermove={onPointerMove}
    onpointerup={endMarquee}
    onpointercancel={endMarquee}
    oncontextmenu={showBackgroundMenu}
  >
    {#if forbidden}
      <ForbiddenPanel
        title="Files are locked"
        description="You don't have permission to browse files yet. Ask an administrator to grant you the View intent, or claim admin yourself with the admin password."
        onGranted={refresh}
      />
    {:else if errorMessage}
      <div class="m-gutter p-3 rounded-panel border border-fg-danger/45 text-sm text-fg-danger" role="alert">Failed to load: {errorMessage}</div>
    {:else if loading && entries.length === 0}
      <div class="px-gutter py-6 text-sm text-fg-muted" role="status">Loading…</div>
    {:else if sorted.length === 0}
      <div class="flex flex-col items-center gap-2 px-gutter py-16 text-center">
        {#if searchQuery}
          <Search size="28" class="text-fg-muted" />
          <p class="text-sm font-semibold text-fg-primary">No matches in this folder.</p>
          <p class="text-xs text-fg-muted">Try a shorter name, or search all files with Ctrl P.</p>
        {:else}
          <FolderOpen size="28" class="text-fg-muted" />
          <p class="text-sm font-semibold text-fg-primary">Folder is empty.</p>
          <p class="text-xs text-fg-muted">Drop files here, or use Upload.</p>
        {/if}
      </div>
    {:else if files.viewMode === "grid"}
      <FileGrid
        entries={sorted}
        {dragPayload}
        onOpen={openEntry}
        onMenu={showMenu}
        {onDropOnFolder}
        {onSelect}
        selectedNames={selection}
      />
    {:else}
      <FileList
        entries={sorted}
        {dragPayload}
        onOpen={openEntry}
        onMenu={showMenu}
        onShare={openShare}
        {onDropOnFolder}
        {onSelect}
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
        class="absolute pointer-events-none rounded-chip border border-accent bg-accent/10"
        style="left:{left}px;top:{top}px;width:{width}px;height:{height}px"
      ></div>
    {/if}
  </div>
</section>

<Inspector entry={inspectedEntry} {loc} onClose={closeInspector} />

{#if share}
  <ShareLinkDialog target={share.entry} loc={share.loc} onClose={closeShare} />
{/if}

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
    downloadLabel={downloadLabel(downloadTargets(target))}
    onDownload={() => {
      download(downloadTargets(target));
      hideMenu();
    }}
    onShare={() => {
      openShare(target);
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
    onInspect={() => {
      inspect(target);
      hideMenu();
    }}
    onSelect={() => {
      // Enters mobile selection mode by setting this entry as the single
      // member. Subsequent taps toggle (see onSelect handler above), so the
      // user builds up the selection from here.
      updateSelection(setSingle(target.name));
      hideMenu();
    }}
    onClose={hideMenu}
  />
{/if}

{#if bgMenuOpen}
  <!-- Empty-area context menu: appears when the user right-clicks the
       file canvas anywhere except on an entry. Folder-level actions only. -->
  <div
    class="fixed z-50 sg-menu animate-pop origin-top-left"
    bind:this={bgMenuEl}
    style="left: {Math.max(8, Math.min(bgMenuX, ui.viewportWidth - bgMenuSize.width - 8))}px; top: {Math.max(8, Math.min(bgMenuY, ui.viewportHeight - bgMenuSize.height - 8))}px;"
    role="menu"
  >
    <button
      type="button"
      class="sg-menu-item"
      role="menuitem"
      onclick={() => { newFolder(); hideBackgroundMenu(); }}
    >
      <Folder size="16" />
      New folder
    </button>
    <button
      type="button"
      class="sg-menu-item"
      role="menuitem"
      onclick={() => { fileInputEl?.click(); hideBackgroundMenu(); }}
    >
      <UploadIcon size="16" />
      Upload files
    </button>
    <button
      type="button"
      class="sg-menu-item"
      role="menuitem"
      onclick={() => { folderInputEl?.click(); hideBackgroundMenu(); }}
    >
      <FolderUp size="16" />
      Upload folder
    </button>
    <div class="sg-menu-sep" role="separator"></div>
    <button
      type="button"
      class="sg-menu-item"
      role="menuitem"
      onclick={() => { refresh(); hideBackgroundMenu(); }}
    >
      <RefreshCw size="16" />
      Refresh
    </button>
  </div>
{/if}
