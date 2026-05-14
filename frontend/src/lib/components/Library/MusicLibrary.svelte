<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import MusicIcon from "lucide-svelte/icons/music";
  import Search from "lucide-svelte/icons/search";
  import { tabs } from "$lib/store/tabs.svelte";
  import { formatBytes, formatRelTime } from "$lib/components/Explorer/format";
  import { loadLibrary } from "./loader";
  import type { MediaEntry } from "$lib/types";

  // IntersectionObserver-driven lazy reveal — matches PdfViewer.svelte pattern.
  // 5000-row cap on the backend (MEDIA_LIB_LIMIT) means 200/200-step pagination
  // never has to virtualize.
  const INITIAL = 200;
  const STEP = 200;

  let entries = $state<MediaEntry[]>([]);
  let truncated = $state(false);
  let loading = $state(true);
  let errorMessage = $state<string | null>(null);
  let query = $state("");
  let visibleCount = $state(INITIAL);
  let sentinel = $state<HTMLDivElement | null>(null);
  let observer: IntersectionObserver | null = null;

  const filtered = $derived.by(() => {
    const q = query.trim().toLowerCase();
    if (!q) return entries;
    return entries.filter(
      (entry) =>
        entry.name.toLowerCase().includes(q) ||
        entry.loc.toLowerCase().includes(q),
    );
  });

  const visible = $derived(filtered.slice(0, visibleCount));

  // Reset pagination whenever the filter changes — otherwise scrolling state
  // from a previous query stays attached to the new (often shorter) list.
  $effect(() => {
    void query;
    visibleCount = INITIAL;
  });

  function stripExt(name: string): string {
    const idx = name.lastIndexOf(".");
    return idx > 0 ? name.slice(0, idx) : name;
  }

  function playTrack(entry: MediaEntry) {
    tabs.open({
      kind: "audio",
      title: entry.name,
      icon: "audio",
      payload: { loc: entry.loc, name: entry.name },
    });
  }

  // Re-attach the IntersectionObserver whenever the sentinel element appears
  // or disappears (it's `{#if visibleCount < filtered.length}`-gated). $effect
  // re-runs because `sentinel` is a `$state` ref.
  $effect(() => {
    observer?.disconnect();
    observer = null;
    if (!sentinel) return;
    observer = new IntersectionObserver(
      (intersections) => {
        for (const entry of intersections) {
          if (!entry.isIntersecting) continue;
          if (visibleCount >= filtered.length) continue;
          visibleCount = Math.min(filtered.length, visibleCount + STEP);
        }
      },
      { rootMargin: "300px" },
    );
    observer.observe(sentinel);
  });

  onMount(async () => {
    try {
      const result = await loadLibrary("audio");
      entries = result.entries;
      truncated = result.truncated;
    } catch (error) {
      errorMessage = (error as Error).message;
    } finally {
      loading = false;
    }
  });

  onDestroy(() => {
    observer?.disconnect();
    observer = null;
    sentinel = null;
  });
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <header
    class="flex items-center gap-2 px-4 h-12 border-b border-border-default bg-bg-overlay"
  >
    <MusicIcon size="16" class="text-fg-accent shrink-0" />
    <h2 class="text-sm font-semibold text-fg-primary">Music</h2>
    <span class="text-xs text-fg-muted tabular-nums">
      {#if loading}
        loading…
      {:else}
        {filtered.length} of {entries.length}
      {/if}
    </span>
    <div class="relative ml-auto w-72">
      <Search
        size="14"
        class="absolute left-2 top-1/2 -translate-y-1/2 text-fg-muted pointer-events-none"
      />
      <input
        type="text"
        bind:value={query}
        placeholder="Search tracks or folders…"
        class="w-full h-8 pl-7 pr-3 text-xs bg-bg-base border border-border-default rounded text-fg-primary placeholder:text-fg-muted focus:outline-none focus:border-accent"
      />
    </div>
  </header>

  {#if truncated}
    <div
      class="px-4 py-2 text-xs text-fg-warning bg-bg-elevated border-b border-border-default"
    >
      Library truncated at {entries.length} entries. Refine via search.
    </div>
  {/if}

  <div class="flex-1 overflow-y-auto">
    {#if loading}
      <div class="p-6 text-sm text-fg-muted">Loading library…</div>
    {:else if errorMessage}
      <div class="p-6 text-sm text-fg-danger">{errorMessage}</div>
    {:else if filtered.length === 0}
      <div class="p-6 text-sm text-fg-muted">
        {entries.length === 0 ? "No audio files found." : "No matches."}
      </div>
    {:else}
      <ul class="divide-y divide-border-default">
        {#each visible as entry (entry.loc + "/" + entry.name)}
          <li>
            <button
              type="button"
              class="flex items-center gap-3 w-full px-4 h-10 text-left hover:bg-bg-hover text-sm transition-colors"
              onclick={() => playTrack(entry)}
            >
              <MusicIcon size="14" class="text-fg-muted shrink-0" />
              <span class="flex-1 truncate text-fg-primary"
                >{stripExt(entry.name)}</span
              >
              <span class="hidden md:inline truncate max-w-[30%] text-xs text-fg-muted"
                >{entry.loc}</span
              >
              <span
                class="hidden sm:inline w-20 text-right text-xs text-fg-muted tabular-nums"
                >{formatBytes(entry.size)}</span
              >
              <span
                class="w-16 text-right text-xs text-fg-muted tabular-nums"
                >{formatRelTime(entry.modifiedAt)}</span
              >
            </button>
          </li>
        {/each}
      </ul>
      {#if visibleCount < filtered.length}
        <div bind:this={sentinel} class="h-8" aria-hidden="true"></div>
      {/if}
    {/if}
  </div>
</div>
