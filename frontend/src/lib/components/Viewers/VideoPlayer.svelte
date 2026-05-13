<!-- frontend/src/lib/components/Viewers/VideoPlayer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import { formatTime, clampVolume } from "./media-utils";
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import Volume2 from "lucide-svelte/icons/volume-2";
  import Volume1 from "lucide-svelte/icons/volume-1";
  import VolumeX from "lucide-svelte/icons/volume-x";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let video: HTMLVideoElement | null = $state(null);
  let playing = $state(false);
  let currentTime = $state(0);
  let duration = $state(0);
  let buffered = $state(0);

  function onTimeUpdate() {
    if (!video) return;
    currentTime = video.currentTime;
  }
  function onLoadedMetadata() {
    if (!video) return;
    duration = video.duration;
  }
  function onProgress() {
    if (!video) return;
    const ranges = video.buffered;
    buffered = ranges.length > 0 ? ranges.end(ranges.length - 1) : 0;
  }

  function seek(event: MouseEvent) {
    if (!video || duration === 0) return;
    const bar = event.currentTarget as HTMLElement;
    const rect = bar.getBoundingClientRect();
    const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
    video.currentTime = ratio * duration;
    currentTime = video.currentTime;
  }

  const playedRatio = $derived(duration > 0 ? currentTime / duration : 0);
  const bufferedRatio = $derived(duration > 0 ? buffered / duration : 0);

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

  let volume = $state(1);
  let muted = $state(false);

  function onVolumeChange() {
    if (!video) return;
    volume = video.volume;
    muted = video.muted;
  }

  function toggleMute() {
    if (!video) return;
    video.muted = !video.muted;
  }

  function setVolume(event: MouseEvent) {
    if (!video) return;
    const bar = event.currentTarget as HTMLElement;
    const rect = bar.getBoundingClientRect();
    const ratio = clampVolume((event.clientX - rect.left) / rect.width);
    video.volume = ratio;
    if (ratio > 0) video.muted = false;
  }

  const VolumeIcon = $derived(
    muted || volume === 0 ? VolumeX : volume < 0.5 ? Volume1 : Volume2,
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
    ontimeupdate={onTimeUpdate}
    onloadedmetadata={onLoadedMetadata}
    onprogress={onProgress}
    onvolumechange={onVolumeChange}
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

  <!-- bottom controls bar -->
  <div
    class="absolute bottom-0 left-0 right-0 px-4 pb-2 pt-3 flex flex-col gap-1
           bg-bg-overlay/80 backdrop-blur-sm
           text-fg-primary z-10"
  >
    <!-- scrubber -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      onclick={seek}
      class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer group/scrub
             hover:h-2 transition-all"
    >
      <!-- buffered -->
      <div
        class="absolute top-0 left-0 h-full rounded-full bg-bg-selected"
        style="width: {bufferedRatio * 100}%"
      ></div>
      <!-- played -->
      <div
        class="absolute top-0 left-0 h-full rounded-full bg-accent"
        style="width: {playedRatio * 100}%"
      ></div>
      <!-- thumb -->
      <div
        class="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 w-2.5 h-2.5 rounded-full bg-accent
               opacity-0 group-hover/scrub:opacity-100 transition-opacity"
        style="left: {playedRatio * 100}%"
      ></div>
    </div>

    <!-- controls row -->
    <div class="flex items-center gap-2 h-9">
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

      <span class="font-mono text-xs text-fg-secondary tabular-nums">
        {formatTime(currentTime)} / {formatTime(duration)}
      </span>

      <!-- volume control (button + slide-out slider) -->
      <div class="group/vol flex items-center">
        <button
          type="button"
          onclick={toggleMute}
          aria-label={muted ? "Unmute" : "Mute"}
          class="inline-flex items-center justify-center w-9 h-9 rounded
                 hover:bg-bg-hover hover:text-fg-accent transition-colors"
        >
          <VolumeIcon size="18" />
        </button>
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          onclick={setVolume}
          class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer
                 w-0 group-hover/vol:w-20 ml-0 group-hover/vol:ml-2
                 transition-all overflow-hidden"
        >
          <div
            class="absolute top-0 left-0 h-full rounded-full bg-fg-secondary"
            style="width: {muted ? 0 : volume * 100}%"
          ></div>
        </div>
      </div>

      <!-- spacer; later tasks fill the right side -->
      <div class="ml-auto"></div>
    </div>
  </div>
</div>
