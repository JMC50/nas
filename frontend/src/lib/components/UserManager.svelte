<script lang="ts">
  import { onMount } from "svelte";
  import Shield from "lucide-svelte/icons/shield";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import type { Intent } from "$lib/types";

  interface Props {
    initialUserId?: string;
  }

  let { initialUserId = "" }: Props = $props();

  interface UserView {
    userId: string;
    username: string;
    global_name: string;
    krname: string;
    intents: Intent[];
  }

  const ALL_INTENTS: Intent[] = [
    "VIEW",
    "OPEN",
    "DOWNLOAD",
    "UPLOAD",
    "COPY",
    "DELETE",
    "RENAME",
    "ADMIN",
  ];

  let userId = $state(initialUserId);
  let users: UserView[] = $state([]);
  let user = $state<UserView | null>(null);
  let busy = $state<Set<Intent>>(new Set());
  let deleting = $state(false);
  let loadError = $state("");
  const isSelf = $derived(user !== null && user.userId === auth.current.userId);

  async function loadUsers() {
    loadError = "";
    const response = await fetch("/server/getAllUsers", {
      headers: { Authorization: `Bearer ${auth.token}` },
    });
    if (!response.ok) {
      loadError = response.status === 403 ? "Only administrators can list users." : `Failed to load users (HTTP ${response.status}).`;
      return;
    }
    const data = await response.json();
    const raw = (data.users ?? []) as Partial<UserView>[];
    users = raw.map((entry) => ({
      userId: entry.userId ?? "",
      username: entry.username ?? "",
      global_name: entry.global_name ?? "",
      krname: entry.krname ?? "",
      intents: entry.intents ?? [],
    }));
    if (userId) {
      user = users.find((entry) => entry.userId === userId) ?? null;
    }
  }

  function selectUser(id: string) {
    userId = id;
    user = users.find((entry) => entry.userId === id) ?? null;
  }

  function setBusy(intent: Intent, value: boolean) {
    const next = new Set(busy);
    if (value) next.add(intent);
    else next.delete(intent);
    busy = next;
  }

  async function callToggle(target: UserView, intent: Intent, granted: boolean) {
    const endpoint = granted ? "unauthorize" : "authorize";
    const response = await fetch(
      `/server/${endpoint}?userId=${encodeURIComponent(target.userId)}&intent=${intent}&token=${encodeURIComponent(auth.token)}`,
    );
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    target.intents = granted
      ? target.intents.filter((value) => value !== intent)
      : [...target.intents, intent];
  }

  async function toggleIntent(intent: Intent) {
    if (!user) return;
    const granted = user.intents.includes(intent);
    setBusy(intent, true);
    try {
      await callToggle(user, intent, granted);
      notifications.success(`${granted ? "Revoked" : "Granted"} ${intent}`);
    } catch (cause) {
      notifications.error(`Toggle failed: ${(cause as Error).message}`);
    } finally {
      setBusy(intent, false);
    }
  }

  async function deleteUser() {
    if (!user) return;
    if (user.userId === auth.current.userId) return;
    const label = user.krname || user.global_name || user.username || user.userId;
    if (!confirm(`Delete user "${label}" (${user.userId})? This cannot be undone.`)) return;
    const removedId = user.userId;
    deleting = true;
    try {
      const response = await fetch(
        `/server/users/${encodeURIComponent(removedId)}?token=${encodeURIComponent(auth.token)}`,
        { method: "DELETE" },
      );
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      notifications.success(`Deleted ${label}`);
      users = users.filter((entry) => entry.userId !== removedId);
      user = null;
      userId = "";
    } catch (cause) {
      notifications.error(`Delete failed: ${(cause as Error).message}`);
    } finally {
      deleting = false;
    }
  }

  onMount(loadUsers);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="sg-page-head">
    <Shield size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">User permissions</h1>
    {#if users.length > 0}
      <span class="sg-count">{users.length}</span>
    {/if}
  </header>

  <div class="flex-1 grid grid-rows-[auto_minmax(0,1fr)] md:grid-rows-1 md:grid-cols-[280px_minmax(0,1fr)] min-h-0">
    <aside class="md:border-r border-b md:border-b-0 border-border-default overflow-auto max-h-[40dvh] md:max-h-none bg-bg-surface" aria-label="Users">
      {#if loadError}
        <div class="m-3 p-3 rounded-panel border border-fg-danger/45 text-sm text-fg-danger" role="alert">{loadError}</div>
      {:else if users.length === 0}
        <div class="p-4 text-sm text-fg-muted" role="status">Loading users…</div>
      {/if}
      <ul>
        {#each users as entry (entry.userId)}
          <li>
            <button
              type="button"
              class="user-row {entry.userId === userId ? 'is-current' : ''}"
              onclick={() => selectUser(entry.userId)}
              aria-current={entry.userId === userId ? "true" : undefined}
            >
              <span class="grid place-items-center size-8 shrink-0 rounded-control bg-bg-elevated text-fg-secondary text-xs font-bold" aria-hidden="true">
                {(entry.krname || entry.global_name || entry.username || "?").slice(0, 1).toUpperCase()}
              </span>
              <span class="min-w-0 flex-1">
                <span class="block text-sm text-fg-primary truncate">{entry.krname || entry.global_name || entry.username}</span>
                <span class="block text-xs text-fg-muted truncate font-mono">{entry.userId}</span>
              </span>
              {#if entry.intents.includes("ADMIN")}
                <span class="sg-chip font-mono text-fg-purple">ADMIN</span>
              {/if}
            </button>
          </li>
        {/each}
      </ul>
    </aside>

    <div class="overflow-auto px-gutter py-5">
      {#if !user}
        <div class="text-sm text-fg-muted">Select a user from the list.</div>
      {:else}
        <div class="max-w-xl flex flex-col gap-5">
          <div class="flex items-center gap-3 min-w-0">
            <span class="grid place-items-center size-10 shrink-0 rounded-control bg-bg-elevated text-fg-secondary text-sm font-bold" aria-hidden="true">
              {(user.krname || user.global_name || user.username || "?").slice(0, 1).toUpperCase()}
            </span>
            <div class="min-w-0">
              <h2 class="text-md font-semibold text-fg-primary truncate">
                {user.krname || user.global_name || user.username}
              </h2>
              <div class="text-xs text-fg-muted font-mono truncate">{user.userId}</div>
            </div>
          </div>

          <ul class="sg-panel divide-y divide-border-default" aria-label="Permissions for {user.userId}">
            {#each ALL_INTENTS as intent}
              {@const granted = user.intents.includes(intent)}
              {@const pending = busy.has(intent)}
              <li class="flex items-center justify-between gap-4 px-4 min-h-11">
                <span class="text-sm font-mono text-fg-primary">{intent}</span>
                <button
                  type="button"
                  role="switch"
                  class="sg-switch"
                  onclick={() => toggleIntent(intent)}
                  disabled={pending}
                  aria-label={intent}
                  aria-checked={granted}
                ></button>
              </li>
            {/each}
          </ul>

          <div class="pt-5 border-t border-border-default">
            <h3 class="text-sm font-semibold text-fg-danger mb-2">Danger zone</h3>
            <div class="flex flex-wrap items-center justify-between gap-3">
              <p class="text-xs text-fg-secondary leading-relaxed max-w-[48ch]">
                {isSelf
                  ? "You can't delete your own account from here."
                  : "Permanently remove this user and their permissions. Activity log entries owned by this user are also removed."}
              </p>
              <button
                type="button"
                class="sg-btn sg-btn-danger shrink-0"
                onclick={deleteUser}
                disabled={isSelf || deleting}
                aria-label="Delete user"
              >
                <Trash2 size="16" />
                {deleting ? "Deleting…" : "Delete user"}
              </button>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
</section>

<style>
  .user-row {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    min-height: 52px;
    padding: 6px 12px;
    border-bottom: 1px solid var(--color-border-default);
    text-align: left;
    transition: background-color 70ms var(--ease-snap);
  }
  .user-row:hover {
    background: var(--color-bg-hover);
  }
  .user-row:focus-visible {
    outline-offset: -2px;
  }
  .user-row.is-current {
    background: var(--color-bg-selected);
    box-shadow: inset 2px 0 0 var(--color-accent);
  }
</style>
