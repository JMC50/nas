<script lang="ts">
  import { onMount } from "svelte";
  import Shield from "lucide-svelte/icons/shield";
  import UserIcon from "lucide-svelte/icons/user";
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
  const isSelf = $derived(user !== null && user.userId === auth.current.userId);

  async function loadUsers() {
    const response = await fetch("/server/getAllUsers");
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
  <header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <Shield size="18" class="text-accent" />
    <h1 class="text-sm font-semibold text-fg-primary">User permissions</h1>
  </header>

  <div class="flex-1 grid grid-cols-[280px_1fr] min-h-0">
    <aside class="border-r border-border-default overflow-auto bg-bg-surface/30">
      {#if users.length === 0}
        <div class="p-4 text-xs text-fg-muted">Loading users…</div>
      {/if}
      {#each users as entry (entry.userId)}
        <button
          type="button"
          class="w-full flex items-center gap-2 px-4 py-2.5 border-b border-border-default/40 text-left transition-colors {entry.userId === userId
            ? 'bg-bg-hover text-fg-primary'
            : 'text-fg-secondary hover:bg-bg-hover/60 hover:text-fg-primary'}"
          onclick={() => selectUser(entry.userId)}
        >
          <UserIcon size="14" class="shrink-0 text-fg-muted" />
          <div class="min-w-0 flex-1">
            <div class="text-sm truncate">
              {entry.krname || entry.global_name || entry.username}
            </div>
            <div class="text-xs text-fg-muted truncate font-mono">{entry.userId}</div>
          </div>
          {#if entry.intents.includes("ADMIN")}
            <span class="text-[10px] px-1.5 h-4 rounded bg-accent/15 text-fg-accent border border-accent/30 font-mono leading-4">
              ADMIN
            </span>
          {/if}
        </button>
      {/each}
    </aside>

    <div class="overflow-auto p-6">
      {#if !user}
        <div class="text-sm text-fg-muted">Select a user from the list.</div>
      {:else}
        <div class="max-w-xl">
          <div class="flex items-center gap-3 mb-5">
            <div class="w-10 h-10 rounded-full bg-bg-elevated border border-border-default flex items-center justify-center">
              <UserIcon size="18" class="text-fg-muted" />
            </div>
            <div class="min-w-0">
              <div class="text-base font-semibold text-fg-primary truncate">
                {user.krname || user.global_name || user.username}
              </div>
              <div class="text-xs text-fg-muted font-mono truncate">{user.userId}</div>
            </div>
          </div>

          <div class="rounded-lg bg-bg-surface border border-border-default divide-y divide-border-default/60">
            {#each ALL_INTENTS as intent}
              {@const granted = user.intents.includes(intent)}
              {@const pending = busy.has(intent)}
              <div class="flex items-center justify-between px-4 h-11">
                <div class="text-sm font-mono text-fg-primary">{intent}</div>
                <button
                  type="button"
                  class="relative inline-flex items-center h-5 w-9 rounded-full transition-colors disabled:opacity-60 {granted
                    ? 'bg-accent'
                    : 'bg-bg-elevated border border-border-default'}"
                  onclick={() => toggleIntent(intent)}
                  disabled={pending}
                  aria-label={granted ? `Revoke ${intent}` : `Grant ${intent}`}
                  aria-pressed={granted}
                >
                  <span
                    class="inline-block w-3.5 h-3.5 rounded-full bg-bg-base shadow transition-transform {granted ? 'translate-x-[18px]' : 'translate-x-0.5'}"
                  ></span>
                </button>
              </div>
            {/each}
          </div>

          <div class="mt-6 pt-5 border-t border-border-default/60">
            <div class="text-xs font-semibold text-fg-muted uppercase tracking-wide mb-2">Danger zone</div>
            <div class="flex items-center justify-between gap-4">
              <div class="text-xs text-fg-muted leading-relaxed">
                {isSelf
                  ? "You can't delete your own account from here."
                  : "Permanently remove this user and their permissions. Activity log entries owned by this user are also removed."}
              </div>
              <button
                type="button"
                class="inline-flex items-center gap-1.5 h-8 px-3 rounded-md text-xs font-medium text-fg-danger border border-border-default hover:bg-bg-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors shrink-0"
                onclick={deleteUser}
                disabled={isSelf || deleting}
                aria-label="Delete user"
              >
                <Trash2 size="12" />
                {deleting ? "Deleting…" : "Delete user"}
              </button>
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
</section>
