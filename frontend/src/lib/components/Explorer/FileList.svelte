<script lang="ts">
  import { iconFor, type FolderEntry } from "./icon-for";

  interface Props {
    entries: FolderEntry[];
    onOpen: (entry: FolderEntry) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
  }

  let { entries, onOpen, onMenu }: Props = $props();
</script>

<table class="w-full text-sm">
  <thead>
    <tr class="text-xs text-fg-muted border-b border-border-default">
      <th class="text-left font-normal px-4 py-2">Name</th>
      <th class="text-left font-normal px-4 py-2 w-24">Type</th>
    </tr>
  </thead>
  <tbody>
    {#each entries as entry (entry.name)}
      {@const Icon = iconFor(entry)}
      <tr
        class="border-b border-border-default/40 hover:bg-bg-hover cursor-pointer"
        ondblclick={() => onOpen(entry)}
        oncontextmenu={(event) => onMenu(event, entry)}
      >
        <td class="px-4 py-1.5">
          <div class="flex items-center gap-2 min-w-0">
            <Icon size="14" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
            <span class="truncate text-fg-primary">{entry.name}</span>
          </div>
        </td>
        <td class="px-4 py-1.5 text-xs text-fg-muted">
          {entry.isFolder ? "Folder" : entry.extensions || "file"}
        </td>
      </tr>
    {/each}
  </tbody>
</table>
