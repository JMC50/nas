<script lang="ts">
  import Settings from "lucide-svelte/icons/settings";
  import Appearance from "$lib/components/Settings/Appearance.svelte";
  import Security from "$lib/components/Settings/Security.svelte";
  import About from "$lib/components/Settings/About.svelte";
  import { auth } from "$lib/store/auth.svelte";

  // Best-effort: only show password change for local accounts. Backend gates the call regardless.
  // (auth.svelte.ts doesn't expose authType; rely on backend response for the actual check.)
  const showSecurity = $derived(auth.isAuthenticated);
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
      <About />
    </div>
  </div>
</section>
