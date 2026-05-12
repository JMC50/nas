<script lang="ts">
  import FileText from "lucide-svelte/icons/file-text";
  import DownloadIcon from "lucide-svelte/icons/download";
  import Pencil from "lucide-svelte/icons/pencil";
  import Copy from "lucide-svelte/icons/copy";
  import Trash2 from "lucide-svelte/icons/trash-2";

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
    onDownload: () => void;
    onRename: () => void;
    onCopy: () => void;
    onDelete: () => void;
  }

  let { target, x, y, onOpen, onDownload, onRename, onCopy, onDelete }: Props = $props();
</script>

<div
  class="fixed z-50 min-w-[160px] py-1 rounded-md bg-bg-overlay border border-border-strong shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
  style="left: {x}px; top: {y}px;"
  role="menu"
>
  {#if !target.isFolder}
    <button
      type="button"
      class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-primary hover:bg-bg-hover text-left"
      onclick={onOpen}
      role="menuitem"
    >
      <FileText size="12" />
      Open
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
    class="w-full flex items-center gap-2 px-3 h-8 text-xs text-fg-danger hover:bg-bg-hover text-left"
    onclick={onDelete}
    role="menuitem"
  >
    <Trash2 size="12" />
    Delete
  </button>
</div>
