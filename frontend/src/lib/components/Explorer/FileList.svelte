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
</script>

<table class="w-full text-sm">
  <thead>
    <tr class="text-xs text-fg-muted border-b border-border-default">
      <th class="text-left font-normal px-6 py-2">Name</th>
      <th class="text-left font-normal px-6 py-2 w-24">Type</th>
    </tr>
  </thead>
  <tbody>
    {#each entries as entry (entry.name)}
      {@const Icon = iconFor(entry)}
      <tr
        class="border-b border-border-default/40 hover:bg-bg-hover cursor-pointer {entry.isFolder &&
        dropTargetName === entry.name
          ? 'ring-1 ring-inset ring-accent bg-bg-elevated'
          : ''}"
        draggable="true"
        ondragstart={(event) => onDragStart(event, entry)}
        ondragover={(event) => onDragOver(event, entry)}
        ondragleave={() => onDragLeave(entry)}
        ondrop={(event) => onDrop(event, entry)}
        ondblclick={(event) => onDblClick(event, entry)}
        onauxclick={(event) => onAuxClick(event, entry)}
        oncontextmenu={(event) => onMenu(event, entry)}
      >
        <td class="px-6 py-1.5">
          <div class="flex items-center gap-2 min-w-0">
            <Icon size="14" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
            <span class="truncate text-fg-primary">{entry.name}</span>
          </div>
        </td>
        <td class="px-6 py-1.5 text-xs text-fg-muted">
          {entry.isFolder ? "Folder" : entry.extensions || "file"}
        </td>
      </tr>
    {/each}
  </tbody>
</table>
