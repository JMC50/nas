<script lang="ts">
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import HomeIcon from "lucide-svelte/icons/home";
  import { hasPayload } from "./drag-drop";

  interface Props {
    loc: string[];
    onGoto: (index: number) => void;
    onRoot: () => void;
    onDropToRoot: (event: DragEvent) => void;
    onDropToSegment: (event: DragEvent, index: number) => void;
  }

  let { loc, onGoto, onRoot, onDropToRoot, onDropToSegment }: Props = $props();

  let dropTarget = $state<number | null>(null); // -1 = root, n = segment index

  function rootDragOver(event: DragEvent) {
    if (!hasPayload(event) || loc.length === 0) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    dropTarget = -1;
  }

  function segmentDragOver(event: DragEvent, index: number) {
    if (!hasPayload(event)) return;
    if (index === loc.length - 1) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    dropTarget = index;
  }

  function dragLeave() {
    dropTarget = null;
  }

  function rootDrop(event: DragEvent) {
    dropTarget = null;
    if (loc.length === 0) return;
    onDropToRoot(event);
  }

  function segmentDrop(event: DragEvent, index: number) {
    dropTarget = null;
    if (index === loc.length - 1) return;
    onDropToSegment(event, index);
  }
</script>

<nav
  class="flex items-center gap-1 h-8 px-6 text-xs text-fg-muted border-b border-border-default bg-bg-base"
>
  <button
    type="button"
    class="inline-flex items-center gap-1 px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors {dropTarget ===
    -1
      ? 'bg-bg-elevated text-fg-accent ring-1 ring-accent'
      : ''}"
    onclick={onRoot}
    ondragover={rootDragOver}
    ondragleave={dragLeave}
    ondrop={rootDrop}
  >
    <HomeIcon size="12" />
    <span>root</span>
  </button>
  {#each loc as segment, index (index)}
    <ChevronRight size="11" class="text-fg-disabled shrink-0" />
    {#if index === loc.length - 1}
      <span
        class="inline-flex items-center px-1.5 h-6 rounded bg-bg-elevated text-fg-accent font-medium truncate"
        aria-current="page"
      >
        {segment}
      </span>
    {:else}
      <button
        type="button"
        class="px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors truncate {dropTarget ===
        index
          ? 'bg-bg-elevated text-fg-accent ring-1 ring-accent'
          : ''}"
        onclick={() => onGoto(index)}
        ondragover={(event) => segmentDragOver(event, index)}
        ondragleave={dragLeave}
        ondrop={(event) => segmentDrop(event, index)}
      >
        {segment}
      </button>
    {/if}
  {/each}
</nav>
