<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import MoreHorizontal from "lucide-svelte/icons/more-horizontal";
  import FolderUp from "lucide-svelte/icons/folder-up";
  import FolderPlus from "lucide-svelte/icons/folder-plus";
  import Search from "lucide-svelte/icons/search";
  import { clickOutside } from "$lib/actions/click-outside";

  interface Props {
    onUploadFolder: () => void;
    onNewFolder: () => void;
    onToggleSearch: () => void;
  }

  let { onUploadFolder, onNewFolder, onToggleSearch }: Props = $props();

  let open = $state(false);

  function pick(action: () => void) {
    action();
    open = false;
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "Escape" && open) {
      event.preventDefault();
      open = false;
    }
  }

  onMount(() => window.addEventListener("keydown", onKey));
  onDestroy(() => window.removeEventListener("keydown", onKey));
</script>

<div class="relative" use:clickOutside={{ onOutside: () => (open = false), enabled: open }}>
  <button
    type="button"
    class="inline-flex items-center justify-center w-8 h-8 rounded-md text-fg-muted hover:text-fg-primary hover:bg-bg-hover transition-colors"
    onclick={(event) => {
      event.stopPropagation();
      open = !open;
    }}
    aria-label="More actions"
    aria-expanded={open}
  >
    <MoreHorizontal size="16" />
  </button>
  {#if open}
    <div
      class="absolute right-0 top-full mt-1 py-1 min-w-[180px] rounded-md bg-bg-elevated border border-border-default shadow-lg z-30"
      role="menu"
    >
      <button
        type="button"
        class="w-full flex items-center gap-2 px-3 h-9 text-left text-xs text-fg-primary hover:bg-bg-hover transition-colors"
        onclick={() => pick(onUploadFolder)}
        role="menuitem"
      >
        <FolderUp size="14" class="text-fg-muted shrink-0" />
        <span>Upload folder</span>
      </button>
      <button
        type="button"
        class="w-full flex items-center gap-2 px-3 h-9 text-left text-xs text-fg-primary hover:bg-bg-hover transition-colors"
        onclick={() => pick(onNewFolder)}
        role="menuitem"
      >
        <FolderPlus size="14" class="text-fg-muted shrink-0" />
        <span>New folder</span>
      </button>
      <button
        type="button"
        class="w-full flex items-center gap-2 px-3 h-9 text-left text-xs text-fg-primary hover:bg-bg-hover transition-colors"
        onclick={() => pick(onToggleSearch)}
        role="menuitem"
      >
        <Search size="14" class="text-fg-muted shrink-0" />
        <span>Search</span>
      </button>
    </div>
  {/if}
</div>
