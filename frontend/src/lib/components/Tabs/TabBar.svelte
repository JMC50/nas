<script lang="ts">
  import X from "lucide-svelte/icons/x";
  import Plus from "lucide-svelte/icons/plus";
  import Folder from "lucide-svelte/icons/folder";
  import FileText from "lucide-svelte/icons/file-text";
  import ImageIcon from "lucide-svelte/icons/image";
  import Film from "lucide-svelte/icons/film";
  import Music from "lucide-svelte/icons/music";
  import FileType from "lucide-svelte/icons/file-type";
  import Users from "lucide-svelte/icons/users";
  import Settings from "lucide-svelte/icons/settings";
  import History from "lucide-svelte/icons/history";
  import Link2 from "lucide-svelte/icons/link-2";
  import User from "lucide-svelte/icons/user";
  import Cpu from "lucide-svelte/icons/cpu";
  import { tabs } from "$lib/store/tabs.svelte";
  import { ui } from "$lib/store/ui.svelte";
  import type { ExplorerPayload, Tab, TabKind } from "$lib/types";

  const KIND_TO_ICON: Record<TabKind, typeof Folder> = {
    explorer: Folder,
    text: FileText,
    image: ImageIcon,
    video: Film,
    audio: Music,
    pdf: FileType,
    office: FileText,
    "user-manager": Users,
    settings: Settings,
    activity: History,
    "share-links": Link2,
    account: User,
    system: Cpu,
    "music-library": Music,
    "video-library": Film,
  };

  let sourceId = $state<string | null>(null);
  let dragOverId = $state<string | null>(null);

  function focus(tab: Tab) {
    tabs.setActive(tab.id);
  }

  function close(event: MouseEvent, tab: Tab) {
    event.stopPropagation();
    tabs.close(tab.id);
  }

  function onAuxClick(event: MouseEvent, tab: Tab) {
    if (event.button !== 1) return;
    if (!tab.closable) return;
    event.preventDefault();
    tabs.close(tab.id);
  }

  function onKey(event: KeyboardEvent, tab: Tab) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      tabs.setActive(tab.id);
    }
  }

  function onDragStart(event: DragEvent, tab: Tab) {
    if (!event.dataTransfer) return;
    sourceId = tab.id;
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", tab.id);
  }

  function onDragOver(event: DragEvent, tab: Tab) {
    if (!sourceId || sourceId === tab.id) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    dragOverId = tab.id;
  }

  function onDragLeave(tab: Tab) {
    if (dragOverId === tab.id) dragOverId = null;
  }

  function onDrop(event: DragEvent, tab: Tab) {
    event.preventDefault();
    if (!sourceId || sourceId === tab.id) return;
    tabs.reorder(sourceId, tab.id);
    sourceId = null;
    dragOverId = null;
  }

  function onDragEnd() {
    sourceId = null;
    dragOverId = null;
  }

  function openNewTab() {
    const active = tabs.active;
    const loc =
      active?.kind === "explorer"
        ? ((active.payload as ExplorerPayload | null)?.loc ?? [])
        : [];
    tabs.cloneExplorer(loc);
  }
</script>

<div
  class="h-10 pointer-coarse:h-11 pointer-coarse:box-content flex items-stretch gap-1 pl-1 pr-1 overflow-x-auto bg-bg-base border-b border-border-default [scrollbar-width:none]"
  role="tablist"
  aria-label="Open tabs"
>
  {#each tabs.list as tab (tab.id)}
    {@const Icon = KIND_TO_ICON[tab.kind] ?? Folder}
    {@const isActive = tab.id === tabs.activeId}
    {@const isDragOver = dragOverId === tab.id}
    <div
      role="tab"
      tabindex="0"
      draggable={!ui.isMobile}
      aria-selected={isActive}
      class="tab group {isActive ? 'is-active' : ''} {isDragOver ? 'is-drop' : ''}"
      onclick={() => focus(tab)}
      onauxclick={(event) => onAuxClick(event, tab)}
      onkeydown={(event) => onKey(event, tab)}
      ondragstart={(event) => onDragStart(event, tab)}
      ondragover={(event) => onDragOver(event, tab)}
      ondragleave={() => onDragLeave(tab)}
      ondrop={(event) => onDrop(event, tab)}
      ondragend={onDragEnd}
      title={tab.title}
    >
      <Icon size="16" class="shrink-0" />
      <span class="truncate flex-1 min-w-0 text-left">{tab.title}</span>
      {#if tab.dirty}
        <span class="shrink-0 size-1.5 rounded-full bg-fg-warning" role="img" aria-label="Unsaved changes"></span>
      {/if}
      {#if tab.closable}
        <button
          type="button"
          class="tab-close"
          onclick={(event) => close(event, tab)}
          aria-label="Close tab"
        >
          <X size="14" />
        </button>
      {/if}
    </div>
  {/each}
  <button
    type="button"
    class="sg-btn sg-btn-ghost sg-btn-icon self-center shrink-0 focus-visible:outline-offset-[-2px]"
    onclick={openNewTab}
    aria-label="New explorer tab"
    title="New explorer tab (clone current location)"
  >
    <Plus size="16" />
  </button>
</div>

<style>
  .tab {
    position: relative;
    display: flex;
    flex-shrink: 0;
    align-items: center;
    gap: 8px;
    min-width: 112px;
    max-width: 200px;
    padding: 0 6px 0 12px;
    border-radius: var(--radius-control) var(--radius-control) 0 0;
    color: var(--color-fg-secondary);
    font-size: var(--text-sm);
    user-select: none;
    transition: color 70ms var(--ease-snap), background-color 70ms var(--ease-snap);
  }
  .tab:hover {
    color: var(--color-fg-primary);
    background: var(--color-bg-hover);
  }
  .tab:focus-visible {
    outline-offset: -2px;
  }
  .tab.is-active {
    color: var(--color-fg-primary);
    font-weight: 650;
  }
  .tab.is-active::after {
    content: "";
    position: absolute;
    left: 8px;
    right: 8px;
    bottom: 0;
    height: 2px;
    background: var(--color-accent);
    pointer-events: none;
  }
  .tab.is-drop {
    background: var(--color-bg-selected);
    box-shadow: inset 0 0 0 1px var(--color-accent);
  }
  .tab-close {
    display: grid;
    flex-shrink: 0;
    place-items: center;
    width: 24px;
    height: 24px;
    border-radius: var(--radius-chip);
    color: var(--color-fg-muted);
    opacity: 0;
  }
  .tab.is-active .tab-close,
  .tab:hover .tab-close,
  .tab:focus-within .tab-close {
    opacity: 1;
  }
  .tab-close:hover {
    background: var(--color-bg-hover);
    color: var(--color-fg-primary);
  }
  .tab-close:focus-visible {
    outline-offset: -2px;
  }
  @media (pointer: coarse) {
    .tab {
      min-width: 136px;
      padding-right: 0;
    }
    .tab-close {
      opacity: 1;
      width: 44px;
      height: 44px;
    }
  }
</style>
