import type { Action } from "svelte/action";
import type { FolderEntry } from "./icon-for";
import { hasPayload, NAS_ENTRY_MIME } from "./drag-drop";

export interface DragEntryParams {
  entry: FolderEntry;
  dragPayload: (entry: FolderEntry) => string;
  onDropOnFolder: (event: DragEvent, target: FolderEntry) => void;
  onOpen: (entry: FolderEntry, opts?: { newTab?: boolean }) => void;
  onDropEnter: (name: string) => void;
  onDropLeave: (name: string) => void;
  onDropFinish: () => void;
}

export const dragEntry: Action<HTMLElement, DragEntryParams> = (node, initial) => {
  let params = initial;

  function onDragStart(event: DragEvent) {
    if (!event.dataTransfer) return;
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData(NAS_ENTRY_MIME, params.dragPayload(params.entry));
  }

  function onDragOver(event: DragEvent) {
    if (!params.entry.isFolder || !hasPayload(event)) return;
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = "move";
    params.onDropEnter(params.entry.name);
  }

  function onDragLeave() {
    params.onDropLeave(params.entry.name);
  }

  function onDrop(event: DragEvent) {
    params.onDropFinish();
    params.onDropOnFolder(event, params.entry);
  }

  function onDblClick(event: MouseEvent) {
    params.onOpen(params.entry, {
      newTab: params.entry.isFolder && (event.ctrlKey || event.metaKey),
    });
  }

  function onAuxClick(event: MouseEvent) {
    if (event.button !== 1) return;
    if (!params.entry.isFolder) return;
    event.preventDefault();
    params.onOpen(params.entry, { newTab: true });
  }

  node.addEventListener("dragstart", onDragStart);
  node.addEventListener("dragover", onDragOver);
  node.addEventListener("dragleave", onDragLeave);
  node.addEventListener("drop", onDrop);
  node.addEventListener("dblclick", onDblClick);
  node.addEventListener("auxclick", onAuxClick);

  return {
    update(next: DragEntryParams) {
      params = next;
    },
    destroy() {
      node.removeEventListener("dragstart", onDragStart);
      node.removeEventListener("dragover", onDragOver);
      node.removeEventListener("dragleave", onDragLeave);
      node.removeEventListener("drop", onDrop);
      node.removeEventListener("dblclick", onDblClick);
      node.removeEventListener("auxclick", onAuxClick);
    },
  };
};
