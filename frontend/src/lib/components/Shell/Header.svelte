<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import Menu from "lucide-svelte/icons/menu";
  import Search from "lucide-svelte/icons/search";
  import LogOut from "lucide-svelte/icons/log-out";
  import { auth } from "$lib/store/auth.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { displayName, openAccount, logout } from "./session";

  const name = $derived(displayName());
</script>

<header
  class="row-start-1 col-start-1 lg:hidden h-12 flex items-center gap-2 px-2 md:px-4 border-b border-border-default bg-bg-surface"
>
  <button
    type="button"
    class="sg-btn sg-btn-ghost sg-btn-icon"
    onclick={(event) => ui.openDrawer(event.currentTarget)}
    aria-label="Open navigation menu"
    aria-expanded={ui.drawerOpen}
  >
    <Menu size="20" />
  </button>
  <div class="flex items-center gap-2 min-w-0">
    <Logo size={22} />
    <span class="text-md font-bold text-fg-primary">NAS</span>
  </div>

  <button
    type="button"
    class="hidden md:flex flex-1 max-w-md mx-auto items-center gap-2 h-8 pointer-coarse:h-11 px-2.5 rounded-control border border-border-default bg-bg-inset text-fg-muted hover:border-border-strong transition-colors"
    onclick={() => ui.openQuickOpen()}
    aria-label="Quick open (Ctrl+P)"
  >
    <Search size="16" />
    <span class="flex-1 text-left">Search files…</span>
    <kbd class="px-1.5 rounded-chip border border-border-default text-xs text-fg-secondary">Ctrl P</kbd>
  </button>

  <div class="ml-auto md:ml-0 flex items-center gap-1">
    <button
      type="button"
      class="md:hidden sg-btn sg-btn-ghost sg-btn-icon"
      onclick={() => ui.openQuickOpen()}
      aria-label="Quick open"
    >
      <Search size="18" />
    </button>
    {#if auth.isAuthenticated}
      <button type="button" class="sg-btn sg-btn-ghost gap-2 px-2.5" onclick={openAccount} aria-label="Open account">
        <span class="grid place-items-center size-6 rounded-control bg-fg-info text-accent-fg text-xs font-bold" aria-hidden="true">
          {name.slice(0, 1).toUpperCase()}
        </span>
        <span class="hidden md:inline max-w-[140px] truncate">{name}</span>
      </button>
      <button type="button" class="sg-btn sg-btn-ghost sg-btn-icon" onclick={logout} aria-label="Log out">
        <LogOut size="18" />
      </button>
    {/if}
  </div>
</header>
