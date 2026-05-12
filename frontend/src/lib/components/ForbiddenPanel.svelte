<script lang="ts">
  import Lock from "lucide-svelte/icons/lock";
  import ShieldCheck from "lucide-svelte/icons/shield-check";
  import { claimAdmin } from "$lib/auth-actions";
  import { notifications } from "$lib/store/notifications.svelte";

  interface Props {
    title?: string;
    description?: string;
    onGranted?: () => void;
  }

  let {
    title = "Permission required",
    description = "You don't have access to this view. Ask an administrator to grant you the right intent, or claim admin yourself with the admin password.",
    onGranted,
  }: Props = $props();

  let busy = $state(false);

  async function tryClaim() {
    const password = prompt("Enter the admin password:");
    if (!password) return;
    busy = true;
    const result = await claimAdmin(password);
    busy = false;
    if (result.success) {
      notifications.success("Admin permission granted.");
      onGranted?.();
    } else {
      notifications.error(result.message ?? "Failed to claim admin.");
    }
  }
</script>

<div class="flex items-center justify-center h-full w-full p-6">
  <div class="max-w-md w-full text-center">
    <div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-accent/10 border border-accent/20 mb-5">
      <Lock size="28" class="text-accent" />
    </div>

    <h2 class="text-lg font-semibold text-fg-primary mb-2">{title}</h2>
    <p class="text-sm text-fg-muted mb-6 leading-relaxed">{description}</p>

    <button
      type="button"
      class="inline-flex items-center gap-2 h-10 px-4 rounded-md bg-accent text-accent-fg text-sm font-semibold hover:bg-accent-hover disabled:opacity-60 transition-colors"
      onclick={tryClaim}
      disabled={busy}
    >
      <ShieldCheck size="14" />
      {busy ? "Verifying…" : "Request admin permission"}
    </button>

    <p class="mt-4 text-xs text-fg-muted">
      Or contact your administrator to grant the required permission.
    </p>
  </div>
</div>
