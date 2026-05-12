<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/store/auth.svelte";
  import Header from "$lib/components/Shell/Header.svelte";
  import VerticalNav from "$lib/components/Shell/VerticalNav.svelte";
  import StatusBar from "$lib/components/Shell/StatusBar.svelte";

  let { children } = $props();

  onMount(() => {
    if (!auth.isAuthenticated) {
      goto("/localLogin", { replaceState: true });
    }
  });
</script>

<div class="grid grid-rows-[48px_1fr_28px] grid-cols-[auto_1fr] h-screen w-screen bg-bg-base">
  <Header />
  <VerticalNav />
  <main class="row-start-2 col-start-2 min-w-0 min-h-0 overflow-hidden bg-bg-base">
    {@render children?.()}
  </main>
  <StatusBar />
</div>
