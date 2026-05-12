<script lang="ts">
  import { onMount } from "svelte";
  import History from "lucide-svelte/icons/history";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import { notifications } from "$lib/store/notifications.svelte";

  interface ActivityEntry {
    id?: number;
    activity: string;
    description: string;
    user_id?: number;
    userId?: string;
    krname?: string;
    username?: string;
    time: number;
    loc?: string;
  }

  let entries: ActivityEntry[] = $state([]);
  let loading = $state(true);

  async function load() {
    loading = true;
    try {
      const response = await fetch("/server/getActivityLog");
      const data = await response.json();
      entries = (data.logs ?? data ?? []) as ActivityEntry[];
    } catch (err) {
      notifications.error(`Failed: ${(err as Error).message}`);
    } finally {
      loading = false;
    }
  }

  function formatTime(time: number): string {
    return new Date(time).toLocaleString();
  }

  const ACTIVITY_COLORS: Record<string, string> = {
    UPLOAD: "text-fg-success",
    DELETE: "text-fg-danger",
    DOWNLOAD: "text-fg-info",
    RENAME: "text-fg-warning",
    COPY: "text-fg-link",
    VIEW: "text-fg-muted",
    OPEN: "text-fg-muted",
  };

  function colorFor(activity: string): string {
    return ACTIVITY_COLORS[activity.toUpperCase()] ?? "text-fg-secondary";
  }

  onMount(load);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="flex items-center justify-between gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <div class="flex items-center gap-2">
      <History size="18" class="text-accent" />
      <h1 class="text-sm font-semibold text-fg-primary">Activity</h1>
    </div>
    <button
      type="button"
      class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-md text-fg-muted hover:text-fg-primary hover:bg-bg-hover text-xs transition-colors"
      onclick={load}
      aria-label="Refresh"
    >
      <RefreshCw size="13" />
      <span>Refresh</span>
    </button>
  </header>

  <div class="flex-1 overflow-auto">
    {#if loading}
      <div class="p-6 text-sm text-fg-muted">Loading…</div>
    {:else if entries.length === 0}
      <div class="p-12 text-center text-sm text-fg-muted">No activity yet.</div>
    {:else}
      <ol class="divide-y divide-border-default/40">
        {#each entries as entry (entry.id ?? `${entry.time}-${entry.activity}`)}
          <li class="flex items-start gap-4 px-6 py-3 hover:bg-bg-hover/50 transition-colors">
            <div class="font-mono text-[10px] uppercase tracking-wide w-20 pt-0.5 shrink-0 {colorFor(entry.activity)}">
              {entry.activity}
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm text-fg-primary break-words">{entry.description}</div>
              <div class="text-xs text-fg-muted mt-0.5 flex items-center gap-2 flex-wrap">
                <span class="font-mono">{formatTime(entry.time)}</span>
                {#if entry.krname || entry.username}
                  <span>·</span>
                  <span class="text-fg-secondary">{entry.krname || entry.username}</span>
                {/if}
                {#if entry.loc}
                  <span>·</span>
                  <span class="font-mono truncate">{entry.loc}</span>
                {/if}
              </div>
            </div>
          </li>
        {/each}
      </ol>
    {/if}
  </div>
</section>
