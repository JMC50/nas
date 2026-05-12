<script lang="ts">
  import { onMount } from "svelte";
  import Info from "lucide-svelte/icons/info";

  interface Health {
    status: string;
    db: string;
    schema: string;
  }

  const APP_VERSION = "2.0.0";

  let health = $state<Health | null>(null);

  async function load() {
    try {
      const response = await fetch("/server/healthz");
      health = await response.json();
    } catch {
      health = null;
    }
  }

  function statusColor(value: string): string {
    if (value === "ok" || value === "connected" || value === "valid") return "text-fg-success";
    return "text-fg-danger";
  }

  onMount(load);
</script>

<section class="space-y-3">
  <div class="flex items-center gap-2">
    <Info size="14" class="text-fg-muted" />
    <h2 class="text-xs font-semibold uppercase tracking-wide text-fg-muted">About</h2>
  </div>

  <div class="rounded-lg bg-bg-surface border border-border-default p-4 space-y-2">
    <div class="grid grid-cols-[140px_1fr] gap-x-4 gap-y-1.5 text-xs">
      <div class="text-fg-muted">App version</div>
      <div class="font-mono text-fg-primary">{APP_VERSION}</div>

      <div class="text-fg-muted">Backend</div>
      <div class="font-mono text-fg-primary">Go server</div>

      {#if health}
        <div class="text-fg-muted">Health</div>
        <div class="font-mono {statusColor(health.status)}">{health.status}</div>

        <div class="text-fg-muted">Database</div>
        <div class="font-mono {statusColor(health.db)}">{health.db}</div>

        <div class="text-fg-muted">Schema</div>
        <div class="font-mono {statusColor(health.schema)}">{health.schema}</div>
      {:else}
        <div class="text-fg-muted">Health</div>
        <div class="font-mono text-fg-muted">checking…</div>
      {/if}
    </div>
  </div>
</section>
