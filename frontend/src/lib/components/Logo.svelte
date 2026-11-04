<script lang="ts">
  let { size = 20 }: { size?: number } = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  class="shrink-0 text-accent"
  aria-hidden="true"
>
  <path
    fill="currentColor"
    fill-rule="evenodd"
    d="M3 4h18v4H3zM16 5h2v2h-2zM3 10h18v4H3zM16 11h2v2h-2zM3 16h18v4H3zM16 17h2v2h-2z"
  />
</svg>
