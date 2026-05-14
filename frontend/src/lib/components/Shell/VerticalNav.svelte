<script lang="ts">
  import PanelLeft from "lucide-svelte/icons/panel-left";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { NAV_ITEMS, activate } from "./nav-items";

  const visibleItems = $derived(NAV_ITEMS.filter((item) => !item.adminOnly || auth.isAdmin));
</script>

<aside
  class="row-start-2 col-start-1 flex flex-col border-r border-border-default bg-bg-surface transition-[width] duration-200 ease-smooth {ui.sidebarCollapsed ? 'w-12' : 'w-[200px]'}"
>
  <nav class="flex-1 flex flex-col gap-1 p-1.5">
    {#each visibleItems as item (item.id)}
      {@const Icon = item.icon}
      {@const isActive = tabs.activeId === item.id}
      <button
        type="button"
        class="flex items-center gap-3 h-9 px-2 rounded-md text-sm transition-colors {isActive
          ? 'bg-accent text-accent-fg'
          : 'text-fg-secondary hover:text-fg-primary hover:bg-bg-hover'}"
        onclick={() => activate(item)}
        aria-label={item.title}
        aria-current={isActive ? "page" : undefined}
      >
        <Icon size="16" class="shrink-0" />
        {#if !ui.sidebarCollapsed}
          <span class="truncate">{item.title}</span>
        {/if}
      </button>
    {/each}
  </nav>

  <button
    type="button"
    class="flex items-center gap-3 h-9 px-2 m-1.5 rounded-md text-sm text-fg-muted hover:text-fg-primary hover:bg-bg-hover transition-colors"
    onclick={() => ui.toggleSidebar()}
    aria-label={ui.sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
  >
    <PanelLeft size="16" class="shrink-0" />
    {#if !ui.sidebarCollapsed}
      <span class="truncate">Collapse</span>
    {/if}
  </button>
</aside>
