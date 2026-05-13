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
  import User from "lucide-svelte/icons/user";
  import Cpu from "lucide-svelte/icons/cpu";
  import { tabs } from "$lib/store/tabs.svelte";
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
    account: User,
    system: Cpu,
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
  class="h-9 flex items-end overflow-x-auto bg-bg-base border-b border-border-default"
  role="tablist"
>
  <div class="flex items-end h-full">
    {#each tabs.list as tab (tab.id)}
      {@const Icon = KIND_TO_ICON[tab.kind] ?? Folder}
      {@const isActive = tab.id === tabs.activeId}
      {@const isDragOver = dragOverId === tab.id}
      <div
        role="tab"
        tabindex="0"
        draggable="true"
        aria-selected={isActive}
        class="group relative flex items-center gap-2 h-full px-3 text-xs border-r border-border-default transition-colors min-w-[120px] max-w-[200px] cursor-pointer select-none {isActive
          ? 'bg-bg-base text-fg-primary'
          : 'bg-bg-surface text-fg-muted hover:text-fg-primary hover:bg-bg-elevated'} {isDragOver
          ? 'ring-1 ring-inset ring-accent'
          : ''}"
        onclick={() => focus(tab)}
        onauxclick={(event) => onAuxClick(event, tab)}
        onkeydown={(event) => onKey(event, tab)}
        ondragstart={(event) => onDragStart(event, tab)}
        ondragover={(event) => onDragOver(event, tab)}
        ondragleave={() => onDragLeave(tab)}
        ondrop={(event) => onDrop(event, tab)}
        ondragend={onDragEnd}
      >
        {#if isActive}
          <span class="absolute top-0 left-0 right-0 h-0.5 bg-accent"></span>
        {/if}
        <Icon size="14" class="shrink-0" />
        <span class="truncate flex-1 text-left">
          {tab.title}{#if tab.dirty}<span class="text-fg-warning"> •</span>{/if}
        </span>
        {#if tab.closable}
          <button
            type="button"
            class="shrink-0 w-4 h-4 inline-flex items-center justify-center rounded text-fg-muted hover:text-fg-danger hover:bg-bg-hover opacity-0 group-hover:opacity-100 transition-opacity"
            onclick={(event) => close(event, tab)}
            aria-label="Close tab"
          >
            <X size="12" />
          </button>
        {/if}
      </div>
    {/each}
  </div>
  <button
    type="button"
    class="shrink-0 inline-flex items-center justify-center w-8 h-8 mb-0.5 ml-1 rounded text-fg-muted hover:text-fg-primary hover:bg-bg-elevated"
    onclick={openNewTab}
    aria-label="New explorer tab"
    title="New explorer tab (clone current location)"
  >
    <Plus size="14" />
  </button>
</div>
