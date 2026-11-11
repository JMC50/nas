<script lang="ts">
  import FileText from "lucide-svelte/icons/file-text";
  import FolderOpen from "lucide-svelte/icons/folder-open";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import DownloadIcon from "lucide-svelte/icons/download";
  import Link2 from "lucide-svelte/icons/link-2";
  import Pencil from "lucide-svelte/icons/pencil";
  import Copy from "lucide-svelte/icons/copy";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Info from "lucide-svelte/icons/info";
  import CheckSquare from "lucide-svelte/icons/check-square";
  import { ui } from "$lib/store/ui.svelte";

  interface Target {
    name: string;
    isFolder: boolean;
    extensions: string;
  }

  interface Props {
    target: Target;
    x: number;
    y: number;
    onOpen: () => void;
    onOpenNewTab: () => void;
    downloadLabel: string;
    onDownload: () => void;
    onShare: () => void;
    onRename: () => void;
    onCopy: () => void;
    onDelete: () => void;
    onInspect: () => void;
    // Touch-only "Select" menu item — enters multi-select mode anchored at
    // this entry. Desktop has shift/ctrl-click + drag-marquee for the same
    // purpose, so the item is hidden on fine-pointer layouts at md+.
    onSelect: () => void;
    onClose: () => void;
  }

  let {
    target,
    x,
    y,
    onOpen,
    onOpenNewTab,
    downloadLabel,
    onDownload,
    onShare,
    onRename,
    onCopy,
    onDelete,
    onInspect,
    onSelect,
    onClose,
  }: Props = $props();

  let menuEl: HTMLDivElement | undefined = $state();
  let size = $state({ width: 0, height: 0 });
  const left = $derived(Math.max(8, Math.min(x, ui.viewportWidth - size.width - 8)));
  const top = $derived(Math.max(8, Math.min(y, ui.viewportHeight - size.height - 8)));

  function onMenuKey(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      event.stopPropagation();
      onClose();
      return;
    }
    if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
    event.preventDefault();
    const items = Array.from(menuEl?.querySelectorAll<HTMLElement>('[role="menuitem"]') ?? []);
    const index = items.indexOf(document.activeElement as HTMLElement);
    const step = event.key === "ArrowDown" ? 1 : -1;
    items[(index + step + items.length) % items.length]?.focus();
  }

  $effect(() => {
    if (!menuEl) return;
    size = { width: menuEl.offsetWidth, height: menuEl.offsetHeight };
    menuEl.querySelector<HTMLElement>('[role="menuitem"]')?.focus({ preventScroll: true });
  });
</script>

{#if ui.isMobile}
  <div class="fixed inset-0 z-40 bg-black/55 animate-fade" aria-hidden="true"></div>
{/if}
<div
  bind:this={menuEl}
  class="fixed z-50 sg-menu {ui.isMobile ? 'inset-x-0 bottom-0 sg-menu-sheet animate-sheet' : 'animate-pop origin-top-left'}"
  style={ui.isMobile ? "" : `left: ${left}px; top: ${top}px;`}
  role="menu"
  tabindex="-1"
  aria-label="Actions for {target.name}"
  onkeydown={onMenuKey}
>
  {#if ui.isMobile}
    <div class="px-2.5 pt-1 pb-2 text-xs text-fg-muted truncate">{target.name}</div>
  {/if}
  <button
    type="button"
    class="sg-menu-item"
    onclick={onOpen}
    role="menuitem"
  >
    {#if target.isFolder}
      <FolderOpen size="16" />
    {:else}
      <FileText size="16" />
    {/if}
    Open
  </button>
  {#if ui.touchLayout}
    <button
      type="button"
      class="sg-menu-item"
      onclick={onSelect}
      role="menuitem"
    >
      <CheckSquare size="16" />
      Select
    </button>
  {/if}
  {#if target.isFolder}
    <button
      type="button"
      class="sg-menu-item"
      onclick={onOpenNewTab}
      role="menuitem"
    >
      <ExternalLink size="16" />
      Open in new tab
    </button>
  {/if}
  <button
    type="button"
    class="sg-menu-item"
    onclick={onDownload}
    role="menuitem"
  >
    <DownloadIcon size="16" />
    {downloadLabel}
  </button>
  <button
    type="button"
    class="sg-menu-item"
    onclick={onShare}
    role="menuitem"
  >
    <Link2 size="16" />
    Create download link
  </button>
  <button
    type="button"
    class="sg-menu-item"
    onclick={onRename}
    role="menuitem"
  >
    <Pencil size="16" />
    Rename
  </button>
  <button
    type="button"
    class="sg-menu-item"
    onclick={onCopy}
    role="menuitem"
  >
    <Copy size="16" />
    Copy
  </button>
  <div class="sg-menu-sep" role="separator"></div>
  <button
    type="button"
    class="sg-menu-item"
    onclick={onInspect}
    role="menuitem"
  >
    <Info size="16" />
    Details
  </button>
  <button
    type="button"
    class="sg-menu-item sg-menu-item-danger"
    onclick={onDelete}
    role="menuitem"
  >
    <Trash2 size="16" />
    Delete
  </button>
</div>
