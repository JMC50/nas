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
  <header
    class="flex items-center gap-2 px-4 h-12 border-b border-border-default bg-bg-overlay"
  >
    <FilmIcon size="16" class="text-fg-accent shrink-0" />
    <h2 class="text-sm font-semibold text-fg-primary">Videos</h2>
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
        type="search"
        name="video-lib-q"
        autocomplete="off"
        autocorrect="off"
        spellcheck="false"
        bind:value={query}
        placeholder="Search videos or folders…"
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
        {entries.length === 0 ? "No video files found." : "No matches."}
      </div>
    {:else}
      <div
        class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-4 p-6"
      >
        {#each visible as entry (entry.loc + "/" + entry.name)}
          <button
            type="button"
            class="group flex flex-col text-left rounded overflow-hidden border border-border-default bg-bg-surface hover:border-accent transition-colors"
            onclick={() => playVideo(entry)}
          >
            <div
              class="aspect-video bg-bg-elevated flex items-center justify-center text-fg-muted group-hover:text-fg-accent transition-colors"
            >
              <FilmIcon size="32" />
            </div>
            <div class="flex flex-col gap-0.5 px-2 py-2">
              <span
                class="text-xs font-medium text-fg-primary truncate"
                title={entry.name}>{stripExt(entry.name)}</span
              >
              <span class="text-[10px] text-fg-muted truncate" title={entry.loc}
                >{entry.loc}</span
              >
              <span class="text-[10px] text-fg-muted tabular-nums"
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
