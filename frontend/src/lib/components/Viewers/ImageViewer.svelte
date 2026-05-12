<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import ZoomIn from "lucide-svelte/icons/zoom-in";
  import ZoomOut from "lucide-svelte/icons/zoom-out";
  import RefreshCcw from "lucide-svelte/icons/refresh-ccw";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let zoom = $state(1);

  const imageUrl = $derived(
    `/server/getImageData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  function zoomIn() {
    zoom = Math.min(8, zoom + 0.25);
  }
  function zoomOut() {
    zoom = Math.max(0.1, zoom - 0.25);
  }
  function zoomReset() {
    zoom = 1;
  }
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="flex items-center gap-2 px-3 h-9 border-b border-border-default text-xs text-fg-muted">
    <button
      type="button"
      class="inline-flex items-center justify-center w-7 h-7 rounded hover:bg-bg-hover"
      onclick={zoomOut}
      aria-label="Zoom out"
    >
      <ZoomOut size="14" />
    </button>
    <span class="font-mono w-12 text-center">{Math.round(zoom * 100)}%</span>
    <button
      type="button"
      class="inline-flex items-center justify-center w-7 h-7 rounded hover:bg-bg-hover"
      onclick={zoomIn}
      aria-label="Zoom in"
    >
      <ZoomIn size="14" />
    </button>
    <button
      type="button"
      class="inline-flex items-center justify-center w-7 h-7 rounded hover:bg-bg-hover"
      onclick={zoomReset}
      aria-label="Reset zoom"
    >
      <RefreshCcw size="14" />
    </button>
    <span class="ml-auto truncate text-fg-secondary">{name}</span>
  </div>
  <div class="flex-1 overflow-auto flex items-center justify-center">
    <img
      src={imageUrl}
      alt={name}
      style="transform: scale({zoom}); transform-origin: center; transition: transform 100ms ease;"
      class="max-w-none"
    />
  </div>
</div>
