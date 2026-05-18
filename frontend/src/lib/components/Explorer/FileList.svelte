<script lang="ts">
  import MoreVertical from "lucide-svelte/icons/more-vertical";
  import { iconFor, type FolderEntry } from "./icon-for";
  import { dragEntry } from "./drag-drop-action";
  import { longPress } from "$lib/actions/long-press";
  import { ui } from "$lib/store/ui.svelte";
  import { formatBytes, formatRelTime, formatFullTime } from "./format";

  interface Props {
    entries: FolderEntry[];
    dragPayload: (entry: FolderEntry) => string;
    onOpen: (entry: FolderEntry, opts?: { newTab?: boolean }) => void;
    onMenu: (event: MouseEvent, entry: FolderEntry) => void;
    onDropOnFolder: (event: DragEvent, target: FolderEntry) => void;
    // Click → fires for every entry click (modifier-aware). Explorer.svelte
    // dispatches setSingle / toggleEntry / selectRange / openEntry based on the
    // event and viewport. On mobile with no current selection, the handler
    // opens the entry; selection mode is entered via the context menu's
    // "Select" item instead.
    onSelect: (event: MouseEvent, entry: FolderEntry) => void;
    selectedNames: string[];
  }

  let {
    entries,
    dragPayload,
    onOpen,
    onMenu,
    onDropOnFolder,
    onSelect,
    selectedNames,
  }: Props = $props();

  const selectedSet = $derived(new Set(selectedNames));
  let dropTargetName = $state<string | null>(null);

  // Long-press unconditionally opens the context menu at the touch point.
  // Mobile selection entry is exposed via the menu's "Select" item rather
  // than a competing gesture, so long-press is always menu and never a
  // hidden mode-switch (avoids the 500ms-delay confusion users hit when
  // the same gesture meant different things in different states).
  function handleLongPress(entry: FolderEntry, clientX: number, clientY: number) {
    const synthetic = new MouseEvent("contextmenu", {
      clientX,
      clientY,
      bubbles: true,
    });
    onMenu(synthetic, entry);
  }

  // Kebab tap on mobile — opens the same menu as long-press, anchored at
  // the kebab's screen position. stopPropagation prevents the row's onclick
  // (onSelect) from also firing.
  function onKebab(event: MouseEvent, entry: FolderEntry) {
    event.stopPropagation();
    onMenu(event, entry);
  }

  // Without this, pressing the kebab would also start the row's longPress
  // timer (action is on <tr>, kebab pointerdown bubbles up). After 500ms the
  // timer would fire onMenu, then pointerup → click → onKebab → onMenu again,
  // opening the menu twice. Stopping pointerdown at the kebab keeps the row's
  // longPress action from ever seeing the press.
  function stopRowLongPress(event: PointerEvent) {
    event.stopPropagation();
  }
</script>

<!--
  data-marquee-canvas="false" stops marquee pointerdown trigger from firing on
  table rows — only blank-area mousedowns inside the canvas start a marquee.
-->
<table class="w-full text-sm" data-marquee-canvas="false">
  <thead>
    <tr class="text-xs text-fg-muted border-b border-border-default">
      <th class="text-left font-normal px-6 py-2">Name</th>
      <th class="text-left font-normal px-6 py-2 w-24">Type</th>
      <th class="text-right font-normal px-6 py-2 w-20 font-mono">Size</th>
      <th class="text-left font-normal px-6 py-2 w-32">Modified</th>
      <!-- Kebab column header — empty label, mobile-only width reservation. -->
      <th class="md:hidden w-10" aria-hidden="true"></th>
    </tr>
  </thead>
  <tbody>
    {#each entries as entry (entry.name)}
      {@const Icon = iconFor(entry)}
      {@const isSelected = selectedSet.has(entry.name)}
      {@const isDropTarget = entry.isFolder && dropTargetName === entry.name}
      <tr
        data-entry-name={entry.name}
        class="border-b border-border-default/40 hover:bg-bg-hover cursor-pointer {isDropTarget
          ? 'ring-1 ring-inset ring-accent bg-bg-elevated'
          : ''} {isSelected ? 'ring-1 ring-inset ring-accent bg-bg-elevated' : ''}"
        draggable={!ui.isMobile}
        use:dragEntry={{
          entry,
          dragPayload,
          onDropOnFolder,
          onOpen,
          onDropEnter: (name) => (dropTargetName = name),
          onDropLeave: (name) => {
            if (dropTargetName === name) dropTargetName = null;
          },
          onDropFinish: () => (dropTargetName = null),
        }}
        use:longPress={{
          onLongPress: (clientX, clientY) => handleLongPress(entry, clientX, clientY),
        }}
        oncontextmenu={(event) => onMenu(event, entry)}
        onclick={(event) => onSelect(event, entry)}
      >
        <td class="px-6 py-1.5">
          <div class="flex items-center gap-2 min-w-0">
            <Icon size="14" class={entry.isFolder ? "text-accent" : "text-fg-secondary"} />
            <span class="truncate text-fg-primary">{entry.name}</span>
          </div>
        </td>
        <td class="px-6 py-1.5 text-xs text-fg-muted">
          {entry.isFolder ? "Folder" : entry.extensions || "file"}
        </td>
        <td class="px-6 py-1.5 text-xs text-fg-muted font-mono text-right">
          {entry.isFolder ? "—" : formatBytes(entry.size)}
        </td>
        <td
          class="px-6 py-1.5 text-xs text-fg-muted"
          title={formatFullTime(entry.modifiedAt)}
        >
          {formatRelTime(entry.modifiedAt)}
        </td>
        <!--
          Mobile-only kebab. Hidden at md+ where right-click / two-finger tap
          already invokes the context menu. data-marquee-canvas inherited as
          "false" from the parent <table>, so the kebab cannot start a marquee.
        -->
        <td class="md:hidden px-2 py-1.5 text-right">
          <button
            type="button"
            class="inline-flex items-center justify-center w-8 h-8 rounded text-fg-muted hover:text-fg-primary hover:bg-bg-hover"
            onpointerdown={stopRowLongPress}
            onclick={(event) => onKebab(event, entry)}
            aria-label="More actions for {entry.name}"
          >
            <MoreVertical size="16" />
          </button>
        </td>
      </tr>
    {/each}
  </tbody>
</table>
