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
  let navEl: HTMLElement | undefined = $state();

  $effect(() => {
    void loc;
    if (navEl) navEl.scrollLeft = navEl.scrollWidth;
  });

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
  bind:this={navEl}
  class="flex items-center gap-0.5 min-h-9 px-2 md:px-gutter overflow-x-auto text-sm text-fg-secondary border-b border-border-default bg-bg-base [scrollbar-width:none]"
  aria-label="Folder path"
>
  <button
    type="button"
    class="crumb {dropTarget === -1 ? 'is-drop' : ''}"
    onclick={onRoot}
    ondragover={rootDragOver}
    ondragleave={dragLeave}
    ondrop={rootDrop}
  >
    <HomeIcon size="14" />
    <span>root</span>
  </button>
  {#each loc as segment, index (index)}
    <ChevronRight size="14" class="text-fg-muted shrink-0" />
    {#if index === loc.length - 1}
      <span class="crumb is-current" aria-current="page" title={segment}>{segment}</span>
    {:else}
      <button
        type="button"
        class="crumb {dropTarget === index ? 'is-drop' : ''}"
        onclick={() => onGoto(index)}
        ondragover={(event) => segmentDragOver(event, index)}
        ondragleave={dragLeave}
        ondrop={(event) => segmentDrop(event, index)}
        title={segment}
      >
        <span class="truncate">{segment}</span>
      </button>
    {/if}
  {/each}
</nav>

<style>
  .crumb {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    gap: 6px;
    max-width: 240px;
    min-height: 28px;
    padding: 0 6px;
    border-radius: var(--radius-chip);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: background-color 70ms var(--ease-snap), color 70ms var(--ease-snap);
  }
  button.crumb:hover {
    background: var(--color-bg-hover);
    color: var(--color-fg-primary);
  }
  .crumb:focus-visible {
    outline-offset: -2px;
  }
  .crumb.is-current {
    display: block;
    line-height: 28px;
    color: var(--color-fg-primary);
    font-weight: 650;
  }
  .crumb.is-drop {
    background: var(--color-bg-selected);
    color: var(--color-accent);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
  @media (pointer: coarse) {
    .crumb {
      min-height: 44px;
    }
    .crumb.is-current {
      line-height: 44px;
    }
  }
</style>
