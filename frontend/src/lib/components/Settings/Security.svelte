<script lang="ts">
  import Shield from "lucide-svelte/icons/shield";
  import { auth } from "$lib/store/auth.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import TextField from "$lib/components/Auth/TextField.svelte";

  let current = $state("");
  let next = $state("");
  let confirm = $state("");
  let busy = $state(false);

  function clearFields() {
    current = "";
    next = "";
    confirm = "";
  }

  async function callApi(): Promise<{ success: boolean; message?: string }> {
    const response = await fetch(
      `/server/auth/change-password?token=${encodeURIComponent(auth.token)}`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ oldPassword: current, newPassword: next }),
      },
    );
    return response.json();
  }

  async function submit() {
    if (!current || !next) { notifications.warning("Fill in both password fields."); return; }
    if (next !== confirm) { notifications.warning("New passwords do not match."); return; }
    busy = true;
    try {
      const data = await callApi();
      if (data.success) { notifications.success("Password changed."); clearFields(); }
      else notifications.error(data.message ?? "Change failed.");
    } catch (cause) {
      notifications.error(`Change failed: ${(cause as Error).message}`);
    } finally {
      busy = false;
    }
  }
</script>

<section class="flex flex-col gap-3">
  <div class="flex items-center gap-2">
    <Shield size="18" class="text-fg-secondary" />
    <h2 class="sg-section-title">Account & Security</h2>
  </div>

  <div class="sg-panel p-4 flex flex-col gap-3">
    <div>
      <h3 class="text-sm font-semibold text-fg-primary">Change password</h3>
      <p class="text-xs text-fg-muted mt-1">Local accounts only. OAuth users manage credentials with their provider.</p>
    </div>

    <TextField id="current-password" label="Current password" type="password" value={current} disabled={busy} onInput={(v) => (current = v)} />
    <TextField id="new-password" label="New password" type="password" value={next} disabled={busy} onInput={(v) => (next = v)} />
    <TextField id="confirm-password" label="Confirm new password" type="password" value={confirm} disabled={busy} onInput={(v) => (confirm = v)} />

    <button
      type="button"
      class="sg-btn sg-btn-primary self-start"
      onclick={submit}
      disabled={busy}
    >
      {busy ? "Updating…" : "Update password"}
    </button>
  </div>
</section>
