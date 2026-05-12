<script lang="ts">
  import ChevronRight from "lucide-svelte/icons/chevron-right";
  import HomeIcon from "lucide-svelte/icons/home";
  import { files } from "$lib/store/files.svelte";

  interface Props {
    onGoto: (index: number) => void;
    onRoot: () => void;
  }

  let { onGoto, onRoot }: Props = $props();
</script>

<nav class="flex items-center gap-1 h-8 px-6 text-xs text-fg-muted border-b border-border-default bg-bg-base">
  <button
    type="button"
    class="inline-flex items-center gap-1 px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors"
    onclick={onRoot}
  >
    <HomeIcon size="12" />
    <span>root</span>
  </button>
  {#each files.currentLoc as segment, index (index)}
    <ChevronRight size="11" class="text-fg-disabled shrink-0" />
    <button
      type="button"
      class="px-1.5 h-6 rounded hover:bg-bg-hover hover:text-fg-primary transition-colors truncate"
      onclick={() => onGoto(index)}
    >
      {segment}
    </button>
  {/each}
</nav>
