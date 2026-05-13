<!-- frontend/src/lib/components/Viewers/VideoControls.svelte -->
<script lang="ts">
  import type { SvelteComponentTyped } from "svelte";
  // lucide-svelte icons extend SvelteComponentTyped (legacy class API).
  // `import type` keeps this purely at compile time — at runtime Svelte 5
  // does not export SvelteComponentTyped as a value, only as a type.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  type IconComponent = typeof SvelteComponentTyped<any>;
  import { formatTime } from "./media-utils";
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import Maximize from "lucide-svelte/icons/maximize";
  import Minimize from "lucide-svelte/icons/minimize";
  import PictureInPicture2 from "lucide-svelte/icons/picture-in-picture-2";

  interface Props {
    controlsVisible: boolean;
    playing: boolean;
    currentTime: number;
    duration: number;
    playedRatio: number;
    bufferedRatio: number;
    volume: number;
    muted: boolean;
    VolumeIcon: IconComponent;
    speed: number;
    speedMenuOpen: boolean;
    SPEEDS: readonly number[];
    isFullscreen: boolean;
    pipSupported: boolean;
    onToggle: () => void;
    onSeek: (e: MouseEvent) => void;
    onToggleMute: () => void;
    onSetVolume: (e: MouseEvent) => void;
    onSetSpeed: (rate: number) => void;
    onToggleSpeedMenu: (e: MouseEvent) => void;
    onTogglePip: () => void;
    onToggleFullscreen: () => void;
  }

  let {
    controlsVisible,
    playing,
    currentTime,
    duration,
    playedRatio,
    bufferedRatio,
    volume,
    muted,
    VolumeIcon,
    speed,
    speedMenuOpen,
    SPEEDS,
    isFullscreen,
    pipSupported,
    onToggle,
    onSeek,
    onToggleMute,
    onSetVolume,
    onSetSpeed,
    onToggleSpeedMenu,
    onTogglePip,
    onToggleFullscreen,
  }: Props = $props();
</script>

<!-- bottom controls bar -->
<div
  class="absolute bottom-0 left-0 right-0 px-4 pb-2 pt-3 flex flex-col gap-1
         bg-bg-overlay/80 backdrop-blur-sm
         text-fg-primary z-10
         transition-opacity duration-200 ease-smooth
         {controlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'}"
>
  <!-- scrubber -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={onSeek}
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
      onclick={onToggle}
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
        onclick={onToggleMute}
        aria-label={muted ? "Unmute" : "Mute"}
        class="inline-flex items-center justify-center w-9 h-9 rounded
               hover:bg-bg-hover hover:text-fg-accent transition-colors"
      >
        <VolumeIcon size="18" />
      </button>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={onSetVolume}
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

    <div class="relative" data-speed-menu>
      <button
        type="button"
        onclick={onToggleSpeedMenu}
        aria-label="Playback speed"
        class="inline-flex items-center justify-center min-w-9 h-9 px-2 rounded
               font-mono text-xs tabular-nums
               hover:bg-bg-hover hover:text-fg-accent transition-colors"
      >
        {speed}×
      </button>
      {#if speedMenuOpen}
        <div
          class="absolute bottom-full right-0 mb-2 py-1 rounded-md
                 bg-bg-elevated border border-border-default
                 shadow-lg z-30 min-w-[72px]"
        >
          {#each SPEEDS as rate (rate)}
            <button
              type="button"
              onclick={() => onSetSpeed(rate)}
              class="w-full px-3 py-1.5 text-xs font-mono tabular-nums text-left
                     hover:bg-bg-hover transition-colors
                     {speed === rate ? 'text-fg-accent' : 'text-fg-primary'}"
            >
              {rate}×
            </button>
          {/each}
        </div>
      {/if}
    </div>

    {#if pipSupported}
      <button
        type="button"
        onclick={onTogglePip}
        aria-label="Picture in Picture"
        class="inline-flex items-center justify-center w-9 h-9 rounded
               hover:bg-bg-hover hover:text-fg-accent transition-colors"
      >
        <PictureInPicture2 size="18" />
      </button>
    {/if}
    <button
      type="button"
      onclick={onToggleFullscreen}
      aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
      class="inline-flex items-center justify-center w-9 h-9 rounded
             hover:bg-bg-hover hover:text-fg-accent transition-colors"
    >
      {#if isFullscreen}
        <Minimize size="18" />
      {:else}
        <Maximize size="18" />
      {/if}
    </button>
  </div>
</div>
