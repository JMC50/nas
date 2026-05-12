<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
    kind: "video" | "audio";
  }

  let { loc, name, kind }: Props = $props();

  const endpoint = $derived(kind === "video" ? "getVideoData" : "getAudioData");
  const mediaUrl = $derived(
    `/server/${endpoint}?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="flex items-center px-3 h-9 border-b border-border-default text-xs text-fg-secondary">
    <span class="truncate">{name}</span>
  </div>
  <div class="flex-1 flex items-center justify-center p-4">
    {#if kind === "video"}
      <!-- svelte-ignore a11y_media_has_caption -->
      <video
        src={mediaUrl}
        controls
        class="max-w-full max-h-full rounded-md border border-border-default bg-black"
      ></video>
    {:else}
      <audio src={mediaUrl} controls class="w-full max-w-xl"></audio>
    {/if}
  </div>
</div>
