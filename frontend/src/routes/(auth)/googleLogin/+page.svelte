<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";
  import { auth } from "$lib/store/auth.svelte";

  const LINK_STATE_PREFIX = "link.";

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

  async function signInFlow(code: string) {
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
        window.location.replace(`${window.location.protocol}//${window.location.host}/`);
      } else {
        error = data.message ?? "Google authentication failed.";
      }
    } catch (cause) {
      error = (cause as Error).message ?? "Network error.";
    }
  }

  async function linkFlow(state: string, code: string) {
    try {
      const response = await fetch("/server/auth/link/google/complete", {
        method: "POST",
        headers: { Authorization: `Bearer ${auth.token}`, "Content-Type": "application/json" },
        body: JSON.stringify({ state, code }),
      });
      if (!response.ok) {
        error = `Link failed: ${await response.text()}`;
        return;
      }
      window.location.replace(`${window.location.protocol}//${window.location.host}/`);
    } catch (cause) {
      error = (cause as Error).message ?? "Network error.";
    }
  }

  onMount(async () => {
    const params = new URLSearchParams(location.search);
    const code = params.get("code");
    const state = params.get("state") ?? "";
    if (!code) {
      error = "Missing authorization code from Google.";
      return;
    }
    if (state.startsWith(LINK_STATE_PREFIX)) {
      await linkFlow(state, code);
      return;
    }
    await signInFlow(code);
  });

  async function register() {
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
    } catch (cause) {
      error = (cause as Error).message ?? "Network error.";
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
        <p class="flex items-start gap-2 text-sm text-fg-danger" role="alert">
          <AlertCircle size="16" class="mt-0.5 shrink-0" />
          <span>{error}</span>
        </p>
        {#if !registering}
          <a href="/localLogin" class="mt-5 sg-btn sg-btn-block">Back to sign in</a>
        {/if}
      {/if}

      {#if !registering && !error}
        <div role="status">
          <h1 class="text-lg font-semibold text-fg-primary">Signing you in…</h1>
          <p class="text-sm text-fg-secondary mt-1">Verifying with Google.</p>
        </div>
      {:else if registering}
        <h1 class="text-lg font-semibold text-fg-primary {error ? 'mt-5' : ''}">Welcome!</h1>
        <p class="text-sm text-fg-secondary mt-1 mb-5">Before logging in, please enter your Korean name.</p>

        <div class="flex flex-col gap-3 mb-5">
          <div>
            <label for="google-first" class="sg-label">Email</label>
            <input id="google-first" type="text" value={saveData.email} readonly class="sg-input sg-input-form" />
          </div>
          <div>
            <label for="google-second" class="sg-label">Name</label>
            <input id="google-second" type="text" value={saveData.global_name} readonly class="sg-input sg-input-form" />
          </div>
          <div>
            <label for="krname" class="sg-label">Korean name</label>
            <input id="krname" type="text" bind:value={koreanName} placeholder="홍길동" autocomplete="name" class="sg-input sg-input-form" />
          </div>
        </div>

        <button type="button" class="sg-btn sg-btn-primary sg-btn-block" onclick={register}>
          Complete Registration
        </button>
      {/if}
    </div>
  </div>
</main>
