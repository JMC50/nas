<script lang="ts">
  import { onMount } from "svelte";
  import ServerIcon from "lucide-svelte/icons/server";
  import DiscordIcon from "$lib/components/Auth/DiscordIcon.svelte";
  import GoogleIcon from "$lib/components/Auth/GoogleIcon.svelte";
  import TextField from "$lib/components/Auth/TextField.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";

  interface ProviderView {
    clientId: string;
    redirectUri: string;
    hasSecret: boolean;
  }

  interface ConfigView {
    discord: ProviderView;
    google: ProviderView;
  }

  let view = $state<ConfigView | null>(null);
  let discordId = $state("");
  let discordSecret = $state("");
  let discordRedirect = $state("");
  let googleId = $state("");
  let googleSecret = $state("");
  let googleRedirect = $state("");
  let busy = $state(false);
  let origin = $state("");

  function bearer(): HeadersInit {
    return { Authorization: `Bearer ${auth.token}` };
  }

  function jsonHeaders(): HeadersInit {
    return { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` };
  }

  function applyView(next: ConfigView) {
    view = next;
    discordId = next.discord.clientId;
    discordRedirect = next.discord.redirectUri;
    googleId = next.google.clientId;
    googleRedirect = next.google.redirectUri;
  }

  async function loadView() {
    try {
      const response = await fetch("/server/admin/oauth-config", { headers: bearer() });
      if (!response.ok) {
        notifications.error("Failed to load OAuth config.");
        return;
      }
      applyView(await response.json());
    } catch (cause) {
      notifications.error(`Load failed: ${(cause as Error).message}`);
    }
  }

  function buildBody(): string {
    return JSON.stringify({
      discordClientId: discordId,
      discordClientSecret: discordSecret,
      discordRedirectUri: discordRedirect,
      googleClientId: googleId,
      googleClientSecret: googleSecret,
      googleRedirectUri: googleRedirect,
    });
  }

  async function pushConfig(): Promise<ConfigView | null> {
    const response = await fetch("/server/admin/oauth-config", {
      method: "PUT",
      headers: jsonHeaders(),
      body: buildBody(),
    });
    if (!response.ok) {
      notifications.error("Save failed.");
      return null;
    }
    return await response.json();
  }

  async function save() {
    busy = true;
    try {
      const next = await pushConfig();
      if (next) {
        applyView(next);
        discordSecret = "";
        googleSecret = "";
        notifications.success("OAuth configuration saved.");
      }
    } catch (cause) {
      notifications.error(`Save failed: ${(cause as Error).message}`);
    } finally {
      busy = false;
    }
  }

  const hints = $derived({
    discord: {
      secret: view && view.discord.hasSecret ? "••••••• (stored)" : "Client secret",
      redirect: origin ? `${origin}/login` : "https://your-domain/login",
    },
    google: {
      secret: view && view.google.hasSecret ? "••••••• (stored)" : "Client secret",
      redirect: origin ? `${origin}/googleLogin` : "https://your-domain/googleLogin",
    },
  });

  onMount(() => {
    origin = window.location.origin;
    loadView();
  });
</script>

<section class="space-y-3">
  <div class="flex items-center gap-2">
    <ServerIcon size="14" class="text-fg-muted" />
    <h2 class="text-xs font-semibold uppercase tracking-wide text-fg-muted">Server</h2>
  </div>

  <div class="rounded-lg bg-bg-surface border border-border-default p-4 space-y-3">
    <div class="flex items-center gap-2">
      <DiscordIcon size={14} />
      <div class="text-sm font-medium text-fg-primary">Discord OAuth</div>
    </div>
    <TextField id="discord-client-id" label="Client ID" value={discordId} disabled={busy} onInput={(next) => (discordId = next)} />
    <TextField id="discord-secret" label="Client secret" type="password" value={discordSecret} placeholder={hints.discord.secret} disabled={busy} onInput={(next) => (discordSecret = next)} />
    <TextField id="discord-redirect" label="Redirect URI" value={discordRedirect} placeholder={hints.discord.redirect} disabled={busy} onInput={(next) => (discordRedirect = next)} />
    <p class="text-[11px] text-fg-muted">Leave the secret empty to keep the stored value. Clear Client ID or Redirect URI to disable Discord sign-in.</p>
  </div>

  <div class="rounded-lg bg-bg-surface border border-border-default p-4 space-y-3">
    <div class="flex items-center gap-2">
      <GoogleIcon size={14} />
      <div class="text-sm font-medium text-fg-primary">Google OAuth</div>
    </div>
    <TextField id="google-client-id" label="Client ID" value={googleId} disabled={busy} onInput={(next) => (googleId = next)} />
    <TextField id="google-secret" label="Client secret" type="password" value={googleSecret} placeholder={hints.google.secret} disabled={busy} onInput={(next) => (googleSecret = next)} />
    <TextField id="google-redirect" label="Redirect URI" value={googleRedirect} placeholder={hints.google.redirect} disabled={busy} onInput={(next) => (googleRedirect = next)} />
    <p class="text-[11px] text-fg-muted">Leave the secret empty to keep the stored value. Google sign-in requires all three fields.</p>
  </div>

  <button
    type="button"
    class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-accent text-accent-fg text-xs font-semibold hover:bg-accent-hover disabled:opacity-60 transition-colors"
    onclick={save}
    disabled={busy}
  >
    {busy ? "Saving…" : "Save OAuth configuration"}
  </button>
</section>
