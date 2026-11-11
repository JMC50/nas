<script lang="ts">
  import { tick } from "svelte";
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
  let natural = $state<{ width: number; height: number } | null>(null);
  let animate = $state(false);
  let viewport: HTMLDivElement | undefined = $state();

  const imageUrl = $derived(
    `/server/getImageData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  function zoomIn() {
    zoom = Math.min(8, zoom + 0.25);
  }
  function zoomOut() {
    zoom = Math.max(0.1, zoom - 0.25);
  }
  function fitZoom(): number {
    if (!natural || !viewport || viewport.clientWidth === 0) return 1;
    return Math.min(1, viewport.clientWidth / natural.width, viewport.clientHeight / natural.height);
  }
  function zoomReset() {
    zoom = fitZoom();
  }
  async function onLoad(event: Event) {
    const image = event.currentTarget as HTMLImageElement;
    natural = image.naturalWidth > 0 ? { width: image.naturalWidth, height: image.naturalHeight } : null;
    zoom = fitZoom();
    // 첫 배율은 움직임 없이 적용하려고 스타일을 한 번 계산시킨 뒤 전환을 켬
    await tick();
    void getComputedStyle(image).transform;
    animate = true;
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
    <span class="w-12 shrink-0 text-center tabular-nums">{Math.round(zoom * 100)}%</span>
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
  <!-- 가운데 정렬을 m-auto로 해야 화면보다 커졌을 때 왼쪽과 위쪽도 스크롤로 닿음 -->
  <div bind:this={viewport} class="flex-1 overflow-auto flex">
    <div class="m-auto shrink-0" style={natural ? `width: ${natural.width * zoom}px; height: ${natural.height * zoom}px;` : ""}>
      <img
        src={imageUrl}
        alt={name}
        onload={onLoad}
        style="transform: scale({zoom}); transform-origin: 0 0;"
        class="max-w-none {animate ? 'transition-transform duration-100 ease-snap' : ''}"
      />
    </div>
  </div>
</div>
