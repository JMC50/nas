<script lang="ts">
  import HardDrive from "lucide-svelte/icons/hard-drive";
  import Search from "lucide-svelte/icons/search";
  import LogOut from "lucide-svelte/icons/log-out";
  import UserIcon from "lucide-svelte/icons/user";
  import ThemeToggle from "$lib/components/ThemeToggle.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { tabs } from "$lib/store/tabs.svelte";

  function openQuickFind() {
    ui.openQuickOpen();
  }

  function openAccount() {
    tabs.open({
      id: "system:account",
      kind: "account",
      title: "Account",
      icon: "account",
      payload: null,
      closable: true,
    });
  }

  function logout() {
    auth.clear();
    if (typeof window !== "undefined") {
      window.location.href = "/localLogin";
    }
  }

  const displayName = $derived(
    auth.current.krname || auth.current.global_name || auth.current.username || "Guest",
  );
</script>

<header
  class="row-start-1 col-span-2 h-12 flex items-center justify-between gap-4 px-4 border-b border-border-default bg-bg-surface"
>
  <div class="flex items-center gap-2 min-w-0">
    <HardDrive size="20" class="text-accent shrink-0" />
    <span class="text-fg-primary font-semibold text-sm tracking-tight">NAS</span>
  </div>

  <button
    type="button"
    class="flex-1 max-w-md flex items-center gap-2 h-8 px-3 rounded-md bg-bg-elevated text-fg-muted hover:text-fg-primary hover:bg-bg-hover transition-colors"
    onclick={openQuickFind}
    aria-label="Quick open (Ctrl+P)"
  >
    <Search size="14" />
    <span class="text-xs">Search files… (Ctrl+P)</span>
  </button>

  <div class="flex items-center gap-2">
    <ThemeToggle />
    {#if auth.isAuthenticated}
      <button
        type="button"
        class="flex items-center gap-2 px-2 h-8 rounded-md bg-bg-elevated hover:bg-bg-hover transition-colors"
        onclick={openAccount}
        aria-label="Open account"
      >
        <UserIcon size="14" class="text-fg-muted" />
        <span class="text-xs text-fg-secondary truncate max-w-[140px]">
          {displayName}
        </span>
      </button>
      <button
        type="button"
        class="inline-flex items-center justify-center w-8 h-8 rounded-md text-fg-muted hover:text-fg-danger hover:bg-bg-hover transition-colors"
        onclick={logout}
        aria-label="Log out"
      >
        <LogOut size="16" />
      </button>
    {/if}
  </div>
</header>
