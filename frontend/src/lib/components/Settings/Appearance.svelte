<script lang="ts">
  import Palette from "lucide-svelte/icons/palette";
  import LayoutGrid from "lucide-svelte/icons/layout-grid";
  import LayoutList from "lucide-svelte/icons/list";
  import Sun from "lucide-svelte/icons/sun";
  import Moon from "lucide-svelte/icons/moon";
  import Monitor from "lucide-svelte/icons/monitor";
  import { setMode, userPrefersMode } from "mode-watcher";
  import { files } from "$lib/store/files.svelte";
  import type { ViewMode } from "$lib/types";

  type ModeValue = "system" | "light" | "dark";

  const MODE_OPTIONS: Array<{ value: ModeValue; label: string; icon: typeof Sun }> = [
    { value: "system", label: "System", icon: Monitor },
    { value: "light", label: "Light", icon: Sun },
    { value: "dark", label: "Dark", icon: Moon },
  ];

  const VIEW_OPTIONS: Array<{ value: ViewMode; label: string; icon: typeof LayoutGrid }> = [
    { value: "grid", label: "Grid", icon: LayoutGrid },
    { value: "list", label: "List", icon: LayoutList },
  ];
</script>

<section class="space-y-3">
  <div class="flex items-center gap-2">
    <Palette size="14" class="text-fg-muted" />
    <h2 class="text-xs font-semibold uppercase tracking-wide text-fg-muted">Appearance</h2>
  </div>

  <div class="rounded-lg bg-bg-surface border border-border-default divide-y divide-border-default/60">
    <div class="px-4 py-3">
      <div class="flex items-center justify-between gap-4">
        <div>
          <div class="text-sm text-fg-primary">Theme</div>
          <div class="text-xs text-fg-muted mt-0.5">Choose between dark, light, or follow your OS preference.</div>
        </div>
        <div class="flex items-center p-0.5 rounded-md bg-bg-elevated">
          {#each MODE_OPTIONS as option (option.value)}
            {@const Icon = option.icon}
            {@const active = userPrefersMode.current === option.value}
            <button
              type="button"
              class="inline-flex items-center gap-1.5 h-7 px-2.5 rounded text-xs transition-colors {active ? 'bg-accent text-accent-fg' : 'text-fg-muted hover:text-fg-primary'}"
              onclick={() => setMode(option.value)}
              aria-pressed={active}
            >
              <Icon size="12" />
              <span>{option.label}</span>
            </button>
          {/each}
        </div>
      </div>
    </div>

    <div class="px-4 py-3">
      <div class="flex items-center justify-between gap-4">
        <div>
          <div class="text-sm text-fg-primary">File view</div>
          <div class="text-xs text-fg-muted mt-0.5">Default layout for the file explorer.</div>
        </div>
        <div class="flex items-center p-0.5 rounded-md bg-bg-elevated">
          {#each VIEW_OPTIONS as option (option.value)}
            {@const Icon = option.icon}
            {@const active = files.viewMode === option.value}
            <button
              type="button"
              class="inline-flex items-center gap-1.5 h-7 px-2.5 rounded text-xs transition-colors {active ? 'bg-accent text-accent-fg' : 'text-fg-muted hover:text-fg-primary'}"
              onclick={() => files.setViewMode(option.value)}
              aria-pressed={active}
            >
              <Icon size="12" />
              <span>{option.label}</span>
            </button>
          {/each}
        </div>
      </div>
    </div>
  </div>
</section>
