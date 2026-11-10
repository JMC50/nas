<script lang="ts">
  import { onMount } from "svelte";
  import Info from "lucide-svelte/icons/info";

  interface Health {
    status: string;
    db: string;
    schema: string;
  }

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

<section class="flex flex-col gap-3">
  <div class="flex items-center gap-2">
    <Info size="18" class="text-fg-secondary" />
    <h2 class="sg-section-title">About</h2>
  </div>

  <div class="sg-panel p-4">
    <dl class="grid grid-cols-[120px_minmax(0,1fr)] gap-x-4 gap-y-2 text-sm">
      <dt class="text-fg-muted">App version</dt>
      <dd class="font-mono text-fg-primary">{__APP_VERSION__}</dd>

      <dt class="text-fg-muted">Backend</dt>
      <dd class="text-fg-primary">Go server</dd>

      {#if health}
        <dt class="text-fg-muted">Health</dt>
        <dd class="font-mono {statusColor(health.status)}">{health.status}</dd>

        <dt class="text-fg-muted">Database</dt>
        <dd class="font-mono {statusColor(health.db)}">{health.db}</dd>

        <dt class="text-fg-muted">Schema</dt>
        <dd class="font-mono {statusColor(health.schema)}">{health.schema}</dd>
      {:else}
        <dt class="text-fg-muted">Health</dt>
        <dd class="font-mono text-fg-muted">checking…</dd>
      {/if}
    </dl>
  </div>
</section>
