<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Activity from "lucide-svelte/icons/activity";
  import Cpu from "lucide-svelte/icons/cpu";
  import MemoryStick from "lucide-svelte/icons/memory-stick";
  import HardDrive from "lucide-svelte/icons/hard-drive";
  import Clock from "lucide-svelte/icons/clock";
  import MetricCard from "$lib/components/System/MetricCard.svelte";

  interface DiskInfo {
    total: string;
    used: string;
    free: string;
    usedPercent: number;
  }

  interface SystemSnapshot {
    cpu: number;
    memory: number;
    uptime: string;
    disk: DiskInfo;
  }

  interface Sample {
    timestamp: number;
    cpu: number;
    memory: number;
    disk: number;
  }

  const POLL_INTERVAL_MS = 5000;
  const MAX_SAMPLES = 60;

  let snapshot = $state<SystemSnapshot | null>(null);
  let history = $state<Sample[]>([]);
  let error = $state<string | null>(null);
  let timer: ReturnType<typeof setInterval> | null = null;

  const cpuPoints = $derived(history.map((sample) => sample.cpu));
  const memoryPoints = $derived(history.map((sample) => sample.memory));
  const diskPoints = $derived(history.map((sample) => sample.disk));

  function record(next: SystemSnapshot) {
    const sample: Sample = {
      timestamp: Date.now(),
      cpu: next.cpu,
      memory: next.memory,
      disk: next.disk.usedPercent,
    };
    history = [...history, sample].slice(-MAX_SAMPLES);
  }

  async function fetchStats() {
    try {
      const response = await fetch("/server/getSystemInfo");
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const next = (await response.json()) as SystemSnapshot;
      snapshot = next;
      record(next);
      error = null;
    } catch (cause) {
      error = (cause as Error).message;
    }
  }

  onMount(() => {
    fetchStats();
    timer = setInterval(fetchStats, POLL_INTERVAL_MS);
  });

  onDestroy(() => {
    if (timer) clearInterval(timer);
  });
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="sg-page-head">
    <Activity size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">System</h1>
    {#if history.length > 0}
      <span class="ml-auto text-xs text-fg-muted tabular-nums">
        {history.length}/{MAX_SAMPLES} samples · live
      </span>
    {/if}
  </header>

  <div class="flex-1 overflow-auto px-gutter py-6">
    {#if error && !snapshot}
      <div class="p-4 rounded-panel border border-fg-danger/45 text-fg-danger text-sm max-w-md" role="alert">
        Failed to load system info: {error}
      </div>
    {:else if !snapshot}
      <div class="text-sm text-fg-muted" role="status">Loading system metrics…</div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 max-w-6xl">
        <MetricCard label="CPU" value={snapshot.cpu} history={cpuPoints} icon={Cpu} />
        <MetricCard label="Memory" value={snapshot.memory} history={memoryPoints} icon={MemoryStick} />
        <MetricCard
          label="Disk"
          value={snapshot.disk.usedPercent}
          history={diskPoints}
          icon={HardDrive}
          detail={`${snapshot.disk.used} / ${snapshot.disk.total} · ${snapshot.disk.free} free`}
        />
      </div>

      <div class="mt-4 max-w-6xl grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="sg-panel p-5">
          <div class="flex items-center gap-2 text-fg-secondary text-sm font-semibold mb-3">
            <Clock size="16" />
            <h2>Uptime</h2>
          </div>
          <div class="text-xl font-semibold tabular-nums text-fg-primary truncate">{snapshot.uptime}</div>
        </div>

        <div class="sg-panel p-5">
          <div class="flex items-center gap-2 text-fg-secondary text-sm font-semibold mb-3">
            <Activity size="16" />
            <h2>Polling</h2>
          </div>
          <div class="text-xl font-semibold tabular-nums text-fg-primary">
            {POLL_INTERVAL_MS / 1000}s
          </div>
          <div class="text-xs text-fg-muted mt-1 tabular-nums">
            History buffer: {MAX_SAMPLES} samples ({Math.round((MAX_SAMPLES * POLL_INTERVAL_MS) / 60000)} min)
          </div>
        </div>
      </div>

      {#if error}
        <div class="mt-4 max-w-6xl p-3 rounded-panel border border-fg-warning/45 text-fg-warning text-sm" role="status">
          Last refresh failed: {error}. Showing the most recent successful snapshot.
        </div>
      {/if}
    {/if}
  </div>
</section>
