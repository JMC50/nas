<script lang="ts">
  import { iconFor, type FolderEntry } from "./icon-for";
  import { dragEntry } from "./drag-drop-action";

  interface Props {
    entries: FolderEntry[];
    dragPayload: (entry: FolderEntry) => string;
    onOpen: (entry: FolderEntry, opts?: { newTab?: boolean }) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
    onDropOnFolder: (event: DragEvent, target: FolderEntry) => void;
  }

  let { entries, dragPayload, onOpen, onMenu, onDropOnFolder }: Props = $props();

  let dropTargetName = $state<string | null>(null);

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
      oncontextmenu={(event) => onMenu(event, entry)}
      onkeydown={(event) => onKey(event, entry)}
      title={entry.name}
    >
      <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
      <span class="text-xs truncate w-full text-center">{entry.name}</span>
    </button>
  {/each}
</div>
