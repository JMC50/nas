<script lang="ts">
  import { auth } from "$lib/store/auth.svelte";

  interface Props {
    loc: string;
    name: string;
  }

  let { loc, name }: Props = $props();

  const pdfUrl = $derived(
    `/server/download?token=${encodeURIComponent(auth.token)}&loc=${encodeURIComponent(loc)}&name=${encodeURIComponent(name)}`,
  );
</script>

<div class="flex flex-col h-full w-full bg-bg-base">
  <div class="flex items-center px-3 h-9 border-b border-border-default text-xs text-fg-secondary">
    <span class="truncate">{name}</span>
  </div>
  <div class="flex-1">
    <iframe src={pdfUrl} class="w-full h-full bg-bg-elevated" title={name}></iframe>
  </div>
</div>
