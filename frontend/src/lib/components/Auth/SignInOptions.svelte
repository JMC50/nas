<script lang="ts">
  import KeyRound from "lucide-svelte/icons/key-round";
  import DiscordIcon from "$lib/components/Auth/DiscordIcon.svelte";
  import GoogleIcon from "$lib/components/Auth/GoogleIcon.svelte";

  interface Props {
    oauthEnabled: boolean;
    localEnabled: boolean;
    onDiscord: () => void;
    onGoogle: () => void;
    onLocal: () => void;
  }

  let { oauthEnabled, localEnabled, onDiscord, onGoogle, onLocal }: Props = $props();
</script>

<div class="space-y-2">
  {#if oauthEnabled}
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2.5 h-10 rounded-md bg-[#5865F2] text-white text-sm font-medium hover:bg-[#4752C4] transition-colors"
      onclick={onDiscord}
    >
      <DiscordIcon size={16} />
      Continue with Discord
    </button>
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2.5 h-10 rounded-md bg-white text-[#3c4043] text-sm font-medium hover:bg-[#f8f9fa] border border-[#dadce0] transition-colors"
      onclick={onGoogle}
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
