<script lang="ts">
  import { onMount } from "svelte";
  import Settings from "lucide-svelte/icons/settings";
  import Appearance from "$lib/components/Settings/Appearance.svelte";
  import Security from "$lib/components/Settings/Security.svelte";
  import ServerSection from "$lib/components/Settings/Server.svelte";
  import About from "$lib/components/Settings/About.svelte";
  import { auth } from "$lib/store/auth.svelte";

  let isAdmin = $state(false);
  const showSecurity = $derived(auth.isAuthenticated);

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

  onMount(checkAdmin);
</script>

<section class="flex flex-col h-full bg-bg-base overflow-hidden">
  <header class="flex items-center gap-2 px-6 h-12 border-b border-border-default bg-bg-surface">
    <Settings size="18" class="text-accent" />
    <h1 class="text-sm font-semibold text-fg-primary">Settings</h1>
  </header>

  <div class="flex-1 overflow-auto p-6">
    <div class="max-w-2xl space-y-6">
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
