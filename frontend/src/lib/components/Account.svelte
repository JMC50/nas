<script lang="ts">
  import { onMount } from "svelte";
  import UserIcon from "lucide-svelte/icons/user";
  import LogOut from "lucide-svelte/icons/log-out";
  import Shield from "lucide-svelte/icons/shield";
  import KeyRound from "lucide-svelte/icons/key-round";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { claimAdmin } from "$lib/auth-actions";
  import SignInOptions from "$lib/components/Auth/SignInOptions.svelte";
  import LinkedAccounts from "$lib/components/LinkedAccounts.svelte";
  import { signOutAddress } from "$lib/oidc";
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
    authType: "oauth" | "local" | "both" | "oidc";
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
      const response = await fetch(`/server/getIntents?userId=${encodeURIComponent(auth.current.userId)}`, {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const data = await response.json();
      intents = data.intents ?? [];
    } catch (cause) {
      notifications.error(`Failed to load permissions: ${(cause as Error).message}`);
    } finally {
      loadingIntents = false;
    }
  }

  function loginLocal() {
    location.href = "/localLogin";
  }

  async function logout() {
    auth.clear();
    const baseUrl = `${window.location.protocol}//${window.location.host}/`;
    window.location.replace(await signOutAddress(baseUrl));
  }

  async function requestAdmin() {
    const password = prompt("Admin password?");
    if (!password) return;
    const result = await claimAdmin(password);
    if (result.success) {
      notifications.success("Admin granted.");
      loadIntents();
    } else {
      notifications.error(result.message ?? "Request failed.");
    }
  }

  onMount(() => {
    loadConfig();
    if (auth.isAuthenticated) loadIntents();
  });
</script>

<section class="flex flex-col h-full bg-bg-base overflow-auto">
  <header class="sg-page-head">
    <UserIcon size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">Account</h1>
  </header>

  <div class="flex-1 px-gutter py-6">
    {#if !auth.isAuthenticated}
      {#if !authConfig}
        <div class="text-sm text-fg-muted" role="status">Loading auth options…</div>
      {:else}
        <div class="max-w-md mx-auto mt-12 p-6 sg-panel">
          <h2 class="text-lg font-semibold text-fg-primary">Sign in</h2>
          <p class="text-sm text-fg-secondary mt-1 mb-6">Choose a provider to continue.</p>
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
      <div class="max-w-2xl flex flex-col gap-4">
        <div class="sg-panel p-4 flex flex-wrap items-center gap-3">
          <span class="grid place-items-center size-12 shrink-0 rounded-control bg-fg-info text-accent-fg text-lg font-bold" aria-hidden="true">
            {(auth.current.krname || auth.current.global_name || auth.current.username || "?").slice(0, 1).toUpperCase()}
          </span>
          <div class="min-w-0 flex-1">
            <div class="text-md font-semibold text-fg-primary truncate">
              {auth.current.krname || auth.current.global_name || auth.current.username}
            </div>
            <div class="text-xs text-fg-muted truncate font-mono">{auth.current.userId}</div>
          </div>
          <button type="button" class="sg-btn" onclick={logout}>
            <LogOut size="16" />
            Sign out
          </button>
        </div>

        {#if authConfig?.authType !== "oidc"}
          <LinkedAccounts />
        {/if}

        <div class="sg-panel p-4">
          <div class="flex items-center justify-between gap-3 mb-4">
            <div class="flex items-center gap-2">
              <Shield size="18" class="text-fg-secondary" />
              <h2 class="text-sm font-semibold text-fg-primary">Permissions</h2>
            </div>
            {#if !intents.includes("ADMIN")}
              <button type="button" class="sg-btn sg-btn-ghost" onclick={requestAdmin}>
                <KeyRound size="16" />
                Request admin
              </button>
            {/if}
          </div>

          {#if loadingIntents}
            <div class="text-sm text-fg-muted" role="status">Loading…</div>
          {:else}
            <ul class="flex flex-wrap gap-1.5" aria-label="Permissions">
              {#each ALL_INTENTS as intent}
                {@const granted = intents.includes(intent)}
                <li class="sg-chip font-mono {granted ? 'text-fg-success' : 'text-fg-muted'}">
                  <span class="size-1.5 rounded-full {granted ? 'bg-fg-success' : 'border border-fg-muted'}" aria-hidden="true"></span>
                  {intent}
                  <span class="sr-only">{granted ? "granted" : "not granted"}</span>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </div>
    {/if}
  </div>
</section>
