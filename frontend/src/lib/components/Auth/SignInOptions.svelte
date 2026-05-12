<script lang="ts">
  import LogIn from "lucide-svelte/icons/log-in";
  import KeyRound from "lucide-svelte/icons/key-round";

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
      class="w-full flex items-center justify-center gap-2 h-10 rounded-md bg-bg-elevated text-fg-primary text-sm font-medium hover:bg-bg-hover border border-border-default transition-colors"
      onclick={onDiscord}
    >
      <LogIn size="14" />
      Continue with Discord
    </button>
    <button
      type="button"
      class="w-full flex items-center justify-center gap-2 h-10 rounded-md bg-bg-elevated text-fg-primary text-sm font-medium hover:bg-bg-hover border border-border-default transition-colors"
      onclick={onGoogle}
    >
      <LogIn size="14" />
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
