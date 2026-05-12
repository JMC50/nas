<script lang="ts">
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";

  interface UserData {
    userId: string;
    username: string;
    email: string;
    global_name: string;
  }

  let registering = $state(false);
  let saveData = $state<UserData>({ userId: "", username: "", email: "", global_name: "" });
  let koreanName = $state("");
  let error = $state("");

  onMount(async () => {
    const params = new URLSearchParams(location.search);
    const code = params.get("code");
    if (!code) {
      error = "Missing authorization code from Google.";
      return;
    }
    try {
      const response = await fetch(`/server/auth/google/callback?code=${encodeURIComponent(code)}`);
      const data = await response.json();
      if (data.status === "new") {
        saveData = {
          userId: data.userId,
          username: data.username ?? "",
          email: data.email ?? "",
          global_name: data.global_name ?? data.username ?? "",
        };
        registering = true;
      } else if (data.token) {
        useAuth.set({
          userId: data.userId,
          username: data.username ?? "",
          krname: data.krname ?? "",
          global_name: data.global_name ?? data.username ?? "",
          token: data.token,
        });
        const baseUrl = `${window.location.protocol}//${window.location.host}/`;
        window.location.replace(baseUrl);
      } else {
        error = data.message ?? "Google authentication failed.";
      }
    } catch (err) {
      error = (err as Error).message ?? "Network error.";
    }
  });

  async function completeRegistration() {
    if (!koreanName.trim()) {
      error = "Please enter your Korean name.";
      return;
    }
    try {
      const response = await fetch("/server/auth/google/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          userId: saveData.userId,
          username: saveData.username,
          email: saveData.email,
          global_name: saveData.global_name,
          krname: koreanName,
        }),
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
    } catch (err) {
      error = (err as Error).message ?? "Network error.";
    }
  }
</script>

<main class="min-h-screen flex items-center justify-center bg-bg-base p-6">
  <div class="w-full max-w-md bg-bg-surface border border-border-default rounded-lg p-6 shadow-[0_4px_12px_rgba(0,0,0,0.4)]">
    {#if error}
      <div class="mb-4 p-3 rounded-md bg-fg-danger/10 border border-fg-danger/30 text-fg-danger text-sm">
        {error}
      </div>
    {/if}

    {#if !registering && !error}
      <div class="text-lg text-fg-primary font-semibold mb-2">Signing you in…</div>
      <div class="text-sm text-fg-muted">Verifying with Google.</div>
    {:else if registering}
      <h1 class="text-xl text-fg-primary font-semibold mb-2">Welcome!</h1>
      <p class="text-sm text-fg-muted mb-4">Before logging in, please enter your Korean name.</p>

      <div class="space-y-3 mb-4">
        <div>
          <div class="text-xs text-fg-muted mb-1">Google ID</div>
          <input type="text" value={saveData.userId} readonly class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-disabled text-sm" />
        </div>
        <div>
          <div class="text-xs text-fg-muted mb-1">Email</div>
          <input type="text" value={saveData.email} readonly class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-disabled text-sm" />
        </div>
        <div>
          <div class="text-xs text-fg-muted mb-1">Name</div>
          <input type="text" value={saveData.global_name} readonly class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-disabled text-sm" />
        </div>
        <div>
          <label for="krname-input" class="block text-xs text-fg-muted mb-1">Korean Name</label>
          <input
            id="krname-input"
            type="text"
            bind:value={koreanName}
            placeholder="홍길동"
            class="w-full px-3 h-9 rounded-md bg-bg-elevated border border-border-default text-fg-primary text-sm focus:border-border-focus outline-none"
          />
        </div>
      </div>

      <button
        type="button"
        class="w-full h-10 rounded-md bg-accent text-accent-fg font-semibold hover:bg-accent-hover transition-colors"
        onclick={completeRegistration}
      >
        Complete Registration
      </button>
    {/if}
  </div>
</main>
