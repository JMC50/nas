<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import TabBar from "$lib/components/Tabs/TabBar.svelte";
  import TabContent from "$lib/components/Tabs/TabContent.svelte";
  import { installShortcuts, uninstallShortcuts } from "$lib/keyboard";

  onMount(() => installShortcuts());
  onDestroy(() => uninstallShortcuts());
</script>

<div class="grid grid-rows-[36px_1fr] h-full w-full">
  <TabBar />
  <div class="min-h-0 overflow-hidden">
    <TabContent />
  </div>
</div>
