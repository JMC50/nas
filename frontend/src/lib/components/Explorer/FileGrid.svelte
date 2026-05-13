<script lang="ts">
  import { iconFor, type FolderEntry } from "./icon-for";
  import { hasPayload, NAS_ENTRY_MIME } from "./drag-drop";

  interface Props {
    entries: FolderEntry[];
    dragPayload: (entry: FolderEntry) => string;
    onOpen: (entry: FolderEntry, opts?: { newTab?: boolean }) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
    onDropOnFolder: (event: DragEvent, target: FolderEntry) => void;
  }

  let { entries, dragPayload, onOpen, onMenu, onDropOnFolder }: Props = $props();

  let dropTargetName = $state<string | null>(null);

  function onDragStart(event: DragEvent, entry: FolderEntry) {
    if (!event.dataTransfer) return;
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData(NAS_ENTRY_MIME, dragPayload(entry));
  }

  function onDragOver(event: DragEvent, entry: FolderEntry) {
    if (!entry.isFolder || !hasPayload(event)) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    dropTargetName = entry.name;
  }

  function onDragLeave(entry: FolderEntry) {
    if (dropTargetName === entry.name) dropTargetName = null;
  }

  function onDrop(event: DragEvent, entry: FolderEntry) {
    dropTargetName = null;
    onDropOnFolder(event, entry);
  }

  function onDblClick(event: MouseEvent, entry: FolderEntry) {
    onOpen(entry, { newTab: entry.isFolder && (event.ctrlKey || event.metaKey) });
  }

  function onAuxClick(event: MouseEvent, entry: FolderEntry) {
    if (event.button !== 1) return;
    if (!entry.isFolder) return;
    event.preventDefault();
    onOpen(entry, { newTab: true });
  }

  function onKey(event: KeyboardEvent, entry: FolderEntry) {
    if (event.key === "Enter") {
      event.preventDefault();
      onOpen(entry, { newTab: entry.isFolder && (event.ctrlKey || event.metaKey) });
    }
  }
</script>

<div class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-2 p-6">
  {#each entries as entry (entry.name)}
    {@const Icon = iconFor(entry)}
    <button
      type="button"
      class="group flex flex-col items-center gap-1.5 p-3 rounded-md text-fg-primary hover:bg-bg-hover transition-colors focus-visible:outline-2 focus-visible:outline-border-focus {entry.isFolder &&
      dropTargetName === entry.name
        ? 'ring-1 ring-accent bg-bg-elevated'
        : ''}"
      draggable="true"
      ondragstart={(event) => onDragStart(event, entry)}
      ondragover={(event) => onDragOver(event, entry)}
      ondragleave={() => onDragLeave(entry)}
      ondrop={(event) => onDrop(event, entry)}
      ondblclick={(event) => onDblClick(event, entry)}
      onauxclick={(event) => onAuxClick(event, entry)}
      oncontextmenu={(event) => onMenu(event, entry)}
      onkeydown={(event) => onKey(event, entry)}
      title={entry.name}
    >
      <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
      <span class="text-xs truncate w-full text-center">{entry.name}</span>
    </button>
  {/each}
</div>
