<script lang="ts">
  import FileText from "lucide-svelte/icons/file-text";
  import FolderOpen from "lucide-svelte/icons/folder-open";
  import ExternalLink from "lucide-svelte/icons/external-link";
  import DownloadIcon from "lucide-svelte/icons/download";
  import Pencil from "lucide-svelte/icons/pencil";
  import Copy from "lucide-svelte/icons/copy";
  import Trash2 from "lucide-svelte/icons/trash-2";
  import Info from "lucide-svelte/icons/info";

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
    onDownload: () => void;
    onRename: () => void;
    onCopy: () => void;
    onDelete: () => void;
    onInspect: () => void;
  }

  let {
    target,
    x,
    y,
    onOpen,
    onOpenNewTab,
    onDownload,
    onRename,
    onCopy,
    onDelete,
    onInspect,
  }: Props = $props();
</script>

<div
  class="fixed z-50 min-w-[160px] py-1 rounded-md bg-bg-overlay border border-border-strong shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
  style="left: {x}px; top: {y}px;"
  role="menu"
>
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
    onclick={onOpen}
    role="menuitem"
  >
    {#if target.isFolder}
      <FolderOpen size="12" />
    {:else}
      <FileText size="12" />
    {/if}
    Open
  </button>
  {#if target.isFolder}
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
      onclick={onOpenNewTab}
      role="menuitem"
    >
      <ExternalLink size="12" />
      Open in new tab
    </button>
  {/if}
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
    onclick={onDownload}
    role="menuitem"
  >
    <DownloadIcon size="12" />
    Download
  </button>
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
    onclick={onRename}
    role="menuitem"
  >
    <Pencil size="12" />
    Rename
  </button>
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
    onclick={onCopy}
    role="menuitem"
  >
    <Copy size="12" />
    Copy
  </button>
  <div class="h-px bg-border-default mx-1 my-1"></div>
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
    onclick={onInspect}
    role="menuitem"
  >
    <Info size="12" />
    Details
  </button>
  <button
    type="button"
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-danger hover:bg-bg-hover text-left"
    onclick={onDelete}
    role="menuitem"
  >
    <Trash2 size="12" />
    Delete
  </button>
</div>
