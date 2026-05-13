<!-- frontend/src/lib/components/Viewers/AudioPlayer.svelte -->
<script lang="ts">
  import Play from "lucide-svelte/icons/play";
  import Pause from "lucide-svelte/icons/pause";
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  let audio: HTMLAudioElement | null = $state(null);
  let playing = $state(false);

  const mediaUrl = $derived(
    `/server/getAudioData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );

  function toggle() {
    if (!audio) return;
    if (audio.paused) audio.play().catch(() => {});
    else audio.pause();
  }

  function onPlay() { playing = true; }
  function onPause() { playing = false; }
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
    <!-- placeholder: scrubber + time + volume go here in Task 11 -->
  </div>

  <audio
    bind:this={audio}
    src={mediaUrl}
    preload="metadata"
    onplay={onPlay}
    onpause={onPause}
    class="hidden"
  ></audio>
</div>
