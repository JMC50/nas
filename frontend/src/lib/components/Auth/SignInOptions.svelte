<script lang="ts">
  import KeyRound from "lucide-svelte/icons/key-round";
  import DiscordIcon from "$lib/components/Auth/DiscordIcon.svelte";
  import GoogleIcon from "$lib/components/Auth/GoogleIcon.svelte";

  interface Props {
    discordEnabled: boolean;
    discordUrl: string;
    googleEnabled: boolean;
    googleUrl: string;
    localEnabled: boolean;
    onLocal: () => void;
  }

  let {
    discordEnabled,
    discordUrl,
    googleEnabled,
    googleUrl,
    localEnabled,
    onLocal,
  }: Props = $props();

  function goDiscord() {
    if (discordUrl) location.href = discordUrl;
  }

  function goGoogle() {
    if (googleUrl) location.href = googleUrl;
  }
</script>

<div class="space-y-2">
  {#if discordEnabled}
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2.5 h-10 rounded-md bg-[#5865F2] text-white text-sm font-medium hover:bg-[#4752C4] transition-colors"
      onclick={goDiscord}
    >
      <DiscordIcon size={16} />
      Continue with Discord
    </button>
  {/if}
  {#if googleEnabled}
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2.5 h-10 rounded-md bg-white text-[#3c4043] text-sm font-medium hover:bg-[#f8f9fa] border border-[#dadce0] transition-colors"
      onclick={goGoogle}
    >
      <GoogleIcon size={16} />
      Continue with Google
    </button>
  {/if}
  {#if localEnabled}
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2 h-10 rounded-md bg-accent text-accent-fg text-sm font-semibold hover:bg-accent-hover transition-colors"
      onclick={onLocal}
    >
      <KeyRound size="14" />
      ID / Password
    </button>
  {/if}
</div>
