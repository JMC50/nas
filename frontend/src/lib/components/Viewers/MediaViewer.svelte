<script lang="ts">
  import VideoPlayer from "./VideoPlayer.svelte";
  import AudioPlayer from "./AudioPlayer.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();
</script>

{#if kind === "video"}
  <VideoPlayer {loc} {name} />
{:else}
  <AudioPlayer {loc} {name} />
{/if}
