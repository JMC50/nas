<script lang="ts">
  import { onMount } from "svelte";
  import Lock from "lucide-svelte/icons/lock";
  import ShieldCheck from "lucide-svelte/icons/shield-check";
  import { claimAdmin } from "$lib/auth-actions";
  import { oidcAddresses } from "$lib/oidc";
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
  // 통합 로그인 서버는 관리자 비밀번호를 받지 않음. 관리자 권한은 발급자의 그룹으로만 옴
  let ssoOnly = $state(false);

  onMount(async () => {
    ssoOnly = (await oidcAddresses()) !== null;
  });

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

<!--
  data-marquee-canvas="false" opts out of Explorer's marquee pointer capture.
  Without this, Explorer's containerEl swallows pointerdown via setPointerCapture
  and the button never sees a complete click sequence — so prompt() never fires.
-->
<div class="flex items-center justify-center h-full w-full p-gutter" data-marquee-canvas="false">
  <div class="max-w-md w-full text-center">
    <div class="inline-grid place-items-center size-14 rounded-panel bg-bg-elevated border border-border-default mb-5">
      <Lock size="26" class="text-fg-secondary" />
    </div>

    <h2 class="text-lg font-semibold text-fg-primary mb-2">{title}</h2>
    {#if ssoOnly}
      <p class="text-sm text-fg-secondary leading-relaxed">
        You don't have access to this view. Ask an administrator to grant you the right intent.
      </p>
    {:else}
      <p class="text-sm text-fg-secondary mb-6 leading-relaxed">{description}</p>

      <button
        type="button"
        class="sg-btn sg-btn-primary"
        onclick={tryClaim}
        disabled={busy}
      >
        <ShieldCheck size="16" />
        {busy ? "Verifying…" : "Request admin permission"}
      </button>

      <p class="mt-4 text-xs text-fg-muted">
        Or contact your administrator to grant the required permission.
      </p>
    {/if}
  </div>
</div>
