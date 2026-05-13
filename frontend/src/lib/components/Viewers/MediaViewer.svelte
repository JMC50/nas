<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";
  import VideoPlayer from "./VideoPlayer.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();

  // audio path unchanged for now
  const audioUrl = $derived(
    `/server/getAudioData?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

{#if kind === "video"}
  <VideoPlayer {loc} {name} />
{:else}
  <div class="flex flex-col h-full w-full bg-bg-base">
    <div class="flex items-center px-3 h-9 border-b border-border-default text-xs text-fg-secondary">
      <span class="truncate">{name}</span>
    </div>
    <div class="flex-1 flex items-center justify-center p-4">
      <audio src={audioUrl} controls class="w-full max-w-xl"></audio>
    </div>
  </div>
{/if}
