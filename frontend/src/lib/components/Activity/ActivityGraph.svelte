<script lang="ts">
  import Sparkline from "$lib/components/System/Sparkline.svelte";
  import { dailyCounts, typeDistribution, type ActivityEntry } from "./aggregate";

  interface Props {
    entries: ActivityEntry[];
  }

  let { entries }: Props = $props();

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

  const daily = $derived(dailyCounts(entries, 30));
  const counts = $derived(daily.map((d) => d.count));
  const ceiling = $derived(Math.max(1, ...counts));
  const dist = $derived(typeDistribution(entries));
</script>

<section class="px-6 py-3 border-b border-border-default bg-bg-base flex gap-4">
  <div class="flex-1">
    <div class="text-[10px] uppercase tracking-wide text-fg-muted mb-1">Last 30 days</div>
    {#if entries.length === 0}
      <div class="h-12 bg-bg-elevated rounded animate-pulse"></div>
    {:else}
      <Sparkline points={counts} color="#fabd2f" height={48} max={ceiling} />
      <div class="flex justify-between text-[10px] font-mono text-fg-muted mt-1">
        <span>{daily[0]?.day ?? ""}</span>
        <span>{daily.at(-1)?.day ?? "today"}</span>
      </div>
    {/if}
  </div>

  <div class="w-72">
    <div class="text-[10px] uppercase tracking-wide text-fg-muted mb-1">Distribution</div>
    {#if entries.length === 0}
      <div class="h-2 bg-bg-elevated rounded-full animate-pulse"></div>
      <div class="flex flex-wrap gap-x-2 gap-y-1 mt-1.5">
        <span class="h-3 w-12 bg-bg-elevated rounded animate-pulse"></span>
        <span class="h-3 w-12 bg-bg-elevated rounded animate-pulse"></span>
        <span class="h-3 w-12 bg-bg-elevated rounded animate-pulse"></span>
      </div>
    {:else}
      <div class="h-2 rounded-full overflow-hidden bg-bg-elevated flex">
        {#each dist as share (share.type)}
          <div
            class={ACTIVITY_DOT[share.type] ?? "bg-fg-secondary"}
            style="width: {(share.ratio * 100).toFixed(2)}%"
          ></div>
        {/each}
      </div>
      <div class="flex flex-wrap gap-x-2 gap-y-1 mt-1.5 text-[10px]">
        {#each dist as share (share.type)}
          <span
            title="{share.count} · {(share.ratio * 100).toFixed(0)}%"
            class="inline-flex items-center gap-1"
          >
            <span class="w-1.5 h-1.5 rounded-full {ACTIVITY_DOT[share.type] ?? 'bg-fg-secondary'}"></span>
            {share.type}
            <span class="font-mono text-fg-muted">{share.count}</span>
          </span>
        {/each}
      </div>
    {/if}
  </div>
</section>
