<!-- frontend/src/lib/components/Viewers/VideoPlayer.svelte -->
<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import { clampVolume, pickMediaKey } from "./media-utils";
  import Play from "lucide-svelte/icons/play";
  import Volume2 from "lucide-svelte/icons/volume-2";
  import Volume1 from "lucide-svelte/icons/volume-1";
  import VolumeX from "lucide-svelte/icons/volume-x";
  import VideoControls from "./VideoControls.svelte";

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
  // Set when the browser fails to decode the source — typically AVI/MKV/MOV
  // containers with codecs Chromium can't handle natively. Fallback UI offers
  // a direct download instead of a silent black box.
  let loadError = $state(false);

  function onMediaError() {
    loadError = true;
  }

  const downloadUrl = $derived(
    `/server/download?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

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

  let container: HTMLDivElement | null = $state(null);
  let isFullscreen = $state(false);

  function toggleFullscreen() {
    if (!container) return;
    if (document.fullscreenElement === container) {
      document.exitFullscreen().catch(() => {});
    } else {
      container.requestFullscreen().catch(() => {});
    }
  }

  function onFullscreenChange() {
    isFullscreen = document.fullscreenElement === container;
  }

  async function togglePip() {
    if (!video) return;
    try {
      if (document.pictureInPictureElement === video) {
        await document.exitPictureInPicture();
      } else {
        await video.requestPictureInPicture();
      }
    } catch {
      /* user dismissed or browser blocked — no-op */
    }
  }

  const pipSupported = $derived(
    typeof document !== "undefined" && document.pictureInPictureEnabled,
  );

  $effect(() => {
    document.addEventListener("fullscreenchange", onFullscreenChange);
    return () => document.removeEventListener("fullscreenchange", onFullscreenChange);
  });

  const SPEEDS = [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];

  let speed = $state(1);
  let speedMenuOpen = $state(false);

  function setSpeed(rate: number) {
    if (!video) return;
    video.playbackRate = rate;
    speed = rate;
    speedMenuOpen = false;
  }

  function onRateChange() {
    if (!video) return;
    speed = video.playbackRate;
  }

  function onToggleSpeedMenu(e: MouseEvent) {
    e.stopPropagation();
    speedMenuOpen = !speedMenuOpen;
  }

  function closeMenu(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest("[data-speed-menu]")) speedMenuOpen = false;
  }

  $effect(() => {
    if (speedMenuOpen) {
      document.addEventListener("click", closeMenu);
      return () => document.removeEventListener("click", closeMenu);
    }
  });

  const SEEK_STEP = 5;
  const VOLUME_STEP = 0.1;

  function onKeyDown(event: KeyboardEvent) {
    if (!video) return;
    // Skip when the player isn't in the active tab — inactive tab content
    // stays mounted (display:none in TabContent), so offsetParent === null
    // there. Without this guard, keys would fire in every mounted player.
    if (!container || container.offsetParent === null) return;
    const target = event.target as HTMLElement;
    if (target.matches("input, textarea, [contenteditable='true']")) return;

    const key = pickMediaKey(event);
    if (!key) return;
    event.preventDefault();

    switch (key) {
      case "toggle": toggle(); break;
      case "mute": toggleMute(); break;
      case "fullscreen": toggleFullscreen(); break;
      case "seekBack": video.currentTime = Math.max(0, video.currentTime - SEEK_STEP); break;
      case "seekForward": video.currentTime = Math.min(duration, video.currentTime + SEEK_STEP); break;
      case "volumeUp": video.volume = clampVolume(video.volume + VOLUME_STEP); video.muted = false; break;
      case "volumeDown": video.volume = clampVolume(video.volume - VOLUME_STEP); break;
    }
  }

  $effect(() => {
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  });

  const IDLE_HIDE_MS = 2500;

  let controlsVisible = $state(true);
  // plain let — not reactive, holds setTimeout handle only
  let idleTimer: number | null = null;

  function showControls() {
    controlsVisible = true;
    if (idleTimer !== null) window.clearTimeout(idleTimer);
    if (playing) {
      idleTimer = window.setTimeout(() => {
        controlsVisible = false;
      }, IDLE_HIDE_MS);
    }
  }

  $effect(() => {
    if (!playing) {
      controlsVisible = true;
      if (idleTimer !== null) window.clearTimeout(idleTimer);
    } else {
      showControls();
    }
    return () => {
      if (idleTimer !== null) window.clearTimeout(idleTimer);
    };
  });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  bind:this={container}
  onmousemove={showControls}
  class="relative h-full w-full bg-black overflow-hidden group
         {controlsVisible ? '' : 'cursor-none'}"
>
  <!-- top bar (filename) -->
  <div
    class="absolute top-0 left-0 right-0 h-9 px-4 flex items-center
           bg-gradient-to-b from-bg-overlay/80 to-transparent
           text-fg-secondary text-xs font-mono z-10 pointer-events-none
           transition-opacity duration-200 ease-smooth
           {controlsVisible ? 'opacity-100' : 'opacity-0'}"
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
    onratechange={onRateChange}
    onerror={onMediaError}
  ></video>

  {#if loadError}
    <div
      class="absolute inset-0 z-30 flex flex-col items-center justify-center gap-3 bg-bg-base/95 text-fg-primary px-6 text-center"
    >
      <div class="text-sm font-medium">Browser cannot play this format</div>
      <div class="text-xs text-fg-muted max-w-md">
        Container or codec is not supported natively (common for .avi, .mkv, .mov).
        Download to play in a local media player.
      </div>
      <a
        href={downloadUrl}
        download={name}
        class="inline-flex items-center gap-2 h-9 px-4 rounded-md bg-accent text-accent-fg text-sm font-medium hover:bg-accent/90 transition-colors"
      >
        Download {name}
      </a>
    </div>
  {/if}

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

  <VideoControls
    {controlsVisible}
    {playing}
    {currentTime}
    {duration}
    {playedRatio}
    {bufferedRatio}
    {volume}
    {muted}
    {VolumeIcon}
    {speed}
    {speedMenuOpen}
    {SPEEDS}
    {isFullscreen}
    {pipSupported}
    onToggle={toggle}
    onSeek={seek}
    onToggleMute={toggleMute}
    onSetVolume={setVolume}
    onSetSpeed={setSpeed}
    {onToggleSpeedMenu}
    onTogglePip={togglePip}
    onToggleFullscreen={toggleFullscreen}
  />
</div>
