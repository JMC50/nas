<!-- frontend/src/lib/components/Viewers/VideoPlayer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let video: HTMLVideoElement | null = $state(null);

  const mediaUrl = $derived(
    `/server/getVideoData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<div class="relative h-full w-full bg-black overflow-hidden group">
  <!-- top bar (filename) -->
  <div
    class="absolute top-0 left-0 right-0 h-9 px-4 flex items-center
           bg-gradient-to-b from-bg-overlay/80 to-transparent
           text-fg-secondary text-xs font-mono z-10 pointer-events-none"
  >
    <span class="truncate">{name}</span>
  </div>

  <!-- video element -->
  <!-- svelte-ignore a11y_media_has_caption -->
  <video
    bind:this={video}
    src={mediaUrl}
    class="w-full h-full object-contain"
    preload="metadata"
  ></video>
</div>
