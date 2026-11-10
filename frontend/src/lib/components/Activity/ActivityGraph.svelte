<script lang="ts">
  import Sparkline from "$lib/components/System/Sparkline.svelte";
  import { dailyCounts, typeDistribution, type ActivityEntry } from "./aggregate";

  interface Props {
    entries: ActivityEntry[];
    loading?: boolean;
  }

  let { entries, loading = false }: Props = $props();

  const ACTIVITY_DOT: Record<string, string> = {
    UPLOAD: "bg-fg-success",
    DELETE: "bg-fg-danger",
    DOWNLOAD: "bg-fg-info",
    RENAME: "bg-fg-warning",
    COPY: "bg-fg-purple",
    MOVE: "bg-fg-purple",
    VIEW: "bg-fg-muted",
    OPEN: "bg-fg-muted",
  };

  const daily = $derived(dailyCounts(entries, 30));
  const counts = $derived(daily.map((d) => d.count));
  const ceiling = $derived(Math.max(1, ...counts));
  const dist = $derived(typeDistribution(entries));
</script>

<section class="px-gutter py-4 border-b border-border-default bg-bg-base flex flex-col md:flex-row gap-4 md:gap-6" aria-label="Activity summary">
  <div class="flex-1 min-w-0">
    <h2 class="text-xs font-semibold text-fg-secondary mb-2">Last 30 days</h2>
    {#if loading}
      <div class="h-12 rounded-control bg-bg-elevated animate-pulse" aria-hidden="true"></div>
      <span class="sr-only" role="status">Loading activity</span>
    {:else if entries.length === 0}
      <p class="h-12 flex items-center text-sm text-fg-muted">No activity in the last 30 days.</p>
    {:else}
      <Sparkline points={counts} color="var(--color-fg-info)" height={48} max={ceiling} />
      <div class="flex justify-between text-xs text-fg-muted tabular-nums mt-1">
        <span>{daily[0]?.day ?? ""}</span>
        <span>{daily.at(-1)?.day ?? "today"}</span>
      </div>
    {/if}
  </div>

  <div class="md:w-72 min-w-0">
    <h2 class="text-xs font-semibold text-fg-secondary mb-2">Distribution</h2>
    {#if loading}
      <div class="h-2 rounded-full bg-bg-elevated animate-pulse" aria-hidden="true"></div>
      <div class="flex flex-wrap gap-x-2 gap-y-1 mt-2" aria-hidden="true">
        <span class="h-3 w-12 rounded-chip bg-bg-elevated animate-pulse"></span>
        <span class="h-3 w-12 rounded-chip bg-bg-elevated animate-pulse"></span>
        <span class="h-3 w-12 rounded-chip bg-bg-elevated animate-pulse"></span>
      </div>
    {:else if entries.length === 0}
      <p class="text-sm text-fg-muted">Nothing to show yet.</p>
    {:else}
      <div class="h-2 rounded-full overflow-hidden bg-bg-elevated flex gap-px">
        {#each dist as share (share.type)}
          <div
            class={ACTIVITY_DOT[share.type] ?? "bg-fg-secondary"}
            style="width: {(share.ratio * 100).toFixed(2)}%"
          ></div>
        {/each}
      </div>
      <ul class="flex flex-wrap gap-x-3 gap-y-1 mt-2 text-xs text-fg-secondary">
        {#each dist as share (share.type)}
          <li title="{share.count} · {(share.ratio * 100).toFixed(0)}%" class="inline-flex items-center gap-1.5">
            <span class="size-2 rounded-full {ACTIVITY_DOT[share.type] ?? 'bg-fg-secondary'}" aria-hidden="true"></span>
            {share.type}
            <span class="text-fg-muted tabular-nums">{share.count}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</section>
