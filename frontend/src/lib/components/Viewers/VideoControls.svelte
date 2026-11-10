<!-- frontend/src/lib/components/Viewers/VideoControls.svelte -->
<script lang="ts">
  import type { SvelteComponentTyped } from "svelte";
  // lucide-svelte icons extend SvelteComponentTyped (legacy class API).
  // `import type` keeps this purely at compile time — at runtime Svelte 5
  // does not export SvelteComponentTyped as a value, only as a type.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  type IconComponent = typeof SvelteComponentTyped<any>;
  import { formatTime } from "./media-utils";
  import { ui } from "$lib/store/ui.svelte";
  import { clickOutside } from "$lib/actions/click-outside";
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import Maximize from "lucide-svelte/icons/maximize";
  import Minimize from "lucide-svelte/icons/minimize";
  import PictureInPicture2 from "lucide-svelte/icons/picture-in-picture-2";
  import MoreHorizontal from "lucide-svelte/icons/more-horizontal";

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

  // Mobile-only: secondary controls (speed, PIP) collapse into a "more" popup
  // anchored to the overflow button so the controls row fits on a 375px width.
  let moreOpen = $state(false);

  function pickSpeed(rate: number) {
    onSetSpeed(rate);
    moreOpen = false;
  }

  function pickPip() {
    onTogglePip();
    moreOpen = false;
  }
</script>

<!-- bottom controls bar -->
<div
  class="absolute bottom-0 left-0 right-0 px-3 pb-1.5 pt-3 flex flex-col gap-1
 bg-bg-overlay/90 border-t border-border-default
 text-fg-primary z-10
 transition-opacity duration-120 ease-snap
 {controlsVisible ? 'opacity-100' : 'opacity-0 pointer-events-none'}"
>
  <!-- scrubber -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    onclick={onSeek}
    class="relative h-2 rounded-full bg-bg-hover cursor-pointer group/scrub"
  >
    <!-- buffered -->
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-border-strong/50"
      style="width: {bufferedRatio * 100}%"
    ></div>
    <!-- played -->
    <div
      class="absolute top-0 left-0 h-full rounded-full bg-accent"
      style="width: {playedRatio * 100}%"
    ></div>
    <!-- thumb -->
    <div
      class="absolute top-1/2 -translate-x-1/2 -translate-y-1/2 size-3 rounded-full bg-accent
 opacity-0 group-hover/scrub:opacity-100 transition-opacity"
      style="left: {playedRatio * 100}%"
    ></div>
  </div>

  <!-- controls row -->
  <div class="flex items-center gap-1 min-h-9">
    <button
      type="button"
      onclick={onToggle}
      aria-label={playing ? "Pause" : "Play"}
      class="sg-btn sg-btn-ghost sg-btn-icon"
    >
      {#if playing}
        <Pause size="18" />
      {:else}
        <Play size="18" />
      {/if}
    </button>

    <span class="px-1 text-xs text-fg-secondary tabular-nums whitespace-nowrap">
      {formatTime(currentTime)} / {formatTime(duration)}
    </span>

    <!-- volume control (button + slide-out slider) -->
    <div class="flex items-center">
      <button
        type="button"
        onclick={onToggleMute}
        aria-label={muted ? "Unmute" : "Mute"}
        class="sg-btn sg-btn-ghost sg-btn-icon"
      >
        <VolumeIcon size="18" />
      </button>
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        onclick={onSetVolume}
        class="hidden md:block relative h-2 w-20 ml-1 rounded-full bg-bg-hover cursor-pointer overflow-hidden"
      >
        <div
          class="absolute top-0 left-0 h-full bg-fg-secondary"
          style="width: {muted ? 0 : volume * 100}%"
        ></div>
      </div>
    </div>

    <!-- spacer; later tasks fill the right side -->
    <div class="ml-auto"></div>

    <!-- Desktop: inline speed menu + PIP -->
    <div class="hidden md:block relative" data-speed-menu>
      <button
        type="button"
        onclick={onToggleSpeedMenu}
        aria-label="Playback speed"
        aria-expanded={speedMenuOpen}
        class="sg-btn sg-btn-ghost tabular-nums"
      >
        {speed}×
      </button>
      {#if speedMenuOpen}
        <div
          class="absolute bottom-full right-0 mb-2 z-30 sg-menu min-w-[96px] animate-pop origin-bottom-right"
          role="menu"
        >
          {#each SPEEDS as rate (rate)}
            <button
              type="button"
              onclick={() => onSetSpeed(rate)}
              class="sg-menu-item tabular-nums {speed === rate ? 'font-semibold text-fg-accent' : ''}"
              role="menuitemradio"
              aria-checked={speed === rate}
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
        class="hidden md:inline-flex sg-btn sg-btn-ghost sg-btn-icon"
      >
        <PictureInPicture2 size="18" />
      </button>
    {/if}

    <!-- Mobile: overflow menu for speed + PIP -->
    {#if ui.isMobile}
      <div class="relative" use:clickOutside={{ onOutside: () => (moreOpen = false), enabled: moreOpen }}>
        <button
          type="button"
          onclick={(event) => {
            event.stopPropagation();
            moreOpen = !moreOpen;
          }}
          aria-label="More controls"
          aria-expanded={moreOpen}
          class="sg-btn sg-btn-ghost sg-btn-icon"
        >
          <MoreHorizontal size="18" />
        </button>
        {#if moreOpen}
          <div
            class="absolute bottom-full right-0 mb-2 z-30 sg-menu min-w-[180px] animate-pop origin-bottom-right"
            role="menu"
          >
            <div class="px-2.5 pt-1 pb-1.5 text-xs font-semibold text-fg-muted">
              Speed
            </div>
            {#each SPEEDS as rate (rate)}
              <button
                type="button"
                onclick={() => pickSpeed(rate)}
                class="sg-menu-item tabular-nums {speed === rate ? 'font-semibold text-fg-accent' : ''}"
                role="menuitemradio"
                aria-checked={speed === rate}
              >
                {rate}×
              </button>
            {/each}
            {#if pipSupported}
              <div class="sg-menu-sep" role="separator"></div>
              <button
                type="button"
                onclick={pickPip}
                class="sg-menu-item"
                role="menuitem"
              >
                <PictureInPicture2 size="16" />
                <span>Picture in Picture</span>
              </button>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    <button
      type="button"
      onclick={onToggleFullscreen}
      aria-label={isFullscreen ? "Exit fullscreen" : "Fullscreen"}
      class="sg-btn sg-btn-ghost sg-btn-icon"
    >
      {#if isFullscreen}
        <Minimize size="18" />
      {:else}
        <Maximize size="18" />
      {/if}
    </button>
  </div>
</div>
