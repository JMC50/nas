<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";
  import { auth } from "$lib/store/auth.svelte";

  const LINK_STATE_PREFIX = "link.";

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

  async function signInFlow() {
    try {
      const response = await fetch(`/server/auth/discord/callback?access_token=${encodeURIComponent(accessToken)}`);
      const data = await response.json();
      if (data.status === "new") {
        saveData = { userId: data.userId, username: data.username, global_name: data.global_name };
        registering = true;
      } else if (data.token) {
        useAuth.set({
          userId: data.userId,
          username: data.username,
          krname: data.krname ?? "",
          global_name: data.global_name,
          token: data.token,
        });
        window.location.replace(`${window.location.protocol}//${window.location.host}/`);
      } else {
        error = data.message ?? "Discord authentication failed.";
      }
    } catch (cause) {
      error = (cause as Error).message;
    }
  }

  async function linkFlow(state: string) {
    try {
      const response = await fetch("/server/auth/link/discord/complete", {
        method: "POST",
        headers: { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ state, access_token: accessToken }),
      });
      if (!response.ok) {
        error = `Link failed: ${await response.text()}`;
        return;
      }
      window.location.replace(`${window.location.protocol}//${window.location.host}/`);
    } catch (cause) {
      error = (cause as Error).message;
    }
  }

  onMount(async () => {
    const fragment = location.hash.startsWith("#") ? location.hash.slice(1) : location.search.slice(1);
    const params = new URLSearchParams(fragment);
    accessToken = params.get("access_token") ?? "";
    const state = params.get("state") ?? "";
    if (!accessToken) {
      error = "Missing access token from Discord.";
      return;
    }
    if (state.startsWith(LINK_STATE_PREFIX)) {
      await linkFlow(state);
      return;
    }
    await signInFlow();
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

<main class="min-h-full flex flex-col items-center justify-center px-gutter py-10">
  <div class="w-full max-w-[400px]">
    <div class="flex items-center gap-2 mb-5">
      <Logo size={26} />
      <span class="text-lg font-bold text-fg-primary">NAS</span>
    </div>

    <div class="sg-panel p-5 sm:p-6">
      {#if error}
        <div class="mb-4 p-3 rounded-control border border-fg-danger/45 text-fg-danger text-sm" role="alert">
          {error}
        </div>
      {/if}

      {#if !registering && !error}
        <div role="status">
          <h1 class="text-lg font-semibold text-fg-primary">Signing you in…</h1>
          <p class="text-sm text-fg-secondary mt-1">Verifying with Discord.</p>
        </div>
      {:else if registering}
        <h1 class="text-lg font-semibold text-fg-primary">One more step</h1>
        <p class="text-sm text-fg-secondary mt-1 mb-5">Enter your Korean name to finish registration.</p>

        <div class="flex flex-col gap-3 mb-5">
          <div>
            <label for="discord-first" class="sg-label">Username</label>
            <input id="discord-first" type="text" value={saveData.username} readonly class="sg-input" />
          </div>
          <div>
            <label for="discord-second" class="sg-label">Display name</label>
            <input id="discord-second" type="text" value={saveData.global_name} readonly class="sg-input" />
          </div>
          <div>
            <label for="krname" class="sg-label">Korean name</label>
            <input id="krname" type="text" bind:value={koreanName} placeholder="홍길동" autocomplete="name" class="sg-input" />
          </div>
        </div>

        <button type="button" class="sg-btn sg-btn-primary sg-btn-block" onclick={register}>
          Complete registration
        </button>
      {/if}
    </div>
  </div>
</main>
