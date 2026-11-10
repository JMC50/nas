<!-- frontend/src/lib/components/Viewers/AudioPlayer.svelte -->
<script lang="ts">
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import Volume2 from "lucide-svelte/icons/volume-2";
  import Volume1 from "lucide-svelte/icons/volume-1";
  import VolumeX from "lucide-svelte/icons/volume-x";
  import { auth } from "$lib/store/auth.svelte";
  import { formatTime, clampVolume, pickMediaKey } from "./media-utils";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let audio: HTMLAudioElement | null = $state(null);
  let playerRoot: HTMLDivElement | null = $state(null);
  let playing = $state(false);
  let currentTime = $state(0);
  let duration = $state(0);
  let buffered = $state(0);
  let volume = $state(1);
  let muted = $state(false);

  const mediaUrl = $derived(
    `/server/getAudioData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  const playedRatio = $derived(duration > 0 ? currentTime / duration : 0);
  const bufferedRatio = $derived(duration > 0 ? buffered / duration : 0);

  const VolumeIcon = $derived(
    muted || volume === 0 ? VolumeX : volume < 0.5 ? Volume1 : Volume2,
  );

  function toggle() {
    if (!audio) return;
    if (audio.paused) audio.play().catch(() => {});
    else audio.pause();
  }

  function onPlay() { playing = true; }
  function onPause() { playing = false; }

  function onTimeUpdate() {
    if (!audio) return;
    currentTime = audio.currentTime;
  }
  function onLoadedMetadata() {
    if (!audio) return;
    duration = audio.duration;
  }
  function onProgress() {
    if (!audio) return;
    const ranges = audio.buffered;
    buffered = ranges.length > 0 ? ranges.end(ranges.length - 1) : 0;
  }
  function onVolumeChange() {
    if (!audio) return;
    volume = audio.volume;
    muted = audio.muted;
  }

  function seek(event: MouseEvent) {
    if (!audio || duration === 0) return;
    const bar = event.currentTarget as HTMLElement;
    const rect = bar.getBoundingClientRect();
    const ratio = Math.max(0, Math.min(1, (event.clientX - rect.left) / rect.width));
    audio.currentTime = ratio * duration;
    currentTime = audio.currentTime;
  }

  function toggleMute() {
    if (!audio) return;
    audio.muted = !audio.muted;
  }

  function setVolume(event: MouseEvent) {
    if (!audio) return;
    const bar = event.currentTarget as HTMLElement;
    const rect = bar.getBoundingClientRect();
    const ratio = clampVolume((event.clientX - rect.left) / rect.width);
    audio.volume = ratio;
    if (ratio > 0) audio.muted = false;
  }

  const SEEK_STEP = 5;
  const VOLUME_STEP = 0.1;

  function onKeyDown(event: KeyboardEvent) {
    if (!audio) return;
    // Skip when the player isn't in the active tab — inactive tab content
    // stays mounted (display:none in TabContent), so offsetParent === null
    // there. Without this guard, keys would fire in every mounted player.
    if (!playerRoot || playerRoot.offsetParent === null) return;
    const target = event.target as HTMLElement;
    if (target.matches("input, textarea, [contenteditable='true']")) return;

    const key = pickMediaKey(event);
    if (!key || key === "fullscreen") return;
    event.preventDefault();

    switch (key) {
      case "toggle": toggle(); break;
      case "mute": toggleMute(); break;
      case "seekBack": audio.currentTime = Math.max(0, audio.currentTime - SEEK_STEP); break;
      case "seekForward": audio.currentTime = Math.min(duration, audio.currentTime + SEEK_STEP); break;
      case "volumeUp": audio.volume = clampVolume(audio.volume + VOLUME_STEP); audio.muted = false; break;
      case "volumeDown": audio.volume = clampVolume(audio.volume - VOLUME_STEP); break;
    }
  }

  $effect(() => {
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  });
</script>

<div bind:this={playerRoot} class="flex flex-col h-full w-full bg-bg-base items-center justify-center p-gutter">
  <div class="w-full max-w-[360px] flex flex-col items-center gap-6">
    <h2 class="text-lg font-semibold text-fg-primary truncate w-full text-center" title={name}>
      {name}
    </h2>

    <button
      type="button"
      onclick={toggle}
      aria-label={playing ? "Pause" : "Play"}
      class="size-24 rounded-panel bg-accent text-accent-fg
 flex items-center justify-center
 hover:bg-accent-hover active:bg-accent-muted transition-colors"
    >
      {#if playing}
        <Pause size="40" fill="currentColor" />
      {:else}
        <Play size="40" fill="currentColor" class="ml-1" />
      {/if}
    </button>

    <!-- equalizer -->
    <div class="flex items-end justify-center gap-1 h-8">
      {#each [1, 2, 3, 4, 5] as i (i)}
        <span
          class="eq-bar block w-1.5 h-full bg-accent origin-bottom"
          style="animation: equalize-{i} {0.5 + i * 0.1}s ease-in-out infinite;
                 animation-play-state: {playing ? 'running' : 'paused'};
                 transform: scaleY({playing ? 1 : 0.3});"
        ></span>
      {/each}
    </div>

    <!-- scrubber -->
    <div class="w-full flex flex-col gap-1">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={seek}
        class="relative h-2 rounded-full overflow-hidden bg-bg-overlay cursor-pointer"
      >
        <div
          class="absolute top-0 left-0 h-full bg-border-strong/50"
          style="width: {bufferedRatio * 100}%"
        ></div>
        <div
          class="absolute top-0 left-0 h-full bg-accent"
          style="width: {playedRatio * 100}%"
        ></div>
      </div>
      <div class="flex justify-between text-xs text-fg-secondary tabular-nums">
        <span>{formatTime(currentTime)}</span>
        <span>{formatTime(duration)}</span>
      </div>
    </div>

    <!-- volume row -->
    <div class="w-full flex items-center gap-2">
      <button
        type="button"
        onclick={toggleMute}
        aria-label={muted ? "Unmute" : "Mute"}
        class="sg-btn sg-btn-ghost sg-btn-icon"
      >
        <VolumeIcon size="16" />
      </button>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={setVolume}
        class="flex-1 relative h-2 rounded-full overflow-hidden bg-bg-overlay cursor-pointer"
      >
        <div
          class="absolute top-0 left-0 h-full bg-fg-secondary"
          style="width: {muted ? 0 : volume * 100}%"
        ></div>
      </div>
    </div>
  </div>

  <audio
    bind:this={audio}
    src={mediaUrl}
    preload="metadata"
    onplay={onPlay}
    onpause={onPause}
    ontimeupdate={onTimeUpdate}
    onloadedmetadata={onLoadedMetadata}
    onprogress={onProgress}
    onvolumechange={onVolumeChange}
    class="hidden"
  ></audio>
</div>
