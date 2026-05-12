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
  <header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <Activity size="18" class="text-accent" />
    <h1 class="text-sm font-semibold text-fg-primary">System</h1>
    {#if history.length > 0}
      <span class="ml-2 text-[10px] text-fg-muted font-mono">
        {history.length}/{MAX_SAMPLES} samples · live
      </span>
    {/if}
  </header>

  <div class="flex-1 overflow-auto p-6">
    {#if error && !snapshot}
      <div class="p-4 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-sm max-w-md">
        Failed to load system info: {error}
      </div>
    {:else if !snapshot}
      <div class="text-sm text-fg-muted">Loading system metrics…</div>
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
        <div class="p-5 rounded-lg bg-bg-surface border border-border-default">
          <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-3">
            <Clock size="14" />
            <span>Uptime</span>
          </div>
          <div class="text-2xl font-mono font-semibold text-fg-primary truncate">{snapshot.uptime}</div>
        </div>

        <div class="p-5 rounded-lg bg-bg-surface border border-border-default">
          <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-3">
            <Activity size="14" />
            <span>Polling</span>
          </div>
          <div class="text-2xl font-mono font-semibold text-fg-primary">
            {POLL_INTERVAL_MS / 1000}s
          </div>
          <div class="text-xs text-fg-muted mt-1">
            History buffer: {MAX_SAMPLES} samples ({Math.round((MAX_SAMPLES * POLL_INTERVAL_MS) / 60000)} min)
          </div>
        </div>
      </div>

      {#if error}
        <div class="mt-4 max-w-6xl p-3 rounded-md bg-fg-warning/10 border border-fg-warning/30 text-fg-warning text-xs">
          Last refresh failed: {error}. Showing the most recent successful snapshot.
        </div>
      {/if}
    {/if}
  </div>
</section>
