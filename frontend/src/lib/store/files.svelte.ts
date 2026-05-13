import type { SortDir, SortKey, ViewMode } from "$lib/types";

const VIEW_MODE_STORAGE_KEY = "nas:files:viewMode";

class FilesStore {
  viewMode = $state<ViewMode>("grid");
  sortBy = $state<SortKey>("name");
  sortDir = $state<SortDir>("asc");
  loading = $state<boolean>(false);
  errorMessage = $state<string | null>(null);

  constructor() {
    if (typeof localStorage !== "undefined") {
      const stored = localStorage.getItem(VIEW_MODE_STORAGE_KEY);
      if (stored === "grid" || stored === "list") {
        this.viewMode = stored;
      }
    }
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
