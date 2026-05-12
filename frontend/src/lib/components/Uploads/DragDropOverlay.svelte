<script lang="ts">
  import Upload from "lucide-svelte/icons/upload";
  import { onMount, onDestroy } from "svelte";
  import { uploads } from "$lib/store/uploads.svelte";
  import { pumpUploadQueue } from "$lib/upload-worker";
  import { files } from "$lib/store/files.svelte";

  let visible = $state(false);
  let dragDepth = 0;

  function hasFiles(event: DragEvent): boolean {
    return event.dataTransfer?.types.includes("Files") ?? false;
  }

  function onDragEnter(event: DragEvent) {
    if (!hasFiles(event)) return;
    event.preventDefault();
    dragDepth++;
    visible = true;
  }

  function onDragOver(event: DragEvent) {
    if (!hasFiles(event)) return;
    event.preventDefault();
  }

  function onDragLeave(event: DragEvent) {
    if (!hasFiles(event)) return;
    event.preventDefault();
    dragDepth = Math.max(0, dragDepth - 1);
    if (dragDepth === 0) visible = false;
  }

  function onDrop(event: DragEvent) {
    if (!hasFiles(event)) return;
    event.preventDefault();
    visible = false;
    dragDepth = 0;
    const droppedFiles = event.dataTransfer?.files;
    if (!droppedFiles || droppedFiles.length === 0) return;
    const loc = files.pathDisplay;
    for (const droppedFile of droppedFiles) {
      uploads.enqueue({
        file: droppedFile,
        loc,
        filename: droppedFile.name,
      });
    }
    pumpUploadQueue();
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

{#if visible}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-bg-base/80 backdrop-blur-sm pointer-events-none"
  >
    <div
      class="flex flex-col items-center gap-4 p-12 rounded-lg border-2 border-dashed border-accent bg-bg-elevated"
    >
      <Upload size="48" class="text-accent" />
      <div class="text-lg text-fg-primary font-semibold">Drop files to upload</div>
      <div class="text-sm text-fg-muted font-mono">{files.pathDisplay}</div>
    </div>
  </div>
{/if}
