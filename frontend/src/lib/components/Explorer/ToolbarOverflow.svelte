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
    class="sg-btn sg-btn-ghost sg-btn-icon"
    onclick={(event) => {
      event.stopPropagation();
      open = !open;
    }}
    aria-label="More actions"
    aria-expanded={open}
    aria-haspopup="menu"
  >
    <MoreHorizontal size="18" />
  </button>
  {#if open}
    <div class="absolute right-0 top-full mt-1 z-30 sg-menu animate-pop origin-top-right" role="menu">
      <button type="button" class="sg-menu-item" onclick={() => pick(onUploadFolder)} role="menuitem">
        <FolderUp size="16" />
        <span>Upload folder</span>
      </button>
      <button type="button" class="sg-menu-item" onclick={() => pick(onNewFolder)} role="menuitem">
        <FolderPlus size="16" />
        <span>New folder</span>
      </button>
      <button type="button" class="sg-menu-item" onclick={() => pick(onToggleSearch)} role="menuitem">
        <Search size="16" />
        <span>Search</span>
      </button>
    </div>
  {/if}
</div>
