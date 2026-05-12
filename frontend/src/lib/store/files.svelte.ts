import type { FileEntry, SortDir, SortKey, ViewMode } from "$lib/types";

const VIEW_MODE_STORAGE_KEY = "nas:files:viewMode";

class FilesStore {
  currentLoc = $state<string[]>([]);
  fileList = $state<FileEntry[]>([]);
  selection = $state<Set<string>>(new Set());
  viewMode = $state<ViewMode>("grid");
  sortBy = $state<SortKey>("name");
  sortDir = $state<SortDir>("asc");
  loading = $state<boolean>(false);
  errorMessage = $state<string | null>(null);

  pathDisplay = $derived<string>(
    this.currentLoc.length === 0 ? "/" : "/" + this.currentLoc.join("/"),
  );

  selectedCount = $derived<number>(this.selection.size);

  constructor() {
    if (typeof localStorage !== "undefined") {
      const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
      if (stored === "grid" || stored === "list") {
        this.viewMode = stored;
      }
    }
  }

  setLoc(path: string[]) {
    this.currentLoc = path;
    this.selection = new Set();
  }

  navigateInto(name: string) {
    this.setLoc([...this.currentLoc, name]);
  }

  navigateUp() {
    if (this.currentLoc.length === 0) return;
    this.setLoc(this.currentLoc.slice(0, -1));
  }

  setFileList(list: FileEntry[]) {
    this.fileList = list;
  }

  setSelection(names: Iterable<string>) {
    this.selection = new Set(names);
  }

  toggleSelection(name: string) {
    const next = new Set(this.selection);
    if (next.has(name)) next.delete(name);
    else next.add(name);
    this.selection = next;
  }

  clearSelection() {
    this.selection = new Set();
  }

  setViewMode(mode: ViewMode) {
    this.viewMode = mode;
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(VIEW_MODE_STORAGE_KEY, mode);
    }
  }

  setSort(key: SortKey, dir?: SortDir) {
    this.sortBy = key;
    if (dir) this.sortDir = dir;
  }

  toggleSortDir() {
    this.sortDir = this.sortDir === "asc" ? "desc" : "asc";
  }

  setLoading(value: boolean) {
    this.loading = value;
  }

  setError(message: string | null) {
    this.errorMessage = message;
  }
}

export const files = new FilesStore();
