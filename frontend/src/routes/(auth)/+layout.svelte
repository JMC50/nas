<script lang="ts">
  let { children } = $props();
</script>

<div class="h-screen w-screen bg-bg-base text-fg-primary overflow-auto">
  {@render children?.()}
</div>
