<!-- frontend/src/lib/components/Viewers/VideoPlayer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let video: HTMLVideoElement | null = $state(null);
  let playing = $state(false);

  const mediaUrl = $derived(
    `/server/getVideoData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  function toggle() {
    if (!video) return;
    if (video.paused) video.play().catch(() => {});
    else video.pause();
  }

  function onPlay() { playing = true; }
  function onPause() { playing = false; }
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
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <video
    bind:this={video}
    src={mediaUrl}
    class="w-full h-full object-contain cursor-pointer"
    preload="metadata"
    onclick={toggle}
    onplay={onPlay}
    onpause={onPause}
  ></video>

  <!-- center overlay (visible when paused) -->
  {#if !playing}
    <button
      type="button"
      onclick={toggle}
      aria-label="Play"
      class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2
             w-24 h-24 rounded-full bg-bg-overlay/70 backdrop-blur-sm
             text-fg-primary flex items-center justify-center
             hover:bg-bg-overlay/90 hover:text-fg-accent transition-colors z-20"
    >
      <Play size="40" fill="currentColor" />
    </button>
  {/if}

  <!-- bottom controls bar (placeholder — extends in later tasks) -->
  <div
    class="absolute bottom-0 left-0 right-0 h-14 px-4 flex items-center gap-2
           bg-bg-overlay/80 backdrop-blur-sm
           text-fg-primary z-10"
  >
    <button
      type="button"
      onclick={toggle}
      aria-label={playing ? "Pause" : "Play"}
      class="inline-flex items-center justify-center w-9 h-9 rounded
             hover:bg-bg-hover hover:text-fg-accent transition-colors"
    >
      {#if playing}
        <Pause size="18" />
      {:else}
        <Play size="18" />
      {/if}
    </button>
  </div>
</div>
