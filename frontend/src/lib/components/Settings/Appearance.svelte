<script lang="ts">
  import Palette from "lucide-svelte/icons/palette";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import LayoutList from "lucide-svelte/icons/list";
  import { files } from "$lib/store/files.svelte";
  import type { ViewMode } from "$lib/types";

  const VIEW_OPTIONS: Array<{ value: ViewMode; label: string; icon: typeof LayoutGrid }> = [
    { value: "grid", label: "Grid", icon: LayoutGrid },
    { value: "list", label: "List", icon: LayoutList },
  ];
</script>

<section class="flex flex-col gap-3">
  <div class="flex items-center gap-2">
    <Palette size="18" class="text-fg-secondary" />
    <h2 class="sg-section-title">Appearance</h2>
  </div>

  <div class="sg-panel px-4 py-3 flex flex-wrap items-center justify-between gap-3">
    <div class="min-w-0">
      <h3 class="text-sm font-semibold text-fg-primary">File view</h3>
      <p class="text-xs text-fg-muted mt-1">Default layout for the file explorer.</p>
    </div>
    <div class="sg-seg" role="group" aria-label="File view">
      {#each VIEW_OPTIONS as option (option.value)}
        {@const Icon = option.icon}
        <button
          type="button"
          class="sg-seg-btn"
          onclick={() => files.setViewMode(option.value)}
          aria-pressed={files.viewMode === option.value}
        >
          <Icon size="16" />
          <span>{option.label}</span>
        </button>
      {/each}
    </div>
  </div>
</section>
