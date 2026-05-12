<script lang="ts">
  import { iconFor, type FolderEntry } from "./icon-for";

  interface Props {
    entries: FolderEntry[];
    onOpen: (entry: FolderEntry) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
  }

  let { entries, onOpen, onMenu }: Props = $props();

  function onKey(event: KeyboardEvent, entry: FolderEntry) {
    if (event.key === "Enter") {
      event.preventDefault();
      onOpen(entry);
    }
  }
</script>

<div class="grid grid-cols-[repeat(auto-fill,minmax(120px,1fr))] gap-2 p-6">
  {#each entries as entry (entry.name)}
    {@const Icon = iconFor(entry)}
    <button
      type="button"
      class="group flex flex-col items-center gap-1.5 p-3 rounded-md text-fg-primary hover:bg-bg-hover transition-colors focus-visible:outline-2 focus-visible:outline-border-focus"
      ondblclick={() => onOpen(entry)}
      oncontextmenu={(event) => onMenu(event, entry)}
      onkeydown={(event) => onKey(event, entry)}
      title={entry.name}
    >
      <Icon size="32" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
      <span class="text-xs truncate w-full text-center">{entry.name}</span>
    </button>
  {/each}
</div>
