<script lang="ts">
  import { mode, toggleMode } from "mode-watcher";
  import Sun from "lucide-svelte/icons/sun";
  import Moon from "lucide-svelte/icons/moon";
</script>

<button
  type="button"
  class="inline-flex items-center justify-center w-8 h-8 rounded-md text-fg-primary hover:bg-bg-hover transition-colors"
  onclick={() => toggleMode()}
  aria-label="Toggle theme"
>
  {#if mode.current === "dark"}
    <Sun size="16" />
  {:else}
    <Moon size="16" />
  {/if}
</button>
