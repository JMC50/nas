<script lang="ts">
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";
  import HardDrive from "lucide-svelte/icons/hard-drive";

  interface DiscordUser {
    userId: string;
    username: string;
    global_name: string;
  }

  let registering = $state(false);
  let saveData = $state<DiscordUser>({ userId: "", username: "", global_name: "" });
  let koreanName = $state("");
  let accessToken = "";
  let error = $state("");

  onMount(async () => {
    const fragment = location.hash.startsWith("#") ? location.hash.slice(1) : location.search.slice(1);
    const params = new URLSearchParams(fragment);
    accessToken = params.get("access_token") ?? "";
    if (!accessToken) {
      error = "Missing access token from Discord.";
      return;
    }
    try {
      const response = await fetch(`/server/auth/discord/callback?access_token=${encodeURIComponent(accessToken)}`);
      const data = await response.json();
      if (data.status === "new") {
        saveData = {
          userId: data.userId,
          username: data.username,
          global_name: data.global_name,
        };
        registering = true;
      } else if (data.token) {
        useAuth.set({
          userId: data.userId,
          username: data.username,
          krname: data.krname ?? "",
          global_name: data.global_name,
          token: data.token,
        });
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
      } else {
        error = data.message ?? "Discord authentication failed.";
      }
    } catch (cause) {
      error = (cause as Error).message;
    }
  });

  async function register() {
    if (!koreanName.trim()) {
      error = "Please enter your Korean name.";
      return;
    }
    try {
      const response = await fetch("/server/auth/discord/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ access_token: accessToken, krname: koreanName }),
      });
      const data = await response.json();
      if (data.status === "complete") {
        useAuth.set({
          userId: data.userId,
          username: data.username,
          krname: koreanName,
          global_name: data.global_name,
          token: data.token,
        });
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
      } else {
        error = data.message ?? "Registration failed.";
      }
    } catch (cause) {
      error = (cause as Error).message;
    }
  }
</script>

<main class="min-h-screen flex items-center justify-center bg-bg-base p-6">
  <div class="w-full max-w-md bg-bg-surface border border-border-default rounded-lg p-6 shadow-[0_4px_12px_rgba(0,0,0,0.4)]">
    <div class="flex items-center gap-2 mb-6">
      <HardDrive size="20" class="text-accent" />
      <span class="text-base font-semibold text-fg-primary tracking-tight">NAS</span>
    </div>

    {#if error}
      <div class="mb-4 p-3 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-sm">
        {error}
      </div>
    {/if}

    {#if !registering && !error}
      <div class="text-sm font-semibold text-fg-primary mb-1">Signing you in…</div>
      <div class="text-xs text-fg-muted">Verifying with Discord.</div>
    {:else if registering}
      <h1 class="text-base font-semibold text-fg-primary mb-1">One more step</h1>
      <p class="text-xs text-fg-muted mb-4">Enter your Korean name to finish registration.</p>

      <div class="space-y-3 mb-4">
        <div>
          <div class="text-[11px] text-fg-muted mb-1">Username</div>
          <input type="text" value={saveData.username} readonly class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-disabled text-sm" />
        </div>
        <div>
          <div class="text-[11px] text-fg-muted mb-1">Display name</div>
          <input type="text" value={saveData.global_name} readonly class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-disabled text-sm" />
        </div>
        <div>
          <label for="krname" class="block text-[11px] text-fg-muted mb-1">Korean name</label>
          <input
            id="krname"
            type="text"
            bind:value={koreanName}
            placeholder="홍길동"
            class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none"
          />
        </div>
      </div>

      <button
        type="button"
        class="w-full h-10 rounded-md bg-accent text-accent-fg text-sm font-semibold hover:bg-accent-hover transition-colors"
        onclick={register}
      >
        Complete registration
      </button>
    {/if}
  </div>
</main>
