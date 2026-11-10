<script lang="ts">
  import X from "lucide-svelte/icons/x";
  import { iconFor, type FolderEntry } from "./icon-for";
  import { formatBytes, formatFullTime } from "./format";

  interface Props {
    entry: FolderEntry | null;
    loc: string[];
    onClose: () => void;
  }

  let { entry, loc, onClose }: Props = $props();

  const typeLabel = $derived(
    entry === null
      ? ""
      : entry.isFolder
        ? "Folder"
        : (entry.extensions || "file").toUpperCase(),
  );
  const sizeLabel = $derived(
    entry === null ? "" : entry.isFolder ? "—" : formatBytes(entry.size),
  );
  const fullPath = $derived(entry === null ? "" : [...loc, entry.name].join("/"));

  // --- Mobile swipe-to-dismiss ---
  // Pull the bottom sheet down past DISMISS_THRESHOLD_PX → close. Below the
  // threshold, the sheet snaps back to origin via the transition class. The
  // handlers only attach to the drag handle (md:hidden) so desktop side-panel
  // behavior is untouched.
  const DISMISS_THRESHOLD_PX = 100;
  let dragOffsetY = $state(0);
  let dragStartY: number | null = $state(null);
  let dragPointer: number | null = null;

  function startDrag(event: PointerEvent) {
    if (event.button !== 0) return;
    dragStartY = event.clientY;
    dragOffsetY = 0;
    dragPointer = event.pointerId;
    (event.currentTarget as HTMLElement).setPointerCapture(event.pointerId);
  }

  function moveDrag(event: PointerEvent) {
    if (dragStartY === null) return;
    const delta = event.clientY - dragStartY;
    // Clamp negative deltas to 0 — pulling up does nothing (sheet is already
    // anchored to the screen bottom).
    dragOffsetY = Math.max(0, delta);
  }

  function endDrag(event: PointerEvent) {
    if (dragStartY === null) return;
    const passedThreshold = dragOffsetY > DISMISS_THRESHOLD_PX;
    dragStartY = null;
    if (dragPointer !== null) {
      try {
        (event.currentTarget as HTMLElement).releasePointerCapture(dragPointer);
      } catch {
        // Browser may have already released the capture.
      }
      dragPointer = null;
    }
    if (passedThreshold) {
      // Reset before closing so the next open starts clean.
      dragOffsetY = 0;
      onClose();
    } else {
      // Snap back. Transition class handles the animation.
      dragOffsetY = 0;
    }
  }
</script>

{#if entry}
  {@const Icon = iconFor(entry)}
  <aside
    class="fixed inset-x-0 bottom-0 top-auto max-h-[70dvh] w-full rounded-t-panel border-t border-border-default pb-[env(safe-area-inset-bottom)] animate-sheet md:animate-fade md:top-12 lg:top-0 md:right-0 md:inset-x-auto md:bottom-7 md:pointer-coarse:bottom-11 md:max-h-none md:w-[360px] md:rounded-none md:border-t-0 md:border-l md:pb-0 bg-bg-surface z-40 flex flex-col shadow-overlay {dragStartY === null ? 'transition-transform duration-150' : ''}"
    style="transform: translateY({dragOffsetY}px);"
    aria-label="File details"
  >
    <!--
      Drag handle. Purely visual affordance for touch users — no ARIA role:
      mobile screen readers (VoiceOver, TalkBack) skip aria-hidden elements,
      and the always-present X close button (aria-labelled below) is the
      accessible dismiss path. touch-action:none keeps the browser from
      hijacking the gesture with pull-to-refresh or text-select.
    -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
      class="md:hidden w-full pt-2 pb-1 cursor-grab active:cursor-grabbing touch-none"
      aria-hidden="true"
      onpointerdown={startDrag}
      onpointermove={moveDrag}
      onpointerup={endDrag}
      onpointercancel={endDrag}
    >
      <div class="w-10 h-1 rounded-full bg-border-strong mx-auto"></div>
    </div>
    <header class="min-h-12 flex items-center justify-between gap-2 pl-4 pr-2 border-b border-border-default">
      <h2 class="text-md font-semibold text-fg-primary">File details</h2>
      <button
        type="button"
        class="sg-btn sg-btn-ghost sg-btn-icon"
        onclick={onClose}
        aria-label="Close file details"
      >
        <X size="18" />
      </button>
    </header>

    <div class="flex-1 overflow-y-auto p-4 flex flex-col gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <span class="grid place-items-center size-10 shrink-0 rounded-control bg-bg-elevated">
          <Icon size="20" class={entry.isFolder ? "text-fg-secondary" : "text-fg-muted"} />
        </span>
        <span class="text-sm font-semibold text-fg-primary truncate" title={entry.name}>{entry.name}</span>
      </div>

      <dl class="grid grid-cols-[88px_minmax(0,1fr)] gap-x-3 gap-y-2.5 text-sm">
        <dt class="text-fg-muted">Type</dt>
        <dd class="text-fg-primary truncate">{typeLabel}</dd>
        <dt class="text-fg-muted">Size</dt>
        <dd class="text-fg-primary tabular-nums truncate">{sizeLabel}</dd>
        <dt class="text-fg-muted">Modified</dt>
        <dd class="text-fg-primary tabular-nums truncate">{formatFullTime(entry.modifiedAt)}</dd>
        <dt class="text-fg-muted">Path</dt>
        <dd class="font-mono text-xs leading-5 text-fg-secondary break-all">{fullPath}</dd>
      </dl>
    </div>
  </aside>
{/if}
