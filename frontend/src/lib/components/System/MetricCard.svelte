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

  const COLOR_OK = "#b8bb26";
  const COLOR_WARN = "#fabd2f";
  const COLOR_DANGER = "#fb4934";
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

<div class="p-5 rounded-lg bg-bg-surface border border-border-default flex flex-col gap-3 overflow-hidden">
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide">
      <Icon size="14" />
      <span>{label}</span>
    </div>
    <span
      class="text-[10px] font-mono px-2 py-0.5 rounded-full border"
      style="color: {color}; border-color: {color}40; background-color: {color}10;"
    >
      {value >= THRESHOLD_DANGER ? "HIGH" : value >= THRESHOLD_WARN ? "WARN" : "OK"}
    </span>
  </div>

  <div class="flex items-baseline gap-1.5">
    <span class="text-4xl font-mono font-semibold tabular-nums" style="color: {color};">{valueText}</span>
    <span class="text-sm text-fg-muted font-mono">%</span>
  </div>

  {#if detail}
    <div class="text-xs text-fg-muted font-mono">{detail}</div>
  {/if}

  <div class="-mx-1 mt-auto">
    <Sparkline points={history} {color} />
  </div>
</div>
