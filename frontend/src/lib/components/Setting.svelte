<script lang="ts">
  import { onMount } from "svelte";
  import Settings from "lucide-svelte/icons/settings";
  import Appearance from "$lib/components/Settings/Appearance.svelte";
  import Security from "$lib/components/Settings/Security.svelte";
  import ServerSection from "$lib/components/Settings/Server.svelte";
  import About from "$lib/components/Settings/About.svelte";
  import { auth } from "$lib/store/auth.svelte";
  import { oidcAddresses } from "$lib/oidc";

  let isAdmin = $state(false);
  let passwordless = $state(false);
  const showSecurity = $derived(auth.isAuthenticated && !passwordless);

  async function checkAdmin() {
    if (!auth.token) {
      isAdmin = false;
      return;
    }
    try {
      const response = await fetch(`/server/checkAdmin?token=${encodeURIComponent(auth.token)}`);
      if (!response.ok) {
        isAdmin = false;
        return;
      }
      const data = await response.json();
      isAdmin = data.isAdmin === true;
    } catch {
      isAdmin = false;
    }
  }

  onMount(async () => {
    checkAdmin();
    passwordless = (await oidcAddresses()) !== null;
  });
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="sg-page-head">
    <Settings size="18" class="text-fg-secondary" />
    <h1 class="sg-page-title">Settings</h1>
  </header>

  <div class="flex-1 overflow-auto px-gutter py-6">
    <div class="max-w-2xl flex flex-col gap-10">
      <Appearance />
      {#if showSecurity}
        <Security />
      {/if}
      {#if isAdmin}
        <ServerSection />
      {/if}
      <About />
    </div>
  </div>
</section>
