<script lang="ts">
  import Logo from "$lib/components/Logo.svelte";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import { onMount } from "svelte";
  import { useAuth } from "$lib/store/store";
  import { oidcAddresses } from "$lib/oidc";

  const FAILED = "The sign-in could not be verified. Sign in again.";
  const MESSAGES = new Map([
    ["denied", "Your account has no access to this NAS. Ask an administrator for access, then sign in again."],
    ["conflict", "A NAS account with the same ID already exists. Ask an administrator to link your account."],
    ["unavailable", "The sign-in service did not respond. Try again in a moment."],
  ]);

  let error = $state("");
  let denied = $state(false);
  let loginUrl = $state("/localLogin");
  let logoutUrl = $state("");

  async function takeSession() {
    try {
      const response = await fetch("/server/auth/oidc/session", { method: "POST" });
      const data = await response.json();
      if (!response.ok || !data.token) {
        error = data.message ?? FAILED;
        return;
      }
      useAuth.set({
        userId: data.userId,
        username: data.username ?? "",
        krname: data.krname ?? "",
        global_name: data.global_name ?? data.username ?? "",
        token: data.token,
      });
      window.location.replace(`${window.location.protocol}//${window.location.host}/`);
    } catch {
      error = FAILED;
    }
  }

  onMount(async () => {
    const problem = new URLSearchParams(location.search).get("error");
    if (problem === null) {
      await takeSession();
      if (!error) return;
    } else {
      denied = problem === "denied";
      error = MESSAGES.get(problem) ?? FAILED;
    }
    const oidc = await oidcAddresses();
    if (oidc) {
      loginUrl = oidc.loginUrl;
      logoutUrl = oidc.logoutUrl;
    }
  });
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
        <a href={loginUrl} class="mt-5 sg-btn sg-btn-primary sg-btn-block">Sign in again</a>
        {#if denied && logoutUrl}
          <a href={logoutUrl} class="mt-2 sg-btn sg-btn-block">Use another account</a>
        {/if}
      {:else}
        <div role="status">
          <h1 class="text-lg font-semibold text-fg-primary">Signing you in…</h1>
          <p class="text-sm text-fg-secondary mt-1">Finishing single sign-on.</p>
        </div>
      {/if}
    </div>
  </div>
</main>
