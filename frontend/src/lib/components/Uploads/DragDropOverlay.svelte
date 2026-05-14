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

  // Shared across one drop's recursive walk so each warning fires at most once.
  interface WalkState {
    depthWarned: boolean;
    readWarned: boolean;
  }

  const MAX_DEPTH = 32;

  async function walkEntry(
    entry: FileSystemEntry,
    basePath: string,
    out: CollectedFile[],
    depth: number,
    state: WalkState,
  ): Promise<void> {
    if (depth > MAX_DEPTH) {
      if (!state.depthWarned) {
        state.depthWarned = true;
        notifications.warning(`Folder depth exceeded (${MAX_DEPTH}) — partial upload`);
      }
      return;
    }
    if (entry.name.startsWith(".")) {
      // Skip dotfiles/dotfolders (.DS_Store, .git, __MACOSX is non-dot but harmless).
      return;
    }
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
        let batch: FileSystemEntry[];
        try {
          batch = await new Promise<FileSystemEntry[]>((resolve, reject) =>
            reader.readEntries(resolve, reject),
          );
        } catch (_cause) {
          if (!state.readWarned) {
            state.readWarned = true;
            notifications.warning(
              "Failed to read part of dropped folder — uploads may be incomplete",
            );
          }
          break;
        }
        if (batch.length === 0) break;
        for (const child of batch) {
          await walkEntry(child, subBase, out, depth + 1, state);
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
    const state: WalkState = { depthWarned: false, readWarned: false };
    let entriesOffered = 0;

    const items = event.dataTransfer?.items;
    if (items && items.length > 0 && typeof items[0].webkitGetAsEntry === "function") {
      const walks: Promise<void>[] = [];
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        if (item.kind !== "file") continue;
        const entry = item.webkitGetAsEntry();
        if (entry) {
          entriesOffered++;
          walks.push(walkEntry(entry, here, collected, 0, state));
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
          entriesOffered++;
          collected.push({ file, loc: here });
        }
      }
    }

    for (const { file, loc } of collected) {
      uploads.enqueue({ file, loc, filename: file.name });
    }

    if (collected.length === 0 && entriesOffered > 0) {
      notifications.warning("No uploadable files in dropped folder");
    } else if (collected.length > 0) {
      notifications.info(`Queued ${collected.length} file(s) for upload`);
    }

    pumpQueue();
  }

  onMount(() => {
    window.addEventListener("dragenter", onDragEnter);
    window.addEventListener("dragover", onDragOver);
    window.addEventListener("dragleave", onDragLeave);
    window.addEventListener("drop", onDrop);

    // test-only DEV harness; treeshaken from prod build by import.meta.env.DEV check
    if (import.meta.env.DEV) {
      (window as unknown as Record<string, unknown>).__nasTestEnqueueFolder = (
        items: Array<{ webkitRelativePath: string; name?: string; content?: string }>,
      ) => {
        const here = currentPath ?? "/";
        const collected: CollectedFile[] = [];
        for (const item of items) {
          const rel = item.webkitRelativePath;
          // Apply dotfile guard same as enqueueFolder.
          if (rel.split("/").some((seg) => seg.startsWith("."))) continue;
          const parts = rel.split("/");
          const filename = item.name ?? parts[parts.length - 1];
          const dirParts = parts.slice(0, -1);
          const loc =
            dirParts.length > 0
              ? here === "/"
                ? "/" + dirParts.join("/")
                : here + "/" + dirParts.join("/")
              : here;
          const file = new File([item.content ?? ""], filename);
          collected.push({ file, loc });
        }
        for (const { file, loc } of collected) {
          uploads.enqueue({ file, loc, filename: file.name });
        }
        // Harness mirrors production: any empty-result drop warns (items=[]
        // simulates empty-folder drop; non-empty items all filtered also warns).
        if (collected.length === 0) {
          notifications.warning("No uploadable files in dropped folder");
        } else {
          notifications.info(`Queued ${collected.length} file(s) for upload`);
        }
        pumpQueue();
      };
    }
  });

  onDestroy(() => {
    window.removeEventListener("dragenter", onDragEnter);
    window.removeEventListener("dragover", onDragOver);
    window.removeEventListener("dragleave", onDragLeave);
    window.removeEventListener("drop", onDrop);

    if (import.meta.env.DEV) {
      delete (window as unknown as Record<string, unknown>).__nasTestEnqueueFolder;
    }
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
