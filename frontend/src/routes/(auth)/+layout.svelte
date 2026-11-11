<script lang="ts">
  let { children } = $props();
</script>

<div class="h-app w-screen bg-bg-base text-fg-primary overflow-auto">
  {@render children?.()}
</div>
