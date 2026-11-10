<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import X from "lucide-svelte/icons/x";
  import LogOut from "lucide-svelte/icons/log-out";
  import UserIcon from "lucide-svelte/icons/user";
  import Logo from "$lib/components/Logo.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { NAV_ITEMS, activate, type NavItem } from "./nav-items";
  import { displayName, openAccount, logout } from "./session";

  const visibleItems = $derived(NAV_ITEMS.filter((item) => !item.adminOnly || auth.isAdmin));
  let panel: HTMLElement | undefined = $state();

  function pick(item: NavItem) {
    activate(item);
    ui.closeDrawer();
  }

  function pickAccount() {
    openAccount();
    ui.closeDrawer();
  }

  function focusables(): HTMLElement[] {
    return panel ? Array.from(panel.querySelectorAll<HTMLElement>("button:not(:disabled)")) : [];
  }

  function onKey(event: KeyboardEvent) {
    if (!ui.drawerOpen) return;
    if (event.key === "Escape") {
      event.preventDefault();
      ui.closeDrawer();
      return;
    }
    if (event.key !== "Tab") return;
    const items = focusables();
    if (items.length === 0) return;
    const first = items[0];
    const last = items[items.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  $effect(() => {
    if (!ui.drawerOpen) return;
    tick().then(() => {
      const current = panel?.querySelector<HTMLElement>('[aria-current="page"]');
      (current ?? focusables()[0])?.focus();
    });
  });

  onMount(() => window.addEventListener("keydown", onKey));
  onDestroy(() => window.removeEventListener("keydown", onKey));
</script>

{#if ui.drawerOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div class="fixed inset-0 z-50 lg:hidden">
    <div class="absolute inset-0 bg-black/55 animate-fade" onclick={() => ui.closeDrawer()}></div>
    <div
      bind:this={panel}
      class="absolute top-0 left-0 h-full w-[min(300px,86vw)] flex flex-col bg-bg-base border-r border-border-default shadow-overlay animate-drawer"
      role="dialog"
      aria-modal="true"
      aria-label="Main navigation"
    >
      <div class="h-12 flex items-center gap-2 pl-4 pr-2 border-b border-border-default bg-bg-surface">
        <Logo size={22} />
        <span class="text-md font-bold text-fg-primary">NAS</span>
        <button type="button" class="ml-auto sg-btn sg-btn-ghost sg-btn-icon" onclick={() => ui.closeDrawer()} aria-label="Close navigation menu">
          <X size="18" />
        </button>
      </div>

      <nav class="flex-1 overflow-y-auto flex flex-col gap-1 p-2" aria-label="Main">
        {#each visibleItems as item (item.id)}
          {@const Icon = item.icon}
          {@const isActive = tabs.activeId === item.id}
          <button
            type="button"
            class="drawer-item {isActive ? 'is-current' : ''}"
            onclick={() => pick(item)}
            aria-label={item.title}
            aria-current={isActive ? "page" : undefined}
          >
            <Icon size="20" />
            <span class="truncate">{item.title}</span>
          </button>
        {/each}
      </nav>

      {#if auth.isAuthenticated}
        <div class="flex flex-col gap-1 p-2 border-t border-border-default">
          <button type="button" class="drawer-item" onclick={pickAccount}>
            <UserIcon size="20" />
            <span class="truncate">{displayName()}</span>
          </button>
          <button type="button" class="drawer-item" onclick={logout}>
            <LogOut size="20" />
            <span>Log out</span>
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .drawer-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: 12px;
    min-height: 44px;
    padding: 0 12px;
    border-radius: var(--radius-control);
    color: var(--color-fg-secondary);
    font-size: var(--text-sm);
    text-align: left;
  }
  .drawer-item:hover {
    background: var(--color-bg-hover);
    color: var(--color-fg-primary);
  }
  .drawer-item:focus-visible {
    outline-offset: -2px;
  }
  .drawer-item.is-current {
    background: var(--color-bg-selected);
    color: var(--color-fg-primary);
    font-weight: 650;
    box-shadow: inset 2px 0 0 var(--color-accent);
  }
  .drawer-item.is-current :global(svg) {
    color: var(--color-accent);
  }
</style>
