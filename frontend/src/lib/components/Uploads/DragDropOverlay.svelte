<script lang="ts">
  import Upload from "lucide-svelte/icons/upload";
  import { onMount, onDestroy } from "svelte";
  import { uploads } from "$lib/store/uploads.svelte";
  import { notifications } from "$lib/store/notifications.svelte";
  import { pumpQueue } from "$lib/upload-worker";
  import { tabs } from "$lib/store/tabs.svelte";
  import type { ExplorerPayload } from "$lib/types";

  let visible = $state(false);
  let dragDepth = 0;

  const currentPath = $derived.by(() => {
    const a = tabs.active;
    if (a?.kind !== "explorer") return null;
    const p = a.payload as ExplorerPayload | null;
    const loc = p?.loc ?? [];
    return "/" + loc.join("/");
  });

  function hasFiles(event: DragEvent): boolean {
    return event.dataTransfer?.types.includes("Files") ?? false;
  }

  function onDragEnter(event: DragEvent) {
    if (!hasFiles(event) || currentPath === null) return;
    event.preventDefault();
    dragDepth++;
    visible = true;
  }

  function onDragOver(event: DragEvent) {
    if (!hasFiles(event) || currentPath === null) return;
    event.preventDefault();
  }

  function onDragLeave(event: DragEvent) {
    if (!hasFiles(event)) return;
    event.preventDefault();
    dragDepth = Math.max(0, dragDepth - 1);
    if (dragDepth === 0) visible = false;
  }

  interface CollectedFile {
    file: File;
    loc: string;
  }

  async function walkEntry(
    entry: FileSystemEntry,
    basePath: string,
    out: CollectedFile[],
  ): Promise<void> {
    if (entry.isFile) {
      await new Promise<void>((resolve, reject) => {
        (entry as FileSystemFileEntry).file(
          (file) => {
            out.push({ file, loc: basePath });
            resolve();
          },
          (cause) => reject(cause),
        );
      });
      return;
    }
    if (entry.isDirectory) {
      const dir = entry as FileSystemDirectoryEntry;
      const subBase = basePath === "/" ? "/" + dir.name : basePath + "/" + dir.name;
      const reader = dir.createReader();
      while (true) {
        const batch: FileSystemEntry[] = await new Promise((resolve, reject) =>
          reader.readEntries(resolve, reject),
        );
        if (batch.length === 0) break;
        for (const child of batch) {
          await walkEntry(child, subBase, out);
        }
      }
    }
  }

  async function onDrop(event: DragEvent) {
    if (!hasFiles(event) || currentPath === null) return;
    event.preventDefault();
    visible = false;
    dragDepth = 0;

    const here = currentPath;
    const collected: CollectedFile[] = [];

    const items = event.dataTransfer?.items;
    if (items && items.length > 0 && typeof items[0].webkitGetAsEntry === "function") {
      const walks: Promise<void>[] = [];
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.kind !== "file") continue;
        const entry = item.webkitGetAsEntry();
        if (entry) {
          walks.push(walkEntry(entry, here, collected));
        }
      }
      try {
        await Promise.all(walks);
      } catch (cause) {
        notifications.error(`Folder read failed: ${(cause as Error).message}`);
        return;
      }
    } else {
      const droppedFiles = event.dataTransfer?.files;
      if (droppedFiles) {
        for (const file of droppedFiles) {
          collected.push({ file, loc: here });
        }
      }
    }

    for (const { file, loc } of collected) {
      uploads.enqueue({ file, loc, filename: file.name });
    }
    pumpQueue();
  }

  onMount(() => {
    window.addEventListener("dragenter", onDragEnter);
    window.addEventListener("dragover", onDragOver);
    window.addEventListener("dragleave", onDragLeave);
    window.addEventListener("drop", onDrop);
  });

  onDestroy(() => {
    window.removeEventListener("dragenter", onDragEnter);
    window.removeEventListener("dragover", onDragOver);
    window.removeEventListener("dragleave", onDragLeave);
    window.removeEventListener("drop", onDrop);
  });
</script>

{#if visible && currentPath !== null}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-bg-base/80 backdrop-blur-sm pointer-events-none"
  >
    <div
      class="flex flex-col items-center gap-4 p-12 rounded-lg border-2 border-dashed border-accent bg-bg-elevated"
    >
      <Upload size="48" class="text-accent" />
      <div class="text-lg text-fg-primary font-semibold">Drop files or folders to upload</div>
      <div class="text-sm text-fg-muted font-mono">{currentPath}</div>
    </div>
  </div>
{/if}
