<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { NAV_ITEMS, activate, type NavItem } from "./nav-items";

  const visibleItems = $derived(NAV_ITEMS.filter((item) => !item.adminOnly || auth.isAdmin));

  function pick(item: NavItem) {
    activate(item);
    ui.closeDrawer();
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "Escape" && ui.drawerOpen) {
      event.preventDefault();
      ui.closeDrawer();
    }
  }

  onMount(() => window.addEventListener("keydown", onKey));
  onDestroy(() => window.removeEventListener("keydown", onKey));
</script>

{#if ui.drawerOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 md:hidden">
    <div
      class="absolute inset-0 bg-bg-base/60 backdrop-blur-sm"
      onclick={() => ui.closeDrawer()}
    ></div>
    <aside
      class="absolute top-0 left-0 w-64 h-full bg-bg-surface border-r border-border-default flex flex-col pt-[env(safe-area-inset-top)] shadow-[0_0_24px_rgba(0,0,0,0.5)]"
      aria-label="Main navigation"
    >
      <nav class="flex-1 flex flex-col gap-1 p-2">
        {#each visibleItems as item (item.id)}
          {@const Icon = item.icon}
          {@const isActive = tabs.activeId === item.id}
          <button
            type="button"
            class="flex items-center gap-3 h-10 px-3 rounded-md text-sm transition-colors {isActive
              ? 'bg-accent text-accent-fg'
              : 'text-fg-secondary hover:text-fg-primary hover:bg-bg-hover'}"
            onclick={() => pick(item)}
            aria-label={item.title}
            aria-current={isActive ? "page" : undefined}
          >
            <Icon size="18" class="shrink-0" />
            <span class="truncate">{item.title}</span>
          </button>
        {/each}
      </nav>
    </aside>
  </div>
{/if}
