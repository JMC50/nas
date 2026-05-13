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
