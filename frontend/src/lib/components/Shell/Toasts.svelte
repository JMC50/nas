<script lang="ts">
  import X from "lucide-svelte/icons/x";
  import CheckCircle2 from "lucide-svelte/icons/check-circle-2";
  import AlertCircle from "lucide-svelte/icons/alert-circle";
  import AlertTriangle from "lucide-svelte/icons/alert-triangle";
  import Info from "lucide-svelte/icons/info";
  import { notifications } from "$lib/store/notifications.svelte";
  import type { NotificationKind } from "$lib/types";

  const ICONS: Record<NotificationKind, typeof Info> = {
    success: CheckCircle2,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
  };
  const TONE: Record<NotificationKind, string> = {
    success: "text-fg-success border-l-fg-success",
    error: "text-fg-danger border-l-fg-danger",
    warning: "text-fg-warning border-l-fg-warning",
    info: "text-fg-info border-l-fg-info",
  };
</script>

<div
  class="fixed z-[70] right-gutter bottom-9 lg:bottom-10 w-[min(380px,calc(100vw-2*var(--spacing-gutter)))] flex flex-col gap-2 pointer-events-none"
  role="status"
  aria-live="polite"
>
  {#each notifications.queue as item (item.id)}
    {@const Icon = ICONS[item.kind] ?? Info}
    <div
      class="pointer-events-auto grid grid-cols-[18px_minmax(0,1fr)_auto] items-start gap-2.5 py-2 pl-3 pr-1.5 rounded-panel bg-bg-overlay border border-border-default border-l-2 shadow-overlay animate-toast {TONE[item.kind]}"
    >
      <Icon size="18" class="mt-1.5" />
      <p class="text-sm text-fg-primary py-1 break-words">{item.message}</p>
      <button
        type="button"
        class="sg-btn sg-btn-ghost sg-btn-icon"
        onclick={() => notifications.dismiss(item.id)}
        aria-label="Dismiss notification"
      >
        <X size="16" />
      </button>
    </div>
  {/each}
</div>
