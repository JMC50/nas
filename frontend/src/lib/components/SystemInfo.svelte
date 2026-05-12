<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Cpu from "lucide-svelte/icons/cpu";
  import MemoryStick from "lucide-svelte/icons/memory-stick";
  import HardDrive from "lucide-svelte/icons/hard-drive";
  import Clock from "lucide-svelte/icons/clock";

  interface DiskInfo {
    total: string;
    used: string;
    available: string;
    usagePercentage: string;
  }

  interface SystemInfoData {
    cpu: string;
    memory: string;
    uptime: string;
    disk: DiskInfo;
  }

  const POLL_INTERVAL_MS = 5000;

  let data: SystemInfoData | null = $state(null);
  let error: string | null = $state(null);
  let timer: ReturnType<typeof setInterval> | null = null;

  async function fetchSystemData() {
    try {
      const response = await fetch("/server/getSystemInfo");
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      data = await response.json();
      error = null;
    } catch (err) {
      error = (err as Error).message;
    }
  }

  function diskPercent(): number {
    if (!data) return 0;
    return Number.parseInt(data.disk.usagePercentage.replace("%", ""), 10) || 0;
  }

  onMount(() => {
    fetchSystemData();
    timer = setInterval(fetchSystemData, POLL_INTERVAL_MS);
  });

  onDestroy(() => {
    if (timer) clearInterval(timer);
  });
</script>

<section class="h-full overflow-auto p-6 bg-bg-base">
  <header class="flex items-center gap-2 mb-6">
    <Cpu size="20" class="text-accent" />
    <h1 class="text-xl font-semibold text-fg-primary">System</h1>
  </header>

  {#if error}
    <div class="p-3 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-sm">
      Failed to load system info: {error}
    </div>
  {:else if !data}
    <div class="text-sm text-fg-muted">Loading…</div>
  {:else}
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="p-4 rounded-lg bg-bg-surface border border-border-default">
        <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-2">
          <Cpu size="14" />
          <span>CPU</span>
        </div>
        <div class="text-2xl font-mono font-semibold text-fg-primary">{data.cpu}</div>
      </div>

      <div class="p-4 rounded-lg bg-bg-surface border border-border-default">
        <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-2">
          <MemoryStick size="14" />
          <span>Memory</span>
        </div>
        <div class="text-2xl font-mono font-semibold text-fg-primary">{data.memory}</div>
      </div>

      <div class="p-4 rounded-lg bg-bg-surface border border-border-default">
        <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-2">
          <Clock size="14" />
          <span>Uptime</span>
        </div>
        <div class="text-2xl font-mono font-semibold text-fg-primary truncate">{data.uptime}</div>
      </div>

      <div class="p-4 rounded-lg bg-bg-surface border border-border-default">
        <div class="flex items-center gap-2 text-fg-muted text-xs uppercase tracking-wide mb-2">
          <HardDrive size="14" />
          <span>Disk</span>
        </div>
        <div class="text-2xl font-mono font-semibold text-fg-primary mb-2">
          {data.disk.usagePercentage}
        </div>
        <div class="text-xs text-fg-muted font-mono mb-1.5">
          {data.disk.used}B / {data.disk.total}B
        </div>
        <div class="h-1.5 rounded-full bg-bg-elevated overflow-hidden">
          <div
            class="h-full transition-[width] duration-300 {diskPercent() > 80 ? 'bg-fg-danger' : diskPercent() > 60 ? 'bg-fg-warning' : 'bg-fg-success'}"
            style="width: {diskPercent()}%;"
          ></div>
        </div>
      </div>
    </div>
  {/if}
</section>
