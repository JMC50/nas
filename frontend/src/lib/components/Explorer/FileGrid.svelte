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
    // dispatches setSingle / toggleEntry / selectRange based on the event.
    onSelect: (event: MouseEvent, entry: FolderEntry) => void;
    // Mobile long-press → enter selection mode with this entry as initial pick.
    // When undefined or on desktop, FileGrid falls back to context menu.
    onLongSelect?: (entry: FolderEntry) => void;
    selectedNames: string[];
  }

  let {
    entries,
    dragPayload,
    onOpen,
    onMenu,
    onDropOnFolder,
    onSelect,
    onLongSelect,
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

  // Long-press on touch:
  //  - if onLongSelect is provided and there's no current selection, enter
  //    selection mode by selecting this entry (mobile multi-select entry path)
  //  - otherwise fall back to the context menu (desktop right-click parity)
  function handleLongPress(entry: FolderEntry, clientX: number, clientY: number) {
    if (onLongSelect && selectedSet.size === 0) {
      onLongSelect(entry);
      return;
    }
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
  class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-2 p-6"
  data-marquee-canvas="false"
>
  {#each entries as entry (entry.name)}
    {@const Icon = iconFor(entry)}
    {@const isSelected = selectedSet.has(entry.name)}
    {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
    <button
      type="button"
      data-entry-name={entry.name}
      class="group flex flex-col items-center gap-1.5 p-3 rounded-md text-fg-primary hover:bg-bg-hover transition-colors focus-visible:outline-2 focus-visible:outline-border-focus {isDropTarget
        ? 'ring-1 ring-accent bg-bg-elevated'
        : ''} {isSelected ? 'ring-1 ring-accent bg-bg-elevated' : ''}"
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
    >
      <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
      <span class="text-xs truncate w-full text-center">{entry.name}</span>
    </button>
  {/each}
</div>
