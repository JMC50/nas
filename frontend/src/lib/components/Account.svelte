<script lang="ts">
  import { onMount } from "svelte";
  import UserIcon from "lucide-svelte/icons/user";
  import LogOut from "lucide-svelte/icons/log-out";
  import Shield from "lucide-svelte/icons/shield";
  import KeyRound from "lucide-svelte/icons/key-round";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import SignInOptions from "$lib/components/Auth/SignInOptions.svelte";
  import type { Intent } from "$lib/types";

  const ALL_INTENTS: Intent[] = [
    "ADMIN",
    "VIEW",
    "OPEN",
    "DOWNLOAD",
    "UPLOAD",
    "COPY",
    "DELETE",
    "RENAME",
  ];

  interface AuthConfig {
    authType: "oauth" | "local" | "both";
    localAuthEnabled: boolean;
    discordEnabled: boolean;
    discordLoginUrl: string;
    googleEnabled: boolean;
    googleLoginUrl: string;
    oauthEnabled: boolean;
  }

  let authConfig: AuthConfig | null = $state(null);
  let intents: Intent[] = $state([]);
  let loadingIntents = $state(false);

  async function loadConfig() {
    try {
      const response = await fetch("/server/auth/config");
      authConfig = await response.json();
    } catch {
      authConfig = null;
    }
  }

  async function loadIntents() {
    if (!auth.current.userId) return;
    loadingIntents = true;
    try {
      const response = await fetch(`/server/getIntents?userId=${encodeURIComponent(auth.current.userId)}`);
      const data = await response.json();
      intents = data.intents ?? [];
    } finally {
      loadingIntents = false;
    }
  }

  function loginLocal() {
    location.href = "/localLogin";
  }

  function logout() {
    auth.clear();
    const baseUrl = `${window.location.protocol}//${window.location.host}/`;
    window.location.replace(baseUrl);
  }

  async function requestAdmin() {
    const password = prompt("Admin password?");
    if (!password) return;
    try {
      const response = await fetch(`/server/requestAdminIntent?token=${encodeURIComponent(auth.token)}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pwd: password }),
      });
      const text = await response.text();
      if (text === "complete") {
        notifications.success("Admin granted.");
        loadIntents();
      } else {
        notifications.error("Wrong password.");
      }
    } catch (cause) {
      notifications.error(`Request failed: ${(cause as Error).message}`);
    }
  }

  onMount(() => {
    loadConfig();
    if (auth.isAuthenticated) loadIntents();
  });
</script>

<section class="flex flex-col h-full bg-bg-base overflow-auto">
  <header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <UserIcon size="18" class="text-accent" />
    <h1 class="text-sm font-semibold text-fg-primary">Account</h1>
  </header>

  <div class="flex-1 p-6">
    {#if !auth.isAuthenticated}
      {#if !authConfig}
        <div class="text-xs text-fg-muted">Loading auth options…</div>
      {:else}
        <div class="max-w-md mx-auto mt-12 p-6 rounded-lg bg-bg-surface border border-border-default">
          <h2 class="text-base font-semibold text-fg-primary mb-1">Sign in</h2>
          <p class="text-xs text-fg-muted mb-6">Choose a provider to continue.</p>
          <SignInOptions
            discordEnabled={authConfig.discordEnabled}
            discordUrl={authConfig.discordLoginUrl}
            googleEnabled={authConfig.googleEnabled}
            googleUrl={authConfig.googleLoginUrl}
            localEnabled={authConfig.localAuthEnabled}
            onLocal={loginLocal}
          />
        </div>
      {/if}
    {:else}
      <div class="max-w-2xl">
        <div class="p-5 rounded-lg bg-bg-surface border border-border-default mb-4">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-12 h-12 rounded-full bg-bg-elevated border border-border-default flex items-center justify-center">
              <UserIcon size="20" class="text-fg-muted" />
            </div>
            <div class="min-w-0">
              <div class="text-base font-semibold text-fg-primary truncate">
                {auth.current.krname || auth.current.global_name || auth.current.username}
              </div>
              <div class="text-xs text-fg-muted truncate font-mono">{auth.current.userId}</div>
            </div>
          </div>

          <button
            type="button"
            class="inline-flex items-center gap-2 h-9 px-3 rounded-md bg-bg-elevated text-fg-primary text-xs font-medium hover:bg-bg-hover border border-border-default transition-colors"
            onclick={logout}
          >
            <LogOut size="13" />
            Sign out
          </button>
        </div>

        <div class="p-5 rounded-lg bg-bg-surface border border-border-default">
          <div class="flex items-center justify-between mb-4">
            <div class="flex items-center gap-2">
              <Shield size="14" class="text-fg-muted" />
              <h2 class="text-sm font-semibold text-fg-primary">Permissions</h2>
            </div>
            {#if !intents.includes("ADMIN")}
              <button
                type="button"
                class="inline-flex items-center gap-1.5 h-7 px-2.5 rounded text-xs text-fg-muted hover:text-fg-primary hover:bg-bg-hover transition-colors"
                onclick={requestAdmin}
              >
                <KeyRound size="11" />
                Request admin
              </button>
            {/if}
          </div>

          {#if loadingIntents}
            <div class="text-xs text-fg-muted">Loading…</div>
          {:else}
            <div class="flex flex-wrap gap-1.5">
              {#each ALL_INTENTS as intent}
                {@const granted = intents.includes(intent)}
                <span
                  class="inline-flex items-center px-2 h-6 rounded-md text-xs font-mono {granted
                    ? 'bg-accent/15 text-fg-accent border border-accent/30'
                    : 'bg-bg-elevated text-fg-disabled border border-border-default'}"
                >
                  {intent}
                </span>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</section>
