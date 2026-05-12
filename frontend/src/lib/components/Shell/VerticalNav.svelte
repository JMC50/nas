<script lang="ts">
  import Folder from "lucide-svelte/icons/folder";
  import Users from "lucide-svelte/icons/users";
  import History from "lucide-svelte/icons/history";
  import Settings from "lucide-svelte/icons/settings";
  import Cpu from "lucide-svelte/icons/cpu";
  import PanelLeft from "lucide-svelte/icons/panel-left";
  import { tabs, EXPLORER_TAB_ID } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import type { TabKind } from "$lib/types";

  interface NavItem {
    id: string;
    kind: TabKind;
    title: string;
    icon: typeof Folder;
  }

  const NAV_ITEMS: NavItem[] = [
    { id: EXPLORER_TAB_ID, kind: "explorer", title: "Files", icon: Folder },
    { id: "system:user-manager", kind: "user-manager", title: "Users", icon: Users },
    { id: "system:activity", kind: "activity", title: "Activity", icon: History },
    { id: "system:settings", kind: "settings", title: "Settings", icon: Settings },
    { id: "system:system", kind: "system", title: "System", icon: Cpu },
  ];

  function activate(item: NavItem) {
    const existing = tabs.list.find((tab) => tab.id === item.id);
    if (existing) {
      tabs.setActive(item.id);
      return;
    }
    tabs.open({
      id: item.id,
      kind: item.kind,
      title: item.title,
      icon: item.kind,
      payload: null,
      closable: item.kind !== "explorer",
    });
  }
</script>

<aside
  class="row-start-2 col-start-1 flex flex-col border-r border-border-default bg-bg-surface transition-[width] duration-200 ease-smooth {ui.sidebarCollapsed ? 'w-12' : 'w-[200px]'}"
>
  <nav class="flex-1 flex flex-col gap-1 p-1.5">
    {#each NAV_ITEMS as item (item.id)}
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
