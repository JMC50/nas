<script lang="ts">
  import { onMount } from "svelte";
  import Link2 from "lucide-svelte/icons/link-2";
  import DiscordIcon from "$lib/components/Auth/DiscordIcon.svelte";
  import GoogleIcon from "$lib/components/Auth/GoogleIcon.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";

  interface IdentityEntry {
    provider: string;
    externalId: string;
  }

  interface ProviderRow {
    id: string;
    label: string;
    icon: typeof DiscordIcon;
  }

  const PROVIDERS: ProviderRow[] = [
    { id: "discord", label: "Discord", icon: DiscordIcon },
    { id: "google", label: "Google", icon: GoogleIcon },
  ];

  let identities = $state<IdentityEntry[]>([]);
  let busy = $state<Set<string>>(new Set());

  function markBusy(provider: string, value: boolean) {
    const next = new Set(busy);
    if (value) next.add(provider);
    else next.delete(provider);
    busy = next;
  }

  async function loadLinked() {
    if (!auth.token) return;
    try {
      const response = await fetch("/server/auth/identities", {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) return;
      const data = await response.json();
      identities = data.identities ?? [];
    } catch (cause) {
      console.warn("loadLinked failed", cause);
    }
  }

  async function connect(provider: string) {
    markBusy(provider, true);
    try {
      const response = await fetch("/server/auth/link/start", {
        method: "POST",
        headers: { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ provider }),
      });
      if (!response.ok) {
        notifications.error(`Connect ${provider} failed: ${await response.text()}`);
        return;
      }
      const data = await response.json();
      if (data.authorizeUrl) location.href = data.authorizeUrl;
    } catch (cause) {
      notifications.error(`Connect failed: ${(cause as Error).message}`);
    } finally {
      markBusy(provider, false);
    }
  }

  async function disconnect(provider: string) {
    if (!confirm(`Disconnect ${provider}? You will need another sign-in method to access this account.`)) return;
    markBusy(provider, true);
    try {
      const response = await fetch(`/server/auth/identities/${provider}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) {
        notifications.error(`Disconnect failed: ${await response.text()}`);
        return;
      }
      notifications.success(`${provider} disconnected.`);
      await loadLinked();
    } catch (cause) {
      notifications.error(`Disconnect failed: ${(cause as Error).message}`);
    } finally {
      markBusy(provider, false);
    }
  }

  function isLinked(provider: string): boolean {
    return identities.some((entry) => entry.provider === provider);
  }

  onMount(loadLinked);
</script>

<div class="p-5 rounded-lg bg-bg-surface border border-border-default">
  <div class="flex items-center gap-2 mb-4">
    <Link2 size="14" class="text-fg-muted" />
    <h2 class="text-sm font-semibold text-fg-primary">Linked accounts</h2>
  </div>

  <div class="divide-y divide-border-default/40">
    {#each PROVIDERS as provider (provider.id)}
      {@const Icon = provider.icon}
      {@const linked = isLinked(provider.id)}
      {@const pending = busy.has(provider.id)}
      <div class="flex items-center justify-between py-3 first:pt-0 last:pb-0">
        <div class="flex items-center gap-2.5">
          <Icon size={16} />
          <div>
            <div class="text-sm text-fg-primary">{provider.label}</div>
            <div class="text-xs text-fg-muted">
              {linked ? "Connected" : "Not connected"}
            </div>
          </div>
        </div>
        {#if linked}
          <button
            type="button"
            class="h-8 px-3 rounded-md bg-bg-elevated border border-border-default text-xs text-fg-primary hover:bg-bg-hover disabled:opacity-60 transition-colors"
            onclick={() => disconnect(provider.id)}
            disabled={pending}
          >
            {pending ? "..." : "Disconnect"}
          </button>
        {:else}
          <button
            type="button"
            class="h-8 px-3 rounded-md bg-accent text-accent-fg text-xs font-semibold hover:bg-accent-hover disabled:opacity-60 transition-colors"
            onclick={() => connect(provider.id)}
            disabled={pending}
          >
            {pending ? "..." : "Connect"}
          </button>
        {/if}
      </div>
    {/each}
  </div>
</div>
