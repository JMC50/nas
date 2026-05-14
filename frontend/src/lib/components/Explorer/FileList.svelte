<script lang="ts">
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

<table class="w-full text-sm">
  <thead>
    <tr class="text-xs text-fg-muted border-b border-border-default">
      <th class="text-left font-normal px-6 py-2">Name</th>
      <th class="text-left font-normal px-6 py-2 w-24">Type</th>
      <th class="text-right font-normal px-6 py-2 w-20 font-mono">Size</th>
      <th class="text-left font-normal px-6 py-2 w-32">Modified</th>
    </tr>
  </thead>
  <tbody>
    {#each entries as entry (entry.name)}
      {@const Icon = iconFor(entry)}
      {@const isSelected = entry.name === selectedName}
      {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
      <tr
        class="border-b border-border-default/40 hover:bg-bg-hover cursor-pointer {isDropTarget
          ? 'ring-1 ring-inset ring-accent bg-bg-elevated'
          : ''} {isSelected ? 'ring-1 ring-inset ring-accent bg-bg-elevated' : ''}"
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
        onclick={() => onSelect(entry)}
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
        <td class="px-6 py-1.5 text-xs text-fg-muted font-mono text-right">
          {entry.isFolder ? "—" : formatBytes(entry.size)}
        </td>
        <td
          class="px-6 py-1.5 text-xs text-fg-muted"
          title={formatFullTime(entry.modifiedAt)}
        >
          {formatRelTime(entry.modifiedAt)}
        </td>
      </tr>
    {/each}
  </tbody>
</table>
