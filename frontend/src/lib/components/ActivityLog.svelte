<script lang="ts">
  import { onMount } from "svelte";
  import History from "lucide-svelte/icons/history";
  import RefreshCw from "lucide-svelte/icons/refresh-cw";
  import { notifications } from "$lib/store/notifications.svelte";

  interface ActivityEntry {
    id?: number;
    activity: string;
    description: string;
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
      const raw = (data.logs ?? data ?? []) as ActivityEntry[];
      entries = raw.sort((a, b) => b.time - a.time);
    } catch (cause) {
      notifications.error(`Failed: ${(cause as Error).message}`);
    } finally {
      loading = false;
    }
  }

  const ACTIVITY_DOT: Record<string, string> = {
    UPLOAD: "bg-fg-success",
    DELETE: "bg-fg-danger",
    DOWNLOAD: "bg-fg-info",
    RENAME: "bg-fg-warning",
    COPY: "bg-fg-link",
    MOVE: "bg-fg-link",
    VIEW: "bg-fg-muted",
    OPEN: "bg-fg-muted",
  };

  const ACTIVITY_TEXT: Record<string, string> = {
    UPLOAD: "text-fg-success",
    DELETE: "text-fg-danger",
    DOWNLOAD: "text-fg-info",
    RENAME: "text-fg-warning",
    COPY: "text-fg-link",
    MOVE: "text-fg-link",
    VIEW: "text-fg-muted",
    OPEN: "text-fg-muted",
  };

  function dotClass(activity: string): string {
    return ACTIVITY_DOT[activity.toUpperCase()] ?? "bg-fg-secondary";
  }

  function textClass(activity: string): string {
    return ACTIVITY_TEXT[activity.toUpperCase()] ?? "text-fg-secondary";
  }

  function relTime(time: number): string {
    const diff = Date.now() - time;
    const minute = 60_000;
    const hour = 60 * minute;
    const day = 24 * hour;
    if (diff < minute) return "just now";
    if (diff < hour) return `${Math.floor(diff / minute)}m ago`;
    if (diff < day) return `${Math.floor(diff / hour)}h ago`;
    if (diff < 7 * day) return `${Math.floor(diff / day)}d ago`;
    return new Date(time).toLocaleDateString();
  }

  function fullTime(time: number): string {
    return new Date(time).toLocaleString();
  }

  onMount(load);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="flex items-center justify-between gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <div class="flex items-center gap-2">
      <History size="18" class="text-accent" />
      <h1 class="text-sm font-semibold text-fg-primary">Activity</h1>
      <span class="text-xs text-fg-muted font-mono ml-2">{entries.length}</span>
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
      <ol class="py-4">
        {#each entries as entry, index (entry.id ?? `${entry.time}-${entry.activity}`)}
          <li class="grid grid-cols-[40px_1fr] gap-3 px-6 group hover:bg-bg-hover/30 transition-colors">
            <div class="relative">
              {#if index > 0}
                <div class="absolute left-1/2 top-0 h-3 w-px -translate-x-1/2 bg-border-default"></div>
              {/if}
              <div class="absolute left-1/2 top-3 w-2.5 h-2.5 -translate-x-1/2 rounded-full ring-2 ring-bg-base {dotClass(entry.activity)}"></div>
              {#if index < entries.length - 1}
                <div class="absolute left-1/2 top-[1.375rem] bottom-0 w-px -translate-x-1/2 bg-border-default"></div>
              {/if}
            </div>
            <div class="py-2 min-w-0">
              <div class="flex items-baseline gap-2 mb-0.5">
                <span class="font-mono text-[10px] uppercase tracking-wide {textClass(entry.activity)}">
                  {entry.activity}
                </span>
                <span class="text-xs text-fg-muted" title={fullTime(entry.time)}>
                  {relTime(entry.time)}
                </span>
              </div>
              <div class="text-sm text-fg-primary break-words">{entry.description}</div>
              <div class="text-xs text-fg-muted mt-0.5 flex items-center gap-2 flex-wrap">
                {#if entry.krname || entry.username}
                  <span class="text-fg-secondary">{entry.krname || entry.username}</span>
                {/if}
                {#if entry.loc}
                  <span class="text-fg-muted">·</span>
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
