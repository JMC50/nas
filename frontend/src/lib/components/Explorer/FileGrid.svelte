<script lang="ts">
  import { iconFor, type FolderEntry } from "./icon-for";
  import { dragEntry } from "./drag-drop-action";
  import { longPress } from "$lib/actions/long-press";
  import { ui } from "$lib/store/ui.svelte";
  import { formatBytes, formatRelTime } from "./format";

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

  function onKey(event: KeyboardEvent, entry: FolderEntry) {
    if (event.key === "Enter") {
      event.preventDefault();
      onOpen(entry, { newTab: entry.isFolder && (event.ctrlKey || event.metaKey) });
    }
  }

  function titleFor(entry: FolderEntry): string {
    if (entry.isFolder) return entry.name;
    return `${entry.name}\n${formatBytes(entry.size)} · ${formatRelTime(entry.modifiedAt)}`;
  }

  // Long-press unconditionally opens the context menu at the touch point.
  // Selection entry on mobile lives in the menu's "Select" item instead of a
  // competing gesture — see FileList.svelte and the v0.3.0 release notes for
  // the unified rationale.
  function handleLongPress(entry: FolderEntry, clientX: number, clientY: number) {
    const synthetic = new MouseEvent("contextmenu", {
      clientX,
      clientY,
      bubbles: true,
    });
    onMenu(synthetic, entry);
  }
</script>

<!--
  data-marquee-canvas="false" stops marquee pointerdown trigger from firing on
  entries — only blank-area mousedowns inside the canvas start a marquee.
-->
<div
  class="grid grid-cols-[repeat(auto-fill,minmax(112px,1fr))] gap-2 p-3 md:p-gutter"
  data-marquee-canvas="false"
>
  {#each entries as entry (entry.name)}
    {@const Icon = iconFor(entry)}
    {@const isSelected = selectedSet.has(entry.name)}
    {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
    <button
      type="button"
      data-entry-name={entry.name}
      class="tile {isSelected ? 'is-selected' : ''} {isDropTarget ? 'is-drop' : ''}"
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
      onkeydown={(event) => onKey(event, entry)}
      onclick={(event) => onSelect(event, entry)}
      title={titleFor(entry)}
      aria-pressed={isSelected}
    >
      <Icon size="32" class={entry.isFolder ? "text-fg-secondary" : "text-fg-muted"} />
      <span class="text-xs w-full text-center truncate">{entry.name}</span>
    </button>
  {/each}
</div>

<style>
  .tile {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    min-height: 100px;
    padding: 14px 8px 10px;
    border-radius: var(--radius-panel);
    color: var(--color-fg-primary);
    transition: background-color 70ms var(--ease-snap);
  }
  .tile:hover {
    background: var(--color-bg-hover);
  }
  .tile.is-selected {
    background: var(--color-bg-selected);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
  .tile.is-drop {
    background: var(--color-bg-selected);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
  @media (pointer: coarse) {
    .tile {
      -webkit-touch-callout: none;
      -webkit-user-select: none;
      user-select: none;
    }
  }
</style>
