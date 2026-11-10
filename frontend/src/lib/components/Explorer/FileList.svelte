<script lang="ts">
  import MoreVertical from "lucide-svelte/icons/more-vertical";
  import { iconFor, type FolderEntry } from "./icon-for";
  import { dragEntry } from "./drag-drop-action";
  import { longPress } from "$lib/actions/long-press";
  import { ui } from "$lib/store/ui.svelte";
  import { formatBytes, formatRelTime, formatFullTime } from "./format";

  interface Props {
    entries: FolderEntry[];
    dragPayload: (entry: FolderEntry) => string;
    onOpen: (entry: FolderEntry, opts?: { newTab?: boolean }) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
    onDropOnFolder: (event: DragEvent, target: FolderEntry) => void;
    // Click → fires for every entry click (modifier-aware). Explorer.svelte
    // dispatches setSingle / toggleEntry / selectRange / openEntry based on the
    // event and viewport. On mobile with no current selection, the handler
    // opens the entry; selection mode is entered via the context menu's
    // "Select" item instead.
    onSelect: (event: MouseEvent, entry: FolderEntry) => void;
    selectedNames: string[];
  }

  let {
    entries,
    dragPayload,
    onOpen,
    onMenu,
    onDropOnFolder,
    onSelect,
    selectedNames,
  }: Props = $props();

  const selectedSet = $derived(new Set(selectedNames));
  let dropTargetName = $state<string | null>(null);

  // Long-press unconditionally opens the context menu at the touch point.
  // Mobile selection entry is exposed via the menu's "Select" item rather
  // than a competing gesture, so long-press is always menu and never a
  // hidden mode-switch (avoids the 500ms-delay confusion users hit when
  // the same gesture meant different things in different states).
  function handleLongPress(entry: FolderEntry, clientX: number, clientY: number) {
    const synthetic = new MouseEvent("contextmenu", {
      clientX,
      clientY,
      bubbles: true,
    });
    onMenu(synthetic, entry);
  }

  // Kebab tap on mobile — opens the same menu as long-press, anchored at
  // the kebab's screen position. stopPropagation prevents the row's onclick
  // (onSelect) from also firing.
  function onKebab(event: MouseEvent, entry: FolderEntry) {
    event.stopPropagation();
    onMenu(event, entry);
  }

  // Without this, pressing the kebab would also start the row's longPress
  // timer (action is on <tr>, kebab pointerdown bubbles up). After 500ms the
  // timer would fire onMenu, then pointerup → click → onKebab → onMenu again,
  // opening the menu twice. Stopping pointerdown at the kebab keeps the row's
  // longPress action from ever seeing the press.
  function stopRowLongPress(event: PointerEvent) {
    event.stopPropagation();
  }

  function onRowKey(event: KeyboardEvent, entry: FolderEntry) {
    if (event.target !== event.currentTarget) return;
    if (event.key === "Enter") {
      event.preventDefault();
      onOpen(entry, { newTab: entry.isFolder && (event.ctrlKey || event.metaKey) });
    } else if (event.key === " ") {
      event.preventDefault();
      onSelect(new MouseEvent("click", { ctrlKey: event.ctrlKey, metaKey: event.metaKey, shiftKey: event.shiftKey }), entry);
    } else if (event.key === "ContextMenu" || (event.shiftKey && event.key === "F10")) {
      event.preventDefault();
      const rect = (event.currentTarget as HTMLElement).getBoundingClientRect();
      onMenu(new MouseEvent("contextmenu", { clientX: rect.left + 24, clientY: rect.bottom }), entry);
    }
  }
</script>

<!--
  data-marquee-canvas="false" stops marquee pointerdown trigger from firing on
  table rows — only blank-area mousedowns inside the canvas start a marquee.
-->
<table class="w-full text-sm table-fixed" data-marquee-canvas="false">
  <thead>
    <tr class="text-xs text-fg-muted border-b border-border-default">
      <th scope="col" class="text-left font-semibold pl-gutter pr-3 h-8">Name</th>
      <th scope="col" class="hidden md:table-cell text-left font-semibold px-3 w-24">Type</th>
      <th scope="col" class="text-right font-semibold px-3 w-24">Size</th>
      <th scope="col" class="hidden sm:table-cell text-left font-semibold px-3 w-32 lg:pr-gutter">Modified</th>
      <th class="md:hidden w-12" aria-hidden="true"></th>
    </tr>
  </thead>
  <tbody>
    {#each entries as entry (entry.name)}
      {@const Icon = iconFor(entry)}
      {@const isSelected = selectedSet.has(entry.name)}
      {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
      <tr
        data-entry-name={entry.name}
        class="row {isSelected ? 'is-selected' : ''} {isDropTarget ? 'is-drop' : ''}"
        draggable={!ui.isMobile}
        use:dragEntry={{
          entry,
          dragPayload,
          onDropOnFolder,
          onOpen,
          onDropEnter: (name) => (dropTargetName = name),
          onDropLeave: (name) => {
            if (dropTargetName === name) dropTargetName = null;
          },
          onDropFinish: () => (dropTargetName = null),
        }}
        use:longPress={{
          onLongPress: (clientX, clientY) => handleLongPress(entry, clientX, clientY),
        }}
        oncontextmenu={(event) => onMenu(event, entry)}
        onclick={(event) => onSelect(event, entry)}
        onkeydown={(event) => onRowKey(event, entry)}
        tabindex="0"
        aria-selected={isSelected}
      >
        <td class="pl-gutter pr-3">
          <div class="flex items-center gap-3 min-w-0">
            <Icon size="16" class="shrink-0 {entry.isFolder ? 'text-fg-secondary' : 'text-fg-muted'}" />
            <span class="truncate text-fg-primary" title={entry.name}>{entry.name}</span>
          </div>
        </td>
        <td class="hidden md:table-cell px-3 text-xs text-fg-muted truncate">
          {entry.isFolder ? "Folder" : entry.extensions || "file"}
        </td>
        <td class="px-3 text-xs text-fg-muted text-right tabular-nums whitespace-nowrap">
          {entry.isFolder ? "—" : formatBytes(entry.size)}
        </td>
        <td class="hidden sm:table-cell px-3 lg:pr-gutter text-xs text-fg-muted whitespace-nowrap" title={formatFullTime(entry.modifiedAt)}>
          {formatRelTime(entry.modifiedAt)}
        </td>
        <!--
          Mobile-only kebab. Hidden at md+ where right-click / two-finger tap
          already invokes the context menu. data-marquee-canvas inherited as
          "false" from the parent <table>, so the kebab cannot start a marquee.
        -->
        <td class="md:hidden pr-1 text-right">
          <button
            type="button"
            class="sg-btn sg-btn-ghost sg-btn-icon"
            onpointerdown={stopRowLongPress}
            onclick={(event) => onKebab(event, entry)}
            aria-label="More actions for {entry.name}"
          >
            <MoreVertical size="18" />
          </button>
        </td>
      </tr>
    {/each}
  </tbody>
</table>

<style>
  .row {
    height: 36px;
    border-bottom: 1px solid var(--color-border-default);
    cursor: pointer;
    transition: background-color 70ms var(--ease-snap);
  }
  .row:hover {
    background: var(--color-bg-hover);
  }
  .row:focus-visible {
    outline-offset: -2px;
  }
  .row.is-selected {
    background: var(--color-bg-selected);
    box-shadow: inset 2px 0 0 var(--color-accent);
  }
  .row.is-drop {
    background: var(--color-bg-selected);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
  @media (pointer: coarse) {
    .row {
      height: 48px;
    }
  }
</style>
