<script lang="ts">
  import Search from "lucide-svelte/icons/search";
  import LogOut from "lucide-svelte/icons/log-out";
  import Logo from "$lib/components/Logo.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { NAV_ITEMS, activate } from "./nav-items";
  import { displayName, openAccount, logout } from "./session";

  const visibleItems = $derived(NAV_ITEMS.filter((item) => !item.adminOnly || auth.isAdmin));
  const name = $derived(auth.isAuthenticated ? displayName() : "Guest");
  const accountActive = $derived(tabs.activeId === "system:account");
</script>

<aside
  class="hidden lg:flex row-span-2 col-start-1 row-start-1 flex-col items-center gap-1 py-3 w-16 border-r border-border-default bg-bg-surface"
>
  <div class="grid place-items-center size-11 mb-2" aria-hidden="true">
    <Logo size={26} />
  </div>

  <nav class="flex flex-col items-center gap-1" aria-label="Main">
    {#each visibleItems as item (item.id)}
      {@const Icon = item.icon}
      {@const isActive = tabs.activeId === item.id}
      <button
        type="button"
        class="rail-item {isActive ? 'is-current' : ''}"
        onclick={() => activate(item)}
        aria-label={item.title}
        aria-current={isActive ? "page" : undefined}
      >
        <Icon size="20" />
        <span class="sg-tip left-[calc(100%+10px)] top-1/2 -translate-y-1/2" aria-hidden="true">{item.title}</span>
      </button>
    {/each}
  </nav>

  <div class="flex-1"></div>

  <button type="button" class="rail-item" onclick={() => ui.openQuickOpen()} aria-label="Quick open (Ctrl+P)">
    <Search size="20" />
    <span class="sg-tip left-[calc(100%+10px)] top-1/2 -translate-y-1/2" aria-hidden="true">Search · Ctrl P</span>
  </button>
  {#if auth.isAuthenticated}
    <button
      type="button"
      class="rail-item {accountActive ? 'is-current' : ''}"
      onclick={openAccount}
      aria-label="Open account"
      aria-current={accountActive ? "page" : undefined}
    >
      <span class="grid place-items-center size-7 rounded-control bg-fg-info text-accent-fg text-xs font-bold" aria-hidden="true">
        {name.slice(0, 1).toUpperCase()}
      </span>
      <span class="sg-tip left-[calc(100%+10px)] top-1/2 -translate-y-1/2" aria-hidden="true">{name}</span>
    </button>
    <button type="button" class="rail-item" onclick={logout} aria-label="Log out">
      <LogOut size="20" />
      <span class="sg-tip left-[calc(100%+10px)] top-1/2 -translate-y-1/2" aria-hidden="true">Log out</span>
    </button>
  {/if}
</aside>

<style>
  .rail-item {
    position: relative;
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border-radius: var(--radius-control);
    color: var(--color-fg-secondary);
    transition: background-color 70ms var(--ease-snap), color 70ms var(--ease-snap);
  }
  .rail-item:hover {
    background: var(--color-bg-hover);
    color: var(--color-fg-primary);
  }
  .rail-item.is-current {
    background: var(--color-bg-selected);
    color: var(--color-accent);
  }
  .rail-item.is-current::before {
    content: "";
    position: absolute;
    left: -10px;
    top: 10px;
    bottom: 10px;
    width: 2px;
    background: var(--color-accent);
  }
</style>
