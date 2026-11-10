<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import FilmIcon from "lucide-svelte/icons/film";
  import Search from "lucide-svelte/icons/search";
  import { tabs } from "$lib/store/tabs.svelte";
  import { formatBytes } from "$lib/components/Explorer/format";
  import { loadLibrary } from "./loader";
  import type { MediaEntry } from "$lib/types";

  // 60-tile initial render + 60-per-sentinel reveal. Backend cap is 5000 so
  // worst-case ~83 sentinel hits — IntersectionObserver handles that trivially.
  const INITIAL = 60;
  const STEP = 60;

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

  $effect(() => {
    void query;
    visibleCount = INITIAL;
  });

  function stripExt(name: string): string {
    const idx = name.lastIndexOf(".");
    return idx > 0 ? name.slice(0, idx) : name;
  }

  function playVideo(entry: MediaEntry) {
    tabs.open({
      kind: "video",
      title: entry.name,
      icon: "video",
      payload: { loc: entry.loc, name: entry.name },
    });
  }

  // See MusicLibrary for rationale on $effect-driven observer re-attach.
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
      const result = await loadLibrary("video");
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
  <header class="sg-page-head flex-wrap py-2">
    <FilmIcon size="18" class="text-fg-secondary shrink-0" />
    <h1 class="sg-page-title">Videos</h1>
    <span class="text-xs text-fg-muted tabular-nums">
      {#if loading}
        loading…
      {:else}
        {filtered.length} of {entries.length}
      {/if}
    </span>
    <div class="relative w-full md:w-72 md:ml-auto">
      <Search
        size="16"
        class="absolute left-2.5 top-1/2 -translate-y-1/2 text-fg-muted pointer-events-none"
      />
      <input
        type="search"
        name="video-lib-q"
        autocomplete="off"
        autocorrect="off"
        spellcheck="false"
        bind:value={query}
        placeholder="Search videos or folders…"
        aria-label="Search videos or folders"
        class="sg-input sg-input-icon"
      />
    </div>
  </header>

  {#if truncated}
    <div
      class="px-gutter py-2 text-sm text-fg-warning border-b border-border-default"
      role="status"
    >
      Library truncated at {entries.length} entries. Refine via search.
    </div>
  {/if}

  <div class="flex-1 overflow-y-auto">
    {#if loading}
      <div class="px-gutter py-6 text-sm text-fg-muted" role="status">Loading library…</div>
    {:else if errorMessage}
      <div class="px-gutter py-6 text-sm text-fg-danger" role="alert">{errorMessage}</div>
    {:else if filtered.length === 0}
      <div class="px-gutter py-16 text-center text-sm text-fg-muted">
        {entries.length === 0 ? "No video files found." : "No matches."}
      </div>
    {:else}
      <div
        class="grid grid-cols-[repeat(auto-fill,minmax(min(100%,220px),1fr))] gap-4 px-gutter py-5"
      >
        {#each visible as entry (entry.loc + "/" + entry.name)}
          <button
            type="button"
            class="video-card group flex flex-col text-left overflow-hidden rounded-panel border border-border-default bg-bg-surface"
            onclick={() => playVideo(entry)}
          >
            <div
              class="aspect-video bg-bg-elevated flex items-center justify-center text-fg-muted"
            >
              <FilmIcon size="32" />
            </div>
            <div class="flex flex-col gap-0.5 px-3 py-2.5">
              <span
                class="text-sm font-semibold text-fg-primary truncate"
                title={entry.name}>{stripExt(entry.name)}</span
              >
              <span class="text-xs text-fg-muted truncate" title={entry.loc}
                >{entry.loc}</span
              >
              <span class="text-xs text-fg-muted tabular-nums"
                >{formatBytes(entry.size)} · —</span
              >
            </div>
          </button>
        {/each}
      </div>
      {#if visibleCount < filtered.length}
        <div bind:this={sentinel} class="h-8" aria-hidden="true"></div>
      {/if}
    {/if}
  </div>
</div>

<style>
  .video-card {
    transition: border-color 70ms var(--ease-snap), background-color 70ms var(--ease-snap);
  }
  .video-card:hover {
    border-color: var(--color-border-strong);
    background: var(--color-bg-elevated);
  }
</style>
