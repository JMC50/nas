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
  <div class="sg-viewer-bar">
    <button
      type="button"
      class="sg-btn sg-btn-ghost sg-btn-icon"
      onclick={zoomOut}
      aria-label="Zoom out"
    >
      <ZoomOut size="16" />
    </button>
    <span class="w-12 text-center tabular-nums">{Math.round(zoom * 100)}%</span>
    <button
      type="button"
      class="sg-btn sg-btn-ghost sg-btn-icon"
      onclick={zoomIn}
      aria-label="Zoom in"
    >
      <ZoomIn size="16" />
    </button>
    <button
      type="button"
      class="sg-btn sg-btn-ghost sg-btn-icon"
      onclick={zoomReset}
      aria-label="Reset zoom"
    >
      <RefreshCcw size="16" />
    </button>
    <span class="ml-auto min-w-0 truncate text-fg-primary font-semibold" title={name}>{name}</span>
  </div>
  <div class="flex-1 overflow-auto flex items-center justify-center">
    <img
      src={imageUrl}
      alt={name}
      style="transform: scale({zoom}); transform-origin: center;"
      class="max-w-none transition-transform duration-100 ease-snap"
    />
  </div>
</div>
