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
    onSelect: (entry: FolderEntry) => void;
    selectedName: string | null;
  }

  let {
    entries,
    dragPayload,
    onOpen,
    onMenu,
    onDropOnFolder,
    onSelect,
    selectedName,
  }: Props = $props();

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

  // Long-press on touch fires onMenu with a synthetic MouseEvent so the
  // existing ContextMenu (anchored at clientX/clientY) opens under the finger.
  function openMenuAt(entry: FolderEntry, clientX: number, clientY: number) {
    const synthetic = new MouseEvent("contextmenu", {
      clientX,
      clientY,
      bubbles: true,
    });
    onMenu(synthetic, entry);
  }
</script>

<div class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-2 p-6">
  {#each entries as entry (entry.name)}
    {@const Icon = iconFor(entry)}
    {@const isSelected = entry.name === selectedName}
    {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
    <button
      type="button"
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
        onLongPress: (clientX, clientY) => openMenuAt(entry, clientX, clientY),
      }}
      oncontextmenu={(event) => onMenu(event, entry)}
      onkeydown={(event) => onKey(event, entry)}
      onclick={() => onSelect(entry)}
      title={titleFor(entry)}
    >
      <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
      <span class="text-xs truncate w-full text-center">{entry.name}</span>
    </button>
  {/each}
</div>
