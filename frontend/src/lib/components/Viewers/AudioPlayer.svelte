<!-- frontend/src/lib/components/Viewers/AudioPlayer.svelte -->
<script lang="ts">
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import Volume2 from "lucide-svelte/icons/volume-2";
  import Volume1 from "lucide-svelte/icons/volume-1";
  import VolumeX from "lucide-svelte/icons/volume-x";
  import { auth } from "$lib/store/auth.svelte";
  import { formatTime, clampVolume } from "./media-utils";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let audio: HTMLAudioElement | null = $state(null);
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
  function onMeta() {
    if (!audio) return;
    duration = audio.duration;
  }
  function onProgress() {
    if (!audio) return;
    const ranges = audio.buffered;
    buffered = ranges.length > 0 ? ranges.end(ranges.length - 1) : 0;
  }
  function onVolume() {
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
</script>

<div class="flex flex-col h-full w-full bg-bg-base items-center justify-center p-8">
  <div class="w-full max-w-[360px] flex flex-col items-center gap-6">
    <h2 class="text-lg font-sans text-fg-primary truncate w-full text-center" title={name}>
      {name}
    </h2>

    <button
      type="button"
      onclick={toggle}
      aria-label={playing ? "Pause" : "Play"}
      class="w-24 h-24 rounded-full bg-accent text-accent-fg
             flex items-center justify-center
             hover:bg-accent-hover transition-colors
             shadow-lg"
    >
      {#if playing}
        <Pause size="40" fill="currentColor" />
      {:else}
        <Play size="40" fill="currentColor" class="ml-1" />
      {/if}
    </button>

    <!-- placeholder: equalizer goes here in Task 12 -->

    <!-- scrubber -->
    <div class="w-full flex flex-col gap-1">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={seek}
        class="relative h-1.5 rounded-full bg-bg-hover cursor-pointer
               hover:h-2 transition-all"
      >
        <div
          class="absolute top-0 left-0 h-full rounded-full bg-bg-selected"
          style="width: {bufferedRatio * 100}%"
        ></div>
        <div
          class="absolute top-0 left-0 h-full rounded-full bg-accent"
          style="width: {playedRatio * 100}%"
        ></div>
      </div>
      <div class="flex justify-between font-mono text-xs text-fg-secondary tabular-nums">
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
        class="inline-flex items-center justify-center w-8 h-8 rounded
               text-fg-muted hover:bg-bg-hover hover:text-fg-accent transition-colors"
      >
        <VolumeIcon size="16" />
      </button>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={setVolume}
        class="flex-1 relative h-1.5 rounded-full bg-bg-hover cursor-pointer"
      >
        <div
          class="absolute top-0 left-0 h-full rounded-full bg-fg-secondary"
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
    onloadedmetadata={onMeta}
    onprogress={onProgress}
    onvolumechange={onVolume}
    class="hidden"
  ></audio>
</div>
