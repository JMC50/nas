<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { auth } from "$lib/store/auth.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import { tabs } from "$lib/store/tabs.svelte";
  import Header from "$lib/components/Shell/Header.svelte";
  import VerticalNav from "$lib/components/Shell/VerticalNav.svelte";
  import MobileDrawer from "$lib/components/Shell/MobileDrawer.svelte";
  import StatusBar from "$lib/components/Shell/StatusBar.svelte";
  import Toasts from "$lib/components/Shell/Toasts.svelte";
  import DragDropOverlay from "$lib/components/Uploads/DragDropOverlay.svelte";
  import UploadPanel from "$lib/components/Uploads/UploadPanel.svelte";
  import QuickOpen from "$lib/components/QuickOpen.svelte";

  let { children } = $props();

  onMount(() => {
    if (!auth.isAuthenticated) {
      goto("/localLogin", { replaceState: true });
      return;
    }
    auth.refreshAdmin();
  });

  $effect(() => {
    if (typeof document === "undefined") return;
    const active = tabs.active;
    document.title = active ? `${active.title} · NAS` : "NAS";
  });
</script>

<div
  class="grid grid-rows-[48px_minmax(0,1fr)_auto] grid-cols-[minmax(0,1fr)] lg:grid-rows-[minmax(0,1fr)_auto] lg:grid-cols-[64px_minmax(0,1fr)] h-app w-screen bg-bg-base"
>
  <Header />
  <VerticalNav />
  <main class="row-start-2 col-start-1 lg:row-start-1 lg:col-start-2 min-w-0 min-h-0 overflow-hidden bg-bg-base">
    {@render children?.()}
  </main>
  <StatusBar />
</div>

<MobileDrawer />
<DragDropOverlay />
<UploadPanel open={ui.uploadsPanelOpen} onClose={() => ui.closeUploadsPanel()} />
<QuickOpen />
<Toasts />
