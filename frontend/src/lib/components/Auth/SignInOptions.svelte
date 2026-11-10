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

<div class="flex flex-col gap-2">
  {#if discordEnabled}
    <button type="button" class="sg-btn sg-btn-block" onclick={goDiscord}>
      <DiscordIcon size={20} class="text-brand-discord" />
      Continue with Discord
    </button>
  {/if}
  {#if googleEnabled}
    <button type="button" class="sg-btn sg-btn-block" onclick={goGoogle}>
      <GoogleIcon size={20} />
      Continue with Google
    </button>
  {/if}
  {#if localEnabled}
    <button type="button" class="sg-btn sg-btn-primary sg-btn-block" onclick={onLocal}>
      <KeyRound size="16" />
      ID / Password
    </button>
  {/if}
</div>
