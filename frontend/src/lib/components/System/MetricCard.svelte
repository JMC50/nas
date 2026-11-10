<script lang="ts">
  import Cpu from "lucide-svelte/icons/cpu";
  import Sparkline from "$lib/components/System/Sparkline.svelte";

  interface Props {
    label: string;
    value: number;
    history: number[];
    icon: typeof Cpu;
    detail?: string;
  }

  let { label, value, history, icon: Icon, detail }: Props = $props();

  const COLOR_OK = "var(--color-fg-success)";
  const COLOR_WARN = "var(--color-fg-warning)";
  const COLOR_DANGER = "var(--color-fg-danger)";
  const THRESHOLD_WARN = 60;
  const THRESHOLD_DANGER = 80;

  const color = $derived(pickColor(value));
  const valueText = $derived(value.toFixed(1));

  function pickColor(percent: number): string {
    if (percent >= THRESHOLD_DANGER) return COLOR_DANGER;
    if (percent >= THRESHOLD_WARN) return COLOR_WARN;
    return COLOR_OK;
  }
</script>

<div class="sg-panel p-5 flex flex-col gap-3 overflow-hidden">
  <div class="flex items-center justify-between gap-2">
    <h2 class="flex items-center gap-2 text-fg-secondary text-sm font-semibold">
      <Icon size="16" />
      <span>{label}</span>
    </h2>
    <span class="sg-chip font-mono" style="color: {color};">
      {value >= THRESHOLD_DANGER ? "HIGH" : value >= THRESHOLD_WARN ? "WARN" : "OK"}
    </span>
  </div>

  <div class="flex items-baseline gap-1">
    <span class="text-2xl font-semibold tabular-nums" style="color: {color};">{valueText}</span>
    <span class="text-sm text-fg-muted">%</span>
  </div>

  {#if detail}
    <div class="text-xs text-fg-muted tabular-nums">{detail}</div>
  {/if}

  <div class="-mx-1 mt-auto">
    <Sparkline points={history} {color} />
  </div>
</div>
