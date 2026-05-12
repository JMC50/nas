<script lang="ts">
  import { onMount } from "svelte";
  import Settings from "lucide-svelte/icons/settings";
  import UserIcon from "lucide-svelte/icons/user";
  import Search from "lucide-svelte/icons/search";
  import { auth } from "$lib/store/auth.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import type { Intent } from "$lib/types";

  interface UserView {
    userId: string;
    username: string;
    global_name: string;
    krname: string;
    intents: Intent[];
  }

  let isAdmin = $state(false);
  let users: UserView[] = $state([]);
  let loading = $state(true);
  let filterText = $state("");

  const filteredUsers = $derived(
    filterText
      ? users.filter((user) => {
          const value = filterText.toLowerCase();
          return (
            user.krname.toLowerCase().includes(value) ||
            user.username.toLowerCase().includes(value) ||
            user.global_name.toLowerCase().includes(value) ||
            user.userId.toLowerCase().includes(value)
          );
        })
      : users,
  );

  async function load() {
    loading = true;
    try {
      if (!auth.token) {
        isAdmin = false;
        return;
      }
      const adminResponse = await fetch(`/server/checkAdmin?token=${encodeURIComponent(auth.token)}`);
      const adminData = await adminResponse.json();
      isAdmin = Boolean(adminData.isAdmin);
      if (!isAdmin) return;
      const usersResponse = await fetch("/server/getAllUsers");
      const usersData = await usersResponse.json();
      users = (usersData.users ?? []) as UserView[];
    } catch (cause) {
      notifications.error(`Failed to load: ${(cause as Error).message}`);
    } finally {
      loading = false;
    }
  }

  function openUser(user: UserView) {
    tabs.open({
      id: `user-manager:${user.userId}`,
      kind: "user-manager",
      title: user.krname || user.global_name || user.username,
      icon: "user-manager",
      payload: { userId: user.userId },
      closable: true,
    });
  }

  onMount(load);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <Settings size="18" class="text-accent" />
    <h1 class="text-sm font-semibold text-fg-primary">Settings</h1>
  </header>

  <div class="flex-1 overflow-auto p-6">
    {#if loading}
      <div class="text-sm text-fg-muted">Loading…</div>
    {:else if !isAdmin}
      <div class="max-w-md p-5 rounded-lg bg-bg-surface border border-border-default">
        <h2 class="text-sm font-semibold text-fg-primary mb-1">Admin only</h2>
        <p class="text-xs text-fg-muted">
          You need the ADMIN permission to manage users. Request it from the Account page.
        </p>
      </div>
    {:else}
      <div class="max-w-3xl">
        <div class="flex items-center gap-3 mb-4">
          <div class="flex items-center gap-1.5 px-2.5 h-8 rounded-md bg-bg-elevated flex-1 max-w-sm">
            <Search size="13" class="text-fg-muted shrink-0" />
            <input
              type="text"
              bind:value={filterText}
              placeholder="Filter users…"
              class="flex-1 bg-transparent text-xs text-fg-primary placeholder:text-fg-muted outline-none"
            />
          </div>
          <span class="text-xs text-fg-muted font-mono">{filteredUsers.length} / {users.length}</span>
        </div>

        <div class="rounded-lg bg-bg-surface border border-border-default divide-y divide-border-default/60">
          {#each filteredUsers as user (user.userId)}
            <button
              type="button"
              class="w-full flex items-center gap-3 px-4 py-3 text-left hover:bg-bg-hover transition-colors"
              onclick={() => openUser(user)}
            >
              <div class="w-9 h-9 rounded-full bg-bg-elevated border border-border-default flex items-center justify-center shrink-0">
                <UserIcon size="14" class="text-fg-muted" />
              </div>
              <div class="min-w-0 flex-1">
                <div class="text-sm text-fg-primary truncate">
                  {user.krname || user.global_name || user.username}
                </div>
                <div class="text-xs text-fg-muted truncate font-mono">{user.userId}</div>
              </div>
              <div class="flex flex-wrap gap-1 max-w-[260px] justify-end">
                {#each user.intents.slice(0, 4) as intent}
                  <span class="text-[10px] px-1.5 h-4 leading-4 rounded bg-bg-elevated border border-border-default font-mono text-fg-muted">
                    {intent}
                  </span>
                {/each}
                {#if user.intents.length > 4}
                  <span class="text-[10px] px-1.5 h-4 leading-4 rounded text-fg-muted">
                    +{user.intents.length - 4}
                  </span>
                {/if}
              </div>
            </button>
          {/each}
          {#if filteredUsers.length === 0}
            <div class="p-6 text-xs text-fg-muted text-center">No users match.</div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</section>
